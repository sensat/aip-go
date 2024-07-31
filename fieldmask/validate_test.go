package fieldmask

import (
	"testing"

	syntaxv1 "go.einride.tech/aip/proto/gen/einride/example/syntax/v1"
	"google.golang.org/genproto/googleapis/example/library/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"gotest.tools/v3/assert"
)

func TestValidate(t *testing.T) {
	// t.Parallel()
	for _, tt := range []struct {
		name          string
		fieldMask     *fieldmaskpb.FieldMask
		message       proto.Message
		errorContains string
	}{
		{
			name:    "valid nil",
			message: &library.Book{},
		},
		{
			name:      "valid *",
			fieldMask: &fieldmaskpb.FieldMask{Paths: []string{"*"}},
			message:   &library.Book{},
		},
		{
			name:          "invalid *",
			fieldMask:     &fieldmaskpb.FieldMask{Paths: []string{"*", "author"}},
			message:       &library.Book{},
			errorContains: "invalid field path: '*' must not be used with other paths",
		},
		{
			name:      "valid empty",
			fieldMask: &fieldmaskpb.FieldMask{},
			message:   &library.Book{},
		},

		{
			name: "valid single",
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{"name", "author"},
			},
			message: &library.Book{},
		},

		{
			name: "invalid single",
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{"name", "foo"},
			},
			message:       &library.Book{},
			errorContains: "invalid field path: foo",
		},

		{
			name: "valid nested",
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{"parent", "book.name"},
			},
			message: &library.CreateBookRequest{},
		},

		{
			name: "invalid nested",
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{"parent", "book.foo"},
			},
			message:       &library.CreateBookRequest{},
			errorContains: "invalid field path: book.foo",
		},

		{
			name: "valid key in message map",
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{"map_string_message.key.string"},
			},
			message: &syntaxv1.Message{MapStringMessage: map[string]*syntaxv1.Message{"key": {
				String_: "value",
			}}},
		},

		{
			name: "valid key in primitive map",
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{"map_string_string.key"},
			},
			message: &syntaxv1.Message{MapStringString: map[string]string{"key": "value"}},
		},

		{
			name: "valid backtick escaped key in primitive map",
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{"map_string_string.`path.with.dot`"},
			},
			message: &syntaxv1.Message{MapStringString: map[string]string{"path.with.dot": "value"}},
		},

		{
			name: "invalid key in primitive map",
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{"map_string_string.nokey"},
			},
			message:       &syntaxv1.Message{MapStringString: map[string]string{"key": "value"}},
			errorContains: "invalid field path: map_string_string.nokey",
		},

		{
			name: "invalid backtick escaped key in primitive map",
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{"map_string_string.`no.path.with.dot`"},
			},
			message:       &syntaxv1.Message{MapStringString: map[string]string{"path.with.dot": "value"}},
			errorContains: "invalid field path: map_string_string.`no.path.with.dot`",
		},

		{
			name: "valid backtick escaped key in map",
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{"map_string_message.`path.with.dot`.string"},
			},
			message: &syntaxv1.Message{MapStringMessage: map[string]*syntaxv1.Message{"path.with.dot": {
				String_: "value",
			}}},
		},

		{
			name: "missing backtick escaped key in map",
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{"map_string_message.`no.path.with.dot`.string"},
			},
			message: &syntaxv1.Message{MapStringMessage: map[string]*syntaxv1.Message{"path.with.dot": {
				String_: "value",
			}}},
			errorContains: "invalid field path: map_string_message.`no.path.with.dot`.string",
		},

		{
			name: "invalid primitive subfield",
			fieldMask: &fieldmaskpb.FieldMask{
				Paths: []string{"string.string"},
			},
			message:       &syntaxv1.Message{String_: "value"},
			errorContains: "invalid field path: string.string",
		},
	} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// t.Parallel()
			if tt.errorContains != "" {
				assert.ErrorContains(t, Validate(tt.fieldMask, tt.message), tt.errorContains)
			} else {
				assert.NilError(t, Validate(tt.fieldMask, tt.message))
			}
		})
	}
}
