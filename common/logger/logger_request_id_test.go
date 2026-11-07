package logger

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common/helper"
)

func captureInfo(t *testing.T, ctx context.Context) string {
	t.Helper()
	var buf bytes.Buffer
	old := gin.DefaultWriter
	gin.DefaultWriter = &buf
	t.Cleanup(func() { gin.DefaultWriter = old })
	Info(ctx, "hello")
	return buf.String()
}

func TestLogLineCarriesBothRequestIds(t *testing.T) {
	ctx := helper.SetClientRequestId(helper.SetRequestID(context.Background(), "2026110712000012345678"), "abc-1")
	line := captureInfo(t, ctx)
	if !strings.Contains(line, "| 2026110712000012345678 | client_request_id=abc-1 |") {
		t.Fatalf("unexpected line: %q", line)
	}
}

func TestLogLineWithoutClientRequestId(t *testing.T) {
	line := captureInfo(t, helper.SetRequestID(context.Background(), "2026110712000012345678"))
	if strings.Contains(line, "client_request_id") || !strings.Contains(line, "| 2026110712000012345678 |") {
		t.Fatalf("unexpected line: %q", line)
	}
}
