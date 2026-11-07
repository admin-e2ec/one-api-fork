package model

import (
	"context"
	"testing"

	"github.com/songquanpeng/one-api/common/helper"
)

func TestApplyRequestIdsWritesBothIds(t *testing.T) {
	ctx := helper.SetClientRequestId(helper.SetRequestID(context.Background(), "2026110712000012345678"), "abc-1")
	log := &Log{}
	applyRequestIds(ctx, log)
	if log.RequestId != "2026110712000012345678" || log.ClientRequestId != "abc-1" {
		t.Fatalf("got request_id=%q client_request_id=%q", log.RequestId, log.ClientRequestId)
	}
	plain := &Log{}
	applyRequestIds(helper.SetRequestID(context.Background(), "x1"), plain)
	if plain.RequestId != "x1" || plain.ClientRequestId != "" {
		t.Fatalf("got request_id=%q client_request_id=%q", plain.RequestId, plain.ClientRequestId)
	}
}
