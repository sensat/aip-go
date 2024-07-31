package fieldbehavior

import (
	"errors"
	"fmt"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

var ErrChangeToImmutableField = errors.New("change to immutable field")
var ErrDifferentMessageTypess = errors.New("can't compare different messages")

func ValidateImmutableFieldsWithMask(m proto.Message, mask *fieldmaskpb.FieldMask) error {
	return ValidateImmutableFieldsAreIdenticalWithMask(m, m.ProtoReflect().New().Interface(), mask)
}

func ValidateImmutableFieldsAreIdentical(msg proto.Message, orig proto.Message) error {
	return ValidateImmutableFieldsAreIdenticalWithMask(msg, orig, &fieldmaskpb.FieldMask{Paths: []string{"*"}})
}

func ValidateImmutableFieldsAreIdenticalWithMask(msg proto.Message, orig proto.Message, mask *fieldmaskpb.FieldMask) error {
	if msg.ProtoReflect().Descriptor().FullName() != orig.ProtoReflect().Descriptor().FullName() {
		return fmt.Errorf("%w: %s and %s", ErrDifferentMessageTypess, msg.ProtoReflect().Descriptor().FullName(), orig.ProtoReflect().Descriptor().FullName())
	}

	return validateImmutableFields(msg.ProtoReflect(), orig.ProtoReflect(), mask, "")
}

func validateImmutableFields(msg protoreflect.Message, orig protoreflect.Message, mask *fieldmaskpb.FieldMask, path string) error {
	for i := 0; i < orig.Descriptor().Fields().Len(); i++ {
		field := orig.Descriptor().Fields().Get(i)
		val := msg.Get(field)
		origVal := orig.Get(field)

		currPath := path
		if len(currPath) > 0 {
			currPath += "."
		}
		currPath += string(field.Name())
		if !hasMask(mask, currPath) {
			continue
		}

		switch {
		case field.IsList():
			if Has(field, annotations.FieldBehavior_IMMUTABLE) {
				if err := validateImmutableList(field, val, origVal, mask, currPath); err != nil {
					return err
				}
			}

			// https://aip.dev/144#update-strategies says update methods should replace the entire list, so we skip validating each element separately.
			continue
		case field.IsMap():
			if Has(field, annotations.FieldBehavior_IMMUTABLE) {
				if err := validateImmutableMap(field, val, origVal, mask, currPath); err != nil {
					return err
				}

				// validateImmutableMap validates each map item individually, so we can continue to the next field
				continue
			}

			// if the types in the map aren't messages, there aren't any submessages or attributes that need validating
			// as you can't set a map's value to IMMUTABLE
			if field.MapValue().Kind() != protoreflect.MessageKind {
				continue
			}

			// if this is a map, verify each message that exists in both maps don't change immutable fields
			var mapErr error
			val.Map().Range(func(key protoreflect.MapKey, val protoreflect.Value) bool {
				currPath += fmt.Sprintf(".`%s`", key)
				if origVal.Map().Has(key) {
					mapErr = validateImmutableFields(val.Message(), origVal.Map().Get(key).Message(), mask, currPath)
				}
				// continue while there are no errors
				return mapErr == nil
			})

			if mapErr != nil {
				return mapErr
			}
		case field.Kind() == protoreflect.MessageKind:
			if err := validateImmutableFields(val.Message(), origVal.Message(), mask, currPath); err != nil {
				return err
			}

		default:
			if Has(field, annotations.FieldBehavior_IMMUTABLE) && !val.Equal(origVal) {
				return fmt.Errorf("%w %q", ErrChangeToImmutableField, currPath)
			}
		}
	}

	return nil
}

// We take an immutable map to mean that the keys must be identical as well as their immutable fields. Other fields are not taken into consideration.
// Take this defintion for example:
//
// map<string, Message> a_map [(google.api.field_behavior) = IMMUTABLE];
//
//	message Message {
//		string mutable_field = 1  [(google.api.field_behavior) = OPTIONAL];
//		string immutable_field = 2  [(google.api.field_behavior) = IMMUTABLE];
//	}
//
// The following maps would pass validation
// {} and  {}
// {"key": {"mutable_field": "value", "immutable_field": "value"}} and  {"key": {"mutable_field": "anothervalue", "immutable_field": "value"}}
//
// The following maps would fail validation
// {"key": {"mutable_field": "value", "immutable_field": "value"}} and  {"key": {"mutable_field": "value", "immutable_field": ""}}
// {"key": {"mutable_field": "value", "immutable_field": "value"}} and  {}
// {"key": {"mutable_field": "value", "immutable_field": "value"}} and  {"key": {"mutable_field": "value", "immutable_field": "value"}, "key2": {"mutable_field": "value", "immutable_field": "value"}}
// {} and  {"key": {"mutable_field": "value", "immutable_field": "value"}}
//
// values of maps of primitives are considered immutable, so the following maps would pass validation
// {"key": "value"} and {"key": "value"}
//
// The following maps would fail validation
// {"key": "value"} and {"key": "anothervalue"}
func validateImmutableMap(field protoreflect.FieldDescriptor, val protoreflect.Value, origVal protoreflect.Value, mask *fieldmaskpb.FieldMask, path string) error {
	// all elements must exist in both maps and their immutable fields must be identical
	origMap := origVal.Map()
	newMap := val.Map()
	if origMap.Len() != newMap.Len() {
		return fmt.Errorf("%w %q: expected map to be of length %d, got %d", ErrChangeToImmutableField, path, origMap.Len(), newMap.Len())
	}

	var err error
	// because we know at this point both maps have the same number of entries,
	// we just need to check if the keys in the original map exist in the new map
	origMap.Range(func(key protoreflect.MapKey, v protoreflect.Value) bool {
		currPath := fmt.Sprintf("%s.`%s`", path, key)

		if !newMap.Has(key) {
			err = fmt.Errorf("%w: element %s is missing", ErrChangeToImmutableField, currPath)
			return false
		}

		if field.MapValue().Kind() != protoreflect.MessageKind {
			origMapVal := origMap.Get(key)
			newMapVal := newMap.Get(key)
			if !origMapVal.Equal(newMapVal) {
				err = fmt.Errorf("%w %s: %v doesn't match %v", ErrChangeToImmutableField, currPath, origMapVal, newMapVal)
			}
		} else {
			err = validateImmutableFields(newMap.Get(key).Message(), origMap.Get(key).Message(), mask, currPath)
		}

		// continue while there are no errors
		return err == nil
	})

	return err
}

func validateImmutableList(field protoreflect.FieldDescriptor, val protoreflect.Value, origVal protoreflect.Value, mask *fieldmaskpb.FieldMask, path string) error {
	// both lists must be the same length and their element's immutable fields must be identical
	list := val.List()
	origList := origVal.List()

	if list.Len() != origList.Len() {
		return fmt.Errorf("%w %q: expected list to be of length %d, got %d", ErrChangeToImmutableField, path, origList.Len(), list.Len())
	}

	for i := 0; i < list.Len(); i++ {
		currPath := fmt.Sprintf("%s.%d", path, i)
		if field.Kind() != protoreflect.MessageKind {
			if err := validateImmutableFields(list.Get(i).Message(), origList.Get(i).Message(), mask, currPath); err != nil {
				return err
			}
		} else {
			if !list.Get(i).Equal(origList.Get(i)) {
				return fmt.Errorf("%w %q: %v doesn't match %v", ErrChangeToImmutableField, currPath, origList.Get(i), list.Get(i))
			}
		}
	}

	return nil
}
