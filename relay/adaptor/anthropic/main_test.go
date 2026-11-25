package anthropic

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/relay/model"
)

func newTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	return c, w
}

func newUpstreamResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestHandler_UsageWithCacheTokens(t *testing.T) {
	c, w := newTestContext()
	body := `{"id":"msg_1","type":"message","role":"assistant","model":"claude-x",
		"content":[{"type":"text","text":"ok"}],"stop_reason":"end_turn",
		"usage":{"input_tokens":10,"output_tokens":5,"cache_read_input_tokens":100,"cache_creation_input_tokens":20}}`
	errResp, usage := Handler(c, newUpstreamResponse(body), 0, "claude-x")
	if errResp != nil {
		t.Fatalf("unexpected error: %+v", errResp)
	}
	assertCacheUsage(t, usage, true)

	var out map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	u := out["usage"].(map[string]any)
	if u["prompt_tokens"] != float64(130) || u["completion_tokens"] != float64(5) || u["total_tokens"] != float64(135) {
		t.Fatalf("unexpected usage json: %v", u)
	}
	if u["cache_read_input_tokens"] != float64(100) || u["cache_creation_input_tokens"] != float64(20) {
		t.Fatalf("missing cache usage fields: %v", u)
	}
	details, ok := u["prompt_tokens_details"].(map[string]any)
	if !ok || details["cached_tokens"] != float64(100) {
		t.Fatalf("missing prompt_tokens_details.cached_tokens: %v", u)
	}
}

func TestHandler_UsageWithoutCacheTokensUnchanged(t *testing.T) {
	c, w := newTestContext()
	body := `{"id":"msg_1","type":"message","role":"assistant","model":"claude-x",
		"content":[{"type":"text","text":"ok"}],"stop_reason":"end_turn",
		"usage":{"input_tokens":10,"output_tokens":5}}`
	_, usage := Handler(c, newUpstreamResponse(body), 0, "claude-x")
	want := model.Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15}
	if usage == nil || *usage != want {
		t.Fatalf("usage = %+v, want %+v", usage, want)
	}
	var out map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	u := out["usage"].(map[string]any)
	for _, k := range []string{"prompt_tokens_details", "cache_read_input_tokens", "cache_creation_input_tokens"} {
		if _, has := u[k]; has {
			t.Fatalf("usage should not contain %s without cache tokens: %v", k, u)
		}
	}
}

// assertCacheUsage checks a usage built from input=10, output=5, cache_read=100, cache_creation=20.
// The stream handler leaves TotalTokens to the caller, as before.
func assertCacheUsage(t *testing.T, usage *model.Usage, checkTotal bool) {
	t.Helper()
	if usage == nil {
		t.Fatal("usage is nil")
	}
	if usage.PromptTokens != 130 || usage.CompletionTokens != 5 || (checkTotal && usage.TotalTokens != 135) {
		t.Fatalf("unexpected token counts: %+v", usage)
	}
	if usage.CacheReadInputTokens != 100 || usage.CacheCreationInputTokens != 20 {
		t.Fatalf("unexpected cache tokens: %+v", usage)
	}
	if usage.PromptTokensDetails == nil || usage.PromptTokensDetails.CachedTokens != 100 {
		t.Fatalf("unexpected prompt_tokens_details: %+v", usage.PromptTokensDetails)
	}
}

func TestStreamHandler_UsageWithCacheTokens(t *testing.T) {
	c, w := newTestContext()
	events := []string{
		`{"type":"message_start","message":{"id":"msg_1","type":"message","role":"assistant","model":"claude-x","content":[],"usage":{"input_tokens":10,"output_tokens":1,"cache_read_input_tokens":100,"cache_creation_input_tokens":20}}}`,
		`{"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`,
		`{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"ok"}}`,
		`{"type":"content_block_stop","index":0}`,
		`{"type":"message_delta","delta":{"stop_reason":"end_turn","stop_sequence":null},"usage":{"output_tokens":4}}`,
		`{"type":"message_stop"}`,
	}
	var sb strings.Builder
	for _, e := range events {
		sb.WriteString("event: x\ndata: " + e + "\n\n")
	}
	errResp, usage := StreamHandler(c, newUpstreamResponse(sb.String()))
	if errResp != nil {
		t.Fatalf("unexpected error: %+v", errResp)
	}
	// Completion tokens keep the existing summing behaviour (1 + 4).
	assertCacheUsage(t, usage, false)
	if !strings.Contains(w.Body.String(), `"content":"ok"`) {
		t.Fatalf("stream output missing text: %s", w.Body.String())
	}
}

func TestStreamHandler_CacheTokensNotDoubleCounted(t *testing.T) {
	c, _ := newTestContext()
	// Newer API versions repeat cumulative cache counts in message_delta.
	body := "data: " + `{"type":"message_start","message":{"id":"msg_1","model":"claude-x","usage":{"input_tokens":10,"output_tokens":1,"cache_read_input_tokens":100,"cache_creation_input_tokens":20}}}` + "\n\n" +
		"data: " + `{"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":4,"cache_read_input_tokens":100,"cache_creation_input_tokens":20}}` + "\n\n"
	_, usage := StreamHandler(c, newUpstreamResponse(body))
	assertCacheUsage(t, usage, false)
}

func TestStreamHandler_UsageWithoutCacheTokensUnchanged(t *testing.T) {
	c, _ := newTestContext()
	body := "data: " + `{"type":"message_start","message":{"id":"msg_1","model":"claude-x","usage":{"input_tokens":10,"output_tokens":1}}}` + "\n\n" +
		"data: " + `{"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":4}}` + "\n\n"
	_, usage := StreamHandler(c, newUpstreamResponse(body))
	want := model.Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 0}
	if usage == nil || *usage != want {
		t.Fatalf("usage = %+v, want %+v", usage, want)
	}
}
