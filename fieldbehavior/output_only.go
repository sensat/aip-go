package fieldbehavior

import (
	"errors"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var (
	ErrMismatchedListLength = errors.New("mismatched list length")
	ErrMissingMapKey        = errors.New("missing map key")
)

// recursively clear output_only fields
func ClearOutputOnlyFields(msg proto.Message) {
	ClearFields(msg, annotations.FieldBehavior_OUTPUT_ONLY)

	reflectMsg := msg.ProtoReflect()
	for i := 0; i < reflectMsg.Descriptor().Fields().Len(); i++ {
		field := reflectMsg.Descriptor().Fields().Get(i)

		// if this isn't a populated message, move on.
		// fieldbehavior.ClearFields handles primitives for us.
		if field.Kind() != protoreflect.MessageKind || !reflectMsg.Has(field) {
			continue
		}

		value := reflectMsg.Get(field)
		switch {
		case field.IsList():
			for i := 0; i < value.List().Len(); i++ {
				ClearOutputOnlyFields(value.List().Get(i).Message().Interface())
			}
		case field.IsMap():
			if field.MapValue().Kind() != protoreflect.MessageKind {
				continue
			}
			value.Map().Range(func(_ protoreflect.MapKey, value protoreflect.Value) bool {
				ClearOutputOnlyFields(value.Message().Interface())
				return true
			})
		default:
			ClearOutputOnlyFields(value.Message().Interface())
		}
	}
}

// recursively copy output_only fields.
// lists are skipped entirely
// output fields in map values are copied if the key exists in both source and destination
func CopyOutputOnlyFields(destination proto.Message, source proto.Message) {
	CopyFields(destination, source, annotations.FieldBehavior_OUTPUT_ONLY)

	src := source.ProtoReflect()
	dst := destination.ProtoReflect()
	for i := 0; i < dst.Descriptor().Fields().Len(); i++ {
		srcField := src.Descriptor().Fields().Get(i)
		dstField := dst.Descriptor().Fields().Get(i)

		// non-messages can't be recursed on, so continue to next field. fieldbehavior.CopyFields has already handled them.
		// if src is empty we don't need to copy anything
		// if dst is not set we can't recursively set its fields.
		if srcField.Kind() != protoreflect.MessageKind || !src.Has(srcField) || !dst.Has(dstField) {
			continue
		}

		srcValue := src.Get(srcField)
		dstValue := dst.Get(dstField)
		switch {
		case srcField.IsList():
			// we can't handle lists because there's no way of ensuring we're dealing with the same elements
			// as we iterate over both lists.
			//
			// we're panicking here to ensure we don't silently skip lists.
			//
			// the current thinking for implementing this functionality for lists is to require all lists
			// of messages with output_only fields or subfields to be either IMMUTABLE or OUTPUT_ONLY themselves.
			// This would guarantee a consistent mapping between fields in both arrays.
			panic("can't copy lists output_only fields") //nolint:forbidigo // we want to warn early that we're not handling lists

		case srcField.IsMap():
			srcMap := srcValue.Map()
			dstMap := dstValue.Map()

			// non-messages don't have fields to recurse on, so continue to next field.
			if srcField.MapValue().Kind() != protoreflect.MessageKind {
				continue
			}

			srcMap.Range(func(key protoreflect.MapKey, _ protoreflect.Value) bool {
				// only copy values if destination also has that entry.
				if dstMap.Has(key) {
					CopyOutputOnlyFields(dstMap.Get(key).Message().Interface(), srcMap.Get(key).Message().Interface())
				}

				return true
			})
		default:
			CopyOutputOnlyFields(dstValue.Message().Interface(), srcValue.Message().Interface())
		}
	}
}
