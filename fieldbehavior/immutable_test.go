package fieldbehavior

import (
	"testing"

	pb "go.einride.tech/aip/proto/gen/einride/example/freight/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"gotest.tools/v3/assert"
)

type testcase struct {
	name   string
	orig   proto.Message
	msg    proto.Message
	mask   *fieldmaskpb.FieldMask
	errMsg string
}

func Test_ValidateImmutableFieldsAreIdenticalWithMask_Maps(t *testing.T) {
	for _, testcase := range []testcase{
		{
			// map itself is immutable, but the message it contains are not
			name: "success: changes to mutable fields in messages of immutable map",
			orig: &pb.Shipment{ImmutableLineItemsMap: map[string]*pb.LineItem{"key1": {Title: "title1"}}},
			msg:  &pb.Shipment{ImmutableLineItemsMap: map[string]*pb.LineItem{"key1": {Title: "title2"}}},
		},
		{
			// values in immutable map of primitives can be mutated
			name:   "fail: changes to values of immutable map of primitives",
			orig:   &pb.Shipment{ImmutablePrimitivesMap: map[string]string{"label1": "value2"}},
			msg:    &pb.Shipment{ImmutablePrimitivesMap: map[string]string{"label1": "value1"}},
			errMsg: "change to immutable field immutable_primitives_map.`label1`: value2 doesn't match value1",
		},
		{
			name:   "fail: change to immutable field in mutable map",
			orig:   &pb.Shipment{LineItemsMap: map[string]*pb.LineItem{"key1": {ExternalReferenceId: "id1"}}},
			msg:    &pb.Shipment{LineItemsMap: map[string]*pb.LineItem{"key1": {ExternalReferenceId: "id2"}}},
			errMsg: "change to immutable field \"line_items_map.`key1`.external_reference_id\"",
		},
		{
			name:   "fail: item removed",
			orig:   &pb.Shipment{ImmutablePrimitivesMap: map[string]string{"label1": "value1", "label2": "value2"}},
			msg:    &pb.Shipment{ImmutablePrimitivesMap: map[string]string{"label1": "value1"}},
			errMsg: "change to immutable field \"immutable_primitives_map\": expected map to be of length 2, got 1",
		},
		{
			name:   "fail: item added",
			orig:   &pb.Shipment{ImmutablePrimitivesMap: map[string]string{"label1": "value1"}},
			msg:    &pb.Shipment{ImmutablePrimitivesMap: map[string]string{"label1": "value1", "label2": "value2"}},
			errMsg: "change to immutable field \"immutable_primitives_map\": expected map to be of length 1, got 2",
		},
		{
			name:   "fail: different keys",
			orig:   &pb.Shipment{ImmutablePrimitivesMap: map[string]string{"label1": "value1"}},
			msg:    &pb.Shipment{ImmutablePrimitivesMap: map[string]string{"label2": "value2"}},
			errMsg: "change to immutable field: element immutable_primitives_map.`label1` is missing",
		},
		{
			name:   "fail: nil map gets set",
			orig:   &pb.Shipment{ImmutablePrimitivesMap: nil},
			msg:    &pb.Shipment{ImmutablePrimitivesMap: map[string]string{"label1": "value1"}},
			errMsg: "change to immutable field \"immutable_primitives_map\": expected map to be of length 0, got 1",
		},
		{
			name:   "fail: map gets nil-ed",
			orig:   &pb.Shipment{ImmutablePrimitivesMap: map[string]string{"label1": "value1"}},
			msg:    &pb.Shipment{ImmutablePrimitivesMap: nil},
			errMsg: "change to immutable field \"immutable_primitives_map\": expected map to be of length 1, got 0",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			err := ValidateImmutableFieldsAreIdenticalWithMask(testcase.msg, testcase.orig, testcase.mask)
			if testcase.errMsg == "" {
				assert.NilError(t, err)
			} else {
				assert.Error(t, err, testcase.errMsg)
			}
		})
	}
}

func Test_ValidateImmutableFieldsAreIdenticalWithMask_Lists(t *testing.T) {
	for _, testcase := range []testcase{
		{
			// map itself is immutable, but the message it contains are not
			name: "success: identical lists",
			orig: &pb.Shipment{ImmutableLineItemList: []*pb.LineItem{{Title: "title1"}}},
			msg:  &pb.Shipment{ImmutableLineItemList: []*pb.LineItem{{Title: "title1"}}},
		},
		{
			// map itself is immutable, but the message it contains are not
			name: "success: change to list not in mask",
			orig: &pb.Shipment{ImmutableLineItemList: []*pb.LineItem{{Title: "title1"}}},
			msg:  &pb.Shipment{ImmutableLineItemList: []*pb.LineItem{{Title: "title1"}, {Title: "title2"}}},
			mask: &fieldmaskpb.FieldMask{Paths: []string{"immutable_primitives_map"}},
		},
		{
			// values in immutable map of primitives can be mutated
			name:   "fail: item added",
			orig:   &pb.Shipment{ImmutableLineItemList: []*pb.LineItem{{Title: "title1"}}},
			msg:    &pb.Shipment{ImmutableLineItemList: []*pb.LineItem{{Title: "title1"}, {Title: "title2"}}},
			errMsg: "change to immutable field \"immutable_line_item_list\": expected list to be of length 1, got 2",
		},
		{
			name:   "fail: item removed",
			orig:   &pb.Shipment{ImmutableLineItemList: []*pb.LineItem{{Title: "title1"}, {Title: "title2"}}},
			msg:    &pb.Shipment{ImmutableLineItemList: []*pb.LineItem{{Title: "title1"}}},
			errMsg: "change to immutable field \"immutable_line_item_list\": expected list to be of length 2, got 1",
		},
		{
			name:   "fail: list gets nil-ed",
			orig:   &pb.Shipment{ImmutableLineItemList: []*pb.LineItem{{Title: "title1"}, {Title: "title2"}}},
			msg:    &pb.Shipment{ImmutableLineItemList: nil},
			errMsg: "change to immutable field \"immutable_line_item_list\": expected list to be of length 2, got 0",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			err := ValidateImmutableFieldsAreIdenticalWithMask(testcase.msg, testcase.orig, testcase.mask)
			if testcase.errMsg == "" {
				assert.NilError(t, err)
			} else {
				assert.Error(t, err, testcase.errMsg)
			}
		})
	}
}

func Test_ValidateImmutableFieldsAreIdenticalWithMask_Primitives(t *testing.T) {
	for _, testcase := range []testcase{
		{
			name: "success: no changes",
			orig: &pb.Shipment{ExternalReferenceId: "id"},
			msg:  &pb.Shipment{ExternalReferenceId: "id"},
		},
		{
			name:   "fail: different messages",
			orig:   &pb.LineItem{},
			msg:    &pb.Shipment{},
			errMsg: "can't compare different messages: einride.example.freight.v1.Shipment and einride.example.freight.v1.LineItem",
		},
		{
			msg:    &pb.Shipment{ExternalReferenceId: "id1"},
			orig:   &pb.Shipment{ExternalReferenceId: "id2"},
			errMsg: "change to immutable field \"external_reference_id\"",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			err := ValidateImmutableFieldsAreIdenticalWithMask(testcase.msg, testcase.orig, testcase.mask)
			if testcase.errMsg == "" {
				assert.NilError(t, err)
			} else {
				assert.Error(t, err, testcase.errMsg)
			}
		})
	}
}

func Test_ValidateImmutableFieldsWithMask(t *testing.T) {
	for _, testcase := range []testcase{
		{
			name: "success: immutable field not set",
			msg:  &pb.Shipment{Name: "name"},
		},
		{
			name:   "fail: immutable field set",
			msg:    &pb.Shipment{ExternalReferenceId: "id"},
			errMsg: `change to immutable field "external_reference_id"`,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			err := ValidateImmutableFieldsWithMask(testcase.msg, testcase.mask)
			if testcase.errMsg == "" {
				assert.NilError(t, err)
			} else {
				assert.Error(t, err, testcase.errMsg)
			}
		})
	}
}
