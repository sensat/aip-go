package fieldmask

import (
	"testing"

	syntaxv1 "go.einride.tech/aip/proto/gen/einride/example/syntax/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"gotest.tools/v3/assert"
)

func TestGet_GetsListNestedFields_WithWildcard(t *testing.T) {
	msg := &syntaxv1.FieldBehaviorMessage{
		RepeatedMessage: []*syntaxv1.FieldBehaviorMessage{{
			OptionalMessage: &syntaxv1.FieldBehaviorMessage{
				Field: "repeated_message_1",
				RepeatedMessage: []*syntaxv1.FieldBehaviorMessage{
					{Field: "inner_repeated_message_1_1"},
					{Field: "inner_repeated_message_1_2"},
					{Field: "inner_repeated_message_1_3"},
				},
			}},
			{OptionalMessage: &syntaxv1.FieldBehaviorMessage{
				Field: "repeated_message_2",
				RepeatedMessage: []*syntaxv1.FieldBehaviorMessage{
					{Field: "inner_repeated_message_2_1"},
					{Field: "inner_repeated_message_2_2"},
					{Field: "inner_repeated_message_2_3"},
				},
			}},
			{OptionalMessage: &syntaxv1.FieldBehaviorMessage{
				Field: "repeated_message_3",
				RepeatedMessage: []*syntaxv1.FieldBehaviorMessage{
					{Field: "inner_repeated_message_3_1"},
					{Field: "inner_repeated_message_3_2"},
					{Field: "inner_repeated_message_3_3"},
				},
			}},
		},
	}

	path := "repeated_message.*.optional_message.repeated_message.*.field"
	fieldMask := &fieldmaskpb.FieldMask{
		Paths: []string{path},
	}
	result, err := Get(fieldMask, msg)
	if err != nil {
		t.Fatalf("Get returned an error: %v", err)
	}

	assert.Equal(t, len(result[path]), 9)
	assert.Equal(t, result[path][0].Value.String(), "inner_repeated_message_1_1")
	assert.Equal(t, result[path][1].Value.String(), "inner_repeated_message_1_2")
	assert.Equal(t, result[path][2].Value.String(), "inner_repeated_message_1_3")
	assert.Equal(t, result[path][3].Value.String(), "inner_repeated_message_2_1")
	assert.Equal(t, result[path][4].Value.String(), "inner_repeated_message_2_2")
	assert.Equal(t, result[path][5].Value.String(), "inner_repeated_message_2_3")
	assert.Equal(t, result[path][6].Value.String(), "inner_repeated_message_3_1")
	assert.Equal(t, result[path][7].Value.String(), "inner_repeated_message_3_2")
	assert.Equal(t, result[path][8].Value.String(), "inner_repeated_message_3_3")
}

func TestGet_GetsListNestedFields_WithSpecificIndex(t *testing.T) {
	msg := &syntaxv1.FieldBehaviorMessage{
		RepeatedMessage: []*syntaxv1.FieldBehaviorMessage{{
			OptionalMessage: &syntaxv1.FieldBehaviorMessage{
				Field: "repeated_message_1",
				RepeatedMessage: []*syntaxv1.FieldBehaviorMessage{
					{Field: "inner_repeated_message_1_1"},
					{Field: "inner_repeated_message_1_2"},
					{Field: "inner_repeated_message_1_3"},
				},
			}},
		},
	}

	path := "repeated_message.0.optional_message.repeated_message.2.field"
	fieldMask := &fieldmaskpb.FieldMask{
		Paths: []string{path},
	}
	result, err := Get(fieldMask, msg)
	if err != nil {
		t.Fatalf("Get returned an error: %v", err)
	}

	assert.Equal(t, result[path][0].Value.String(), "inner_repeated_message_1_3")
}

func TestGet_GetsListAllListItems_WithWildcard(t *testing.T) {
	msg := &syntaxv1.FieldBehaviorMessage{
		RepeatedMessage: []*syntaxv1.FieldBehaviorMessage{
			{Field: "repeated_message_1"},
			{Field: "repeated_message_2"},
			{Field: "repeated_message_3"},
		},
	}

	path := "repeated_message.*"
	fieldMask := &fieldmaskpb.FieldMask{
		Paths: []string{path},
	}
	result, err := Get(fieldMask, msg)
	if err != nil {
		t.Fatalf("Get returned an error: %v", err)
	}

	assert.Equal(t, len(result[path]), 3)
}
