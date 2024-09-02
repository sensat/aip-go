package fieldmask

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// Validate validates that the paths in the provided field mask are syntactically valid and
// refer to known fields in the specified message type.
func Validate(fm *fieldmaskpb.FieldMask, m0 proto.Message) error {
	// special case for '*'
	if stringsContain(WildcardPath, fm.GetPaths()) {
		if len(fm.GetPaths()) != 1 {
			return fmt.Errorf("invalid field path: '*' must not be used with other paths")
		}
		return nil
	}
	for _, path := range fm.GetPaths() {
		m := m0.ProtoReflect()
		md := m.Descriptor()
		fields := SplitPath(path)

		for i := 0; i < len(fields); i++ {
			field := fields[i]
			// Search the field within the message.
			if md == nil {
				return fmt.Errorf("field '%s' not in message. invalid field path: %s", field, path) // not within a message
			}
			name := protoreflect.Name(field)
			fd := md.Fields().ByName(name)

			// The real field name of a group is the message name.
			if fd == nil {
				gd := md.Fields().ByName(protoreflect.Name(strings.ToLower(field)))
				if gd != nil && gd.Kind() == protoreflect.GroupKind && string(gd.Message().Name()) == field {
					fd = gd
				}
			}

			if fd == nil {
				return fmt.Errorf("invalid field path: %s", path)
			}

			switch {
			case fd.IsMap():
				// if we're not in the last field, we want to get an entry in the map
				if i+1 < len(fields) {
					// move to the next field so we can get the map entry
					i++
					val := m.Get(fd).Map().Get(protoreflect.ValueOf(fields[i]).MapKey())

					// key doesn't exist in map
					if !val.IsValid() {
						return fmt.Errorf("key doesn't exist in map. invalid field path: %s", path)
					}

					// if this isn't a message (e.g. a primitive) and we're not at the end of the path, then this path is invalid
					// as we can't address fields in primitive types
					if fd.MapValue().Kind() != protoreflect.MessageKind {
						if i+1 != len(fields) {
							return fmt.Errorf("map value was a primitive, which don't have fields. invalid field path: %s", path)
						}

						return nil
					}

					m = val.Message().Interface().ProtoReflect()
					md = m.Descriptor()
				}
			case fd.IsList():
				// lists aren't addressable by item according to the FieldMask spec
				// https://aip.dev/161#wildcards
				// if we're not at the end of the list of fields, then this path is invalid.
				if i+1 != len(fields) {
					return fmt.Errorf("lists aren't addressable by item. invalid field path: %s. ", path)
				}
			case fd.Kind() == protoreflect.MessageKind:
				m = m.Get(fd).Message().Interface().ProtoReflect()
				md = m.Descriptor()
			default:
				// primitives can't have submessages
				// if we're not at the end of the list of fields, then this path is invalid.
				if i+1 != len(fields) {
					return fmt.Errorf("primitives don't have fields. invalid field path: %s.", path)
				}
			}
		}
	}

	return nil
}

func stringsContain(str string, ss []string) bool {
	for _, s := range ss {
		if s == str {
			return true
		}
	}
	return false
}
