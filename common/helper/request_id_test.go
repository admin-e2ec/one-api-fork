package helper

import (
	"context"
	"strings"
	"testing"
)

func TestSanitizeClientRequestIdAcceptsUuidAndSafeChars(t *testing.T) {
	for _, id := range []string{
		"3f1c2d4e-5a6b-4c7d-8e9f-0a1b2c3d4e5f",
		"client:run_42.step-1",
		"A",
		strings.Repeat("a", MaxClientRequestIdLength),
	} {
		if got := SanitizeClientRequestId(id); got != id {
			t.Errorf("SanitizeClientRequestId(%q) = %q, want unchanged", id, got)
		}
	}
}

func TestSanitizeClientRequestIdRejectsUnsafeOrOversized(t *testing.T) {
	for _, id := range []string{
		"",
		strings.Repeat("a", MaxClientRequestIdLength+1),
		"abc\ndef",
		"abc def",
		"abc\"def",
		"abc/def",
		"abc|def",
		"中文",
		"abc\x00",
	} {
		if got := SanitizeClientRequestId(id); got != "" {
			t.Errorf("SanitizeClientRequestId(%q) = %q, want empty", id, got)
		}
	}
}

func TestClientRequestIdContextRoundTrip(t *testing.T) {
	if got := GetClientRequestId(context.Background()); got != "" {
		t.Fatalf("empty context returned %q", got)
	}
	//nolint:staticcheck // nil context 也要安全返回空串
	if got := GetClientRequestId(nil); got != "" {
		t.Fatalf("nil context returned %q", got)
	}
	ctx := SetClientRequestId(context.Background(), "abc-1")
	if got := GetClientRequestId(ctx); got != "abc-1" {
		t.Fatalf("got %q, want abc-1", got)
	}
	// 与 One-API 自己的请求 id 互不覆盖
	ctx = SetRequestID(ctx, "2026110712000012345678")
	if GetClientRequestId(ctx) != "abc-1" || GetRequestID(ctx) != "2026110712000012345678" {
		t.Fatalf("ids clobbered: client=%q oneapi=%q", GetClientRequestId(ctx), GetRequestID(ctx))
	}
}
