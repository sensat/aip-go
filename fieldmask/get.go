package fieldmask

import (
	"errors"
	"fmt"
	"strconv"

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
func Get(mask *fieldmaskpb.FieldMask, src proto.Message) (map[string][]FieldAttributes, error) {
	paths := mask.GetPaths()
	values := make(map[string][]FieldAttributes, len(paths))
	srcReflect := src.ProtoReflect()
	for _, path := range paths {
		segments := SplitPath(path)
		fieldAttribute, err := getNamedFields(srcReflect, segments)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrFieldNotFound, path)
		}

		values[path] = fieldAttribute
	}

	return values, nil
}

func getNamedFields(src protoreflect.Message, segments []string) ([]FieldAttributes, error) {
	if len(segments) == 0 {
		return nil, ErrFieldNotFound
	}

	field := src.Descriptor().Fields().ByName(protoreflect.Name(segments[0]))
	if field == nil {
		return nil, ErrFieldNotFound
	}

	// a named field in this message
	if len(segments) == 1 {
		return []FieldAttributes{{Descriptor: field, Value: src.Get(field)}}, nil
	}

	// a named field in a nested message
	switch {
	case field.IsList():
		list := src.Get(field).List()
		positions, err := generateListPositions(list.Len(), segments)
		if err != nil {
			return nil, err
		}

		fields := []FieldAttributes{}
		// if we're getting the whole item in the list, append to the list without further recursion
		if len(segments) == 2 {
			for _, pos := range positions {
				fields = append(fields, FieldAttributes{Descriptor: field, Value: list.Get(pos)})
			}
		} else {
			for _, pos := range positions {
				innerFields, err := getNamedFields(list.Get(pos).Message(), segments[2:])
				if err != nil {
					return nil, err
				}

				fields = append(fields, innerFields...)
			}
		}

		return fields, nil
	case field.IsMap():
		key := protoreflect.ValueOf(segments[1]).MapKey()
		srcMap := src.Get(field).Map()

		if !srcMap.Has(key) {
			return nil, fmt.Errorf("%w, map has %s no entry %s", ErrFieldNotFound, segments[0], segments[1])
		}

		// continue iterating into the map entry
		return getNamedFields(srcMap.Get(key).Message(), segments[2:])
	case field.Message() != nil:
		return getNamedFields(src.Get(field).Message(), segments[1:])
	}

	return nil, nil
}

func generateListPositions(listLength int, segments []string) ([]int, error) {
	// if pos is a wildcard, add all positions to the list
	if segments[1] == "*" {
		positions := make([]int, listLength)
		for i := range positions {
			positions[i] = i
		}

		return positions, nil
	}

	// if pos is a valid integer in the range of the list, add it to the positions
	pos, err := strconv.Atoi(segments[1])
	if err != nil {
		return nil, fmt.Errorf("%w: %s is not a valid list position", ErrFieldNotFound, segments[1])
	}

	if pos < 0 || pos >= listLength {
		return nil, fmt.Errorf("%w: %s is not a valid list position", ErrFieldNotFound, segments[1])
	}

	return []int{pos}, nil
}
