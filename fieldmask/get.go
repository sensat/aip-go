package fieldmask

import (
	"errors"
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

var (
	ErrFieldNotFound        = errors.New("field not found")
	ErrListItemNotSupported = errors.New("field masks for list items not supported")
)

type FieldAttributes struct {
	Descriptor protoreflect.FieldDescriptor
	Value      protoreflect.Value
}

// Get retrieves fields in src using a field mask.
//
// Field masks should be validated beforehand.
func Get(mask *fieldmaskpb.FieldMask, src proto.Message) ([]FieldAttributes, error) {
	paths := mask.GetPaths()
	values := make([]FieldAttributes, len(paths))
	srcReflect := src.ProtoReflect()
	for i, path := range paths {
		segments := SplitPath(path)
		field, value, err := getNamedField(srcReflect, segments)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrFieldNotFound, path)
		}

		values[i] = FieldAttributes{Descriptor: field, Value: value}
	}

	return values, nil
}

func getNamedField(src protoreflect.Message, segments []string) (protoreflect.FieldDescriptor, protoreflect.Value, error) {
	if len(segments) == 0 {
		return nil, protoreflect.Value{}, ErrFieldNotFound
	}

	field := src.Descriptor().Fields().ByName(protoreflect.Name(segments[0]))
	if field == nil {
		return nil, protoreflect.Value{}, ErrFieldNotFound
	}

	// a named field in this message
	if len(segments) == 1 {
		return field, src.Get(field), nil
	}

	// a named field in a nested message
	switch {
	case field.IsList():
		// nested fields in repeated not supported
		return nil, protoreflect.Value{}, ErrListItemNotSupported
	case field.IsMap():
		key := protoreflect.ValueOf(segments[1]).MapKey()
		srcMap := src.Get(field).Map()

		if !srcMap.Has(key) {
			return nil, protoreflect.Value{}, fmt.Errorf("%w, map has %s no entry %s", ErrFieldNotFound, segments[0], segments[1])
		}

		// continue iterating into the map entry
		return getNamedField(srcMap.Get(key).Message(), segments[2:])
	case field.Message() != nil:
		return getNamedField(src.Get(field).Message(), segments[1:])
	}

	return nil, protoreflect.Value{}, nil
}
