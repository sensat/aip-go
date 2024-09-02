package fieldbehavior

import (
	"testing"
	"time"

	pb "go.einride.tech/aip/proto/gen/einride/example/freight/v1"
	pbsyntax "go.einride.tech/aip/proto/gen/einride/example/syntax/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gotest.tools/v3/assert"
)

func Test_ClearOutputOnlyFields(t *testing.T) {
	t.Run("only clears output_only fields", func(t *testing.T) {
		t.Parallel()
		msg := &pb.Shipment{Name: "name", CreateTime: timestamppb.New(time.Now())}
		ClearOutputOnlyFields(msg)
		assert.Equal(t, msg.CreateTime.GetSeconds(), int64(0))
		assert.Equal(t, msg.Name, "name")
	})

	t.Run("clears deep output_only fields", func(t *testing.T) {
		t.Parallel()
		msg := &pb.UpdateShipmentRequest{Shipment: &pb.Shipment{CreateTime: timestamppb.New(time.Now())}}
		ClearOutputOnlyFields(msg)
		assert.Equal(t, msg.Shipment.CreateTime.GetSeconds(), int64(0))
	})
}

func Test_CopyOutputOnlyFields(t *testing.T) {
	t.Run("only copies output_only fields", func(t *testing.T) {
		t.Parallel()
		msg := &pb.Shipment{Name: "name1", CreateTime: timestamppb.New(time.Now())}
		dst := &pb.Shipment{Name: "name2"}
		assert.Equal(t, CopyOutputOnlyFields(dst, msg), nil)
		assert.Equal(t, msg.CreateTime.GetSeconds(), dst.CreateTime.GetSeconds())
		assert.Equal(t, dst.Name, "name2")
	})

	t.Run("copies deep output_only fields", func(t *testing.T) {
		t.Parallel()
		msg := &pb.UpdateShipmentRequest{Shipment: &pb.Shipment{CreateTime: timestamppb.New(time.Now())}}
		dst := &pb.UpdateShipmentRequest{Shipment: &pb.Shipment{}}
		assert.Equal(t, CopyOutputOnlyFields(dst, msg), nil)
		assert.Equal(t, msg.Shipment.CreateTime.GetSeconds(), dst.Shipment.CreateTime.GetSeconds())
	})

	t.Run("fails for repeated messages without output_only", func(t *testing.T) {
		t.Parallel()
		msg := &pbsyntax.FieldBehaviorMessage{RepeatedMessage: []*pbsyntax.FieldBehaviorMessage{{OutputOnlyField: "field"}}}
		dst := &pbsyntax.FieldBehaviorMessage{RepeatedMessage: []*pbsyntax.FieldBehaviorMessage{{OutputOnlyField: "field"}}}

		assert.Equal(t, CopyOutputOnlyFields(dst, msg).Error(), "can't copy lists output_only fields")
	})
}
