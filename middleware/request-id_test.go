package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common/helper"
)

type seenIds struct {
	oneapi, client, ctxClient string
	hasClient                 bool
}

func runRequestId(t *testing.T, header string) (seenIds, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	var seen seenIds
	r := gin.New()
	r.Use(RequestId())
	r.GET("/x", func(c *gin.Context) {
		seen.oneapi = c.GetString(helper.RequestIdKey)
		_, seen.hasClient = c.Get(helper.ClientRequestIdKey)
		seen.client = c.GetString(helper.ClientRequestIdKey)
		seen.ctxClient = helper.GetClientRequestId(c.Request.Context())
		c.Status(http.StatusNoContent)
	})
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	if header != "" {
		req.Header.Set(helper.ClientRequestIdHeader, header)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return seen, w
}

func TestRequestIdKeepsOwnIdAndRecordsValidClientId(t *testing.T) {
	const clientId = "3f1c2d4e-5a6b-4c7d-8e9f-0a1b2c3d4e5f"
	seen, w := runRequestId(t, clientId)
	if seen.oneapi == "" || seen.oneapi == clientId {
		t.Fatalf("One-API 自己的 id 应保留且不等于调用方的 id，got %q", seen.oneapi)
	}
	if got := w.Header().Get(helper.RequestIdKey); got != seen.oneapi {
		t.Fatalf("响应头 %s = %q, want %q", helper.RequestIdKey, got, seen.oneapi)
	}
	if seen.client != clientId || seen.ctxClient != clientId {
		t.Fatalf("client id: gin=%q ctx=%q, want %q", seen.client, seen.ctxClient, clientId)
	}
}

func TestRequestIdIgnoresMissingOrInvalidClientId(t *testing.T) {
	for _, header := range []string{"", "bad id with spaces", strings.Repeat("a", helper.MaxClientRequestIdLength+1), "a\"b"} {
		seen, _ := runRequestId(t, header)
		if seen.hasClient || seen.ctxClient != "" {
			t.Errorf("header %q: client id should be absent, got gin=%q ctx=%q", header, seen.client, seen.ctxClient)
		}
		if seen.oneapi == "" {
			t.Errorf("header %q: One-API id missing", header)
		}
	}
}

func TestFormatAccessLogAppendsClientRequestIdOnlyWhenPresent(t *testing.T) {
	base := gin.LogFormatterParams{
		TimeStamp:  time.Date(2026, 11, 7, 12, 0, 0, 0, time.UTC),
		StatusCode: 200,
		Latency:    1500 * time.Millisecond,
		ClientIP:   "10.42.0.7",
		Method:     http.MethodPost,
		Path:       "/v1/chat/completions",
	}
	with := base
	with.Keys = map[string]any{helper.RequestIdKey: "2026110712000012345678", helper.ClientRequestIdKey: "abc-1"}
	line := formatAccessLog(with)
	if !strings.Contains(line, "| 2026110712000012345678 |") || !strings.HasSuffix(line, "/v1/chat/completions | client_request_id=abc-1\n") {
		t.Fatalf("unexpected line: %q", line)
	}
	without := base
	without.Keys = map[string]any{helper.RequestIdKey: "2026110712000012345678"}
	if line := formatAccessLog(without); strings.Contains(line, "client_request_id") {
		t.Fatalf("line should not mention client_request_id: %q", line)
	}
	// Keys 为空（例如 404 在中间件之前结束）时不 panic
	_ = formatAccessLog(base)
}
