package adaptor

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/relay/meta"
)

func upstreamRequest(t *testing.T, forward bool, oneapiId, clientId string) *http.Request {
	t.Helper()
	old := config.RelayForwardRequestId
	config.RelayForwardRequestId = forward
	t.Cleanup(func() { config.RelayForwardRequestId = old })
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	c.Request.Header.Set("Content-Type", "application/json")
	if oneapiId != "" {
		c.Set(helper.RequestIdKey, oneapiId)
	}
	if clientId != "" {
		c.Set(helper.ClientRequestIdKey, clientId)
	}
	req := httptest.NewRequest(http.MethodPost, "http://upstream/v1/chat/completions", nil)
	SetupCommonRequestHeader(c, req, &meta.Meta{})
	return req
}

func spendMetadata(t *testing.T, req *http.Request) map[string]string {
	t.Helper()
	var m map[string]string
	if err := json.Unmarshal([]byte(req.Header.Get(helper.SpendLogsMetadataHeader)), &m); err != nil {
		t.Fatalf("spend logs metadata is not JSON: %v", err)
	}
	return m
}

func TestForwardRequestIdsUsesClientIdWhenPresent(t *testing.T) {
	req := upstreamRequest(t, true, "2026110712000012345678", "abc-1")
	if got := req.Header.Get(helper.ClientRequestIdHeader); got != "abc-1" {
		t.Fatalf("X-Request-Id = %q, want abc-1", got)
	}
	if got := req.Header.Get(helper.RequestIdKey); got != "2026110712000012345678" {
		t.Fatalf("X-Oneapi-Request-Id = %q", got)
	}
	m := spendMetadata(t, req)
	if m["oneapi_request_id"] != "2026110712000012345678" || m["client_request_id"] != "abc-1" || len(m) != 2 {
		t.Fatalf("unexpected spend logs metadata: %v", m)
	}
}

func TestForwardRequestIdsFallsBackToOneapiId(t *testing.T) {
	req := upstreamRequest(t, true, "2026110712000012345678", "")
	if got := req.Header.Get(helper.ClientRequestIdHeader); got != "2026110712000012345678" {
		t.Fatalf("X-Request-Id = %q, want One-API id", got)
	}
	if m := spendMetadata(t, req); len(m) != 1 || m["oneapi_request_id"] == "" {
		t.Fatalf("unexpected spend logs metadata: %v", m)
	}
}

func TestForwardRequestIdsOffByDefault(t *testing.T) {
	req := upstreamRequest(t, false, "2026110712000012345678", "abc-1")
	for _, h := range []string{helper.ClientRequestIdHeader, helper.RequestIdKey, helper.SpendLogsMetadataHeader} {
		if got := req.Header.Get(h); got != "" {
			t.Errorf("%s should not be forwarded when disabled, got %q", h, got)
		}
	}
	if req.Header.Get("Content-Type") != "application/json" {
		t.Fatal("common headers should still be copied")
	}
}
