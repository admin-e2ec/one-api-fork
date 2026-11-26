package anthropic

import (
	"encoding/json"
	"testing"

	"github.com/songquanpeng/one-api/relay/model"
)

func convertJSON(t *testing.T, raw string) (*Request, map[string]any) {
	t.Helper()
	var req model.GeneralOpenAIRequest
	if err := json.Unmarshal([]byte(raw), &req); err != nil {
		t.Fatalf("unmarshal openai request: %v", err)
	}
	claudeReq := ConvertRequest(req)
	b, err := json.Marshal(claudeReq)
	if err != nil {
		t.Fatalf("marshal claude request: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("unmarshal claude request: %v", err)
	}
	return claudeReq, out
}

func marshalString(t *testing.T, raw string) string {
	t.Helper()
	var req model.GeneralOpenAIRequest
	if err := json.Unmarshal([]byte(raw), &req); err != nil {
		t.Fatalf("unmarshal openai request: %v", err)
	}
	b, err := json.Marshal(ConvertRequest(req))
	if err != nil {
		t.Fatalf("marshal claude request: %v", err)
	}
	return string(b)
}

func TestConvertRequest_NoMarkersByteIdentical(t *testing.T) {
	cases := map[string][2]string{
		"string system and mixed messages": {noMarkerRequestJSON, noMarkerGolden},
		"list system":                      {noMarkerListSystemRequestJSON, noMarkerListSystemGolden},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			if got := marshalString(t, c[0]); got != c[1] {
				t.Fatalf("converted request changed\n got: %s\nwant: %s", got, c[1])
			}
		})
	}
}

func block(t *testing.T, out map[string]any, msgIdx, blockIdx int) map[string]any {
	t.Helper()
	msgs := out["messages"].([]any)
	content := msgs[msgIdx].(map[string]any)["content"].([]any)
	return content[blockIdx].(map[string]any)
}

func assertCache(t *testing.T, b map[string]any, wantTTL string) {
	t.Helper()
	cc, ok := b["cache_control"].(map[string]any)
	if !ok {
		t.Fatalf("block has no cache_control: %v", b)
	}
	if cc["type"] != "ephemeral" {
		t.Fatalf("cache_control.type = %v, want ephemeral", cc["type"])
	}
	if wantTTL == "" {
		if _, has := cc["ttl"]; has {
			t.Fatalf("unexpected ttl: %v", cc)
		}
		return
	}
	if cc["ttl"] != wantTTL {
		t.Fatalf("cache_control.ttl = %v, want %s", cc["ttl"], wantTTL)
	}
}

func assertNoCache(t *testing.T, b map[string]any) {
	t.Helper()
	if _, has := b["cache_control"]; has {
		t.Fatalf("block should not carry cache_control: %v", b)
	}
}

func TestConvertRequest_SystemStringWithMessageMarker(t *testing.T) {
	_, out := convertJSON(t, `{"model":"m","messages":[
		{"role":"system","content":"Long system prompt","cache_control":{"type":"ephemeral","ttl":"1h"}},
		{"role":"user","content":"Hi"}]}`)
	system, ok := out["system"].([]any)
	if !ok || len(system) != 1 {
		t.Fatalf("system should be a one-block array, got %#v", out["system"])
	}
	b := system[0].(map[string]any)
	if b["type"] != "text" || b["text"] != "Long system prompt" {
		t.Fatalf("unexpected system block: %v", b)
	}
	assertCache(t, b, "1h")
}

func TestConvertRequest_SystemListWithPartMarker(t *testing.T) {
	_, out := convertJSON(t, `{"model":"m","messages":[
		{"role":"system","content":[
			{"type":"text","text":"Static part"},
			{"type":"text","text":"Cached part","cache_control":{"type":"ephemeral"}},
			{"type":"text","text":"Dynamic part"}]},
		{"role":"user","content":"Hi"}]}`)
	system, ok := out["system"].([]any)
	if !ok || len(system) != 3 {
		t.Fatalf("system should be a three-block array, got %#v", out["system"])
	}
	assertNoCache(t, system[0].(map[string]any))
	assertCache(t, system[1].(map[string]any), "")
	assertNoCache(t, system[2].(map[string]any))
	if system[2].(map[string]any)["text"] != "Dynamic part" {
		t.Fatalf("unexpected system block: %v", system[2])
	}
}

func TestConvertRequest_SystemListWithMessageMarker(t *testing.T) {
	_, out := convertJSON(t, `{"model":"m","messages":[
		{"role":"system","content":[{"type":"text","text":"A"},{"type":"text","text":"B"}],"cache_control":{"type":"ephemeral"}},
		{"role":"user","content":"Hi"}]}`)
	system, ok := out["system"].([]any)
	if !ok || len(system) != 2 {
		t.Fatalf("want 2 system blocks, got %#v", out["system"])
	}
	assertNoCache(t, system[0].(map[string]any))
	assertCache(t, system[1].(map[string]any), "")
}

func TestConvertRequest_UserTextAndImagePartMarkers(t *testing.T) {
	_, out := convertJSON(t, `{"model":"m","messages":[
		{"role":"user","content":[
			{"type":"text","text":"Doc","cache_control":{"type":"ephemeral","ttl":"5m"}},
			{"type":"image_url","image_url":{"url":"data:image/png;base64,iVBORw0KGgo="},"cache_control":{"type":"ephemeral"}},
			{"type":"text","text":"Question"}]}]}`)
	assertCache(t, block(t, out, 0, 0), "5m")
	img := block(t, out, 0, 1)
	if img["type"] != "image" {
		t.Fatalf("want image block, got %v", img)
	}
	assertCache(t, img, "")
	assertNoCache(t, block(t, out, 0, 2))
	if _, isString := out["system"].(string); out["system"] != nil && !isString {
		t.Fatalf("system should stay absent/string without system markers: %#v", out["system"])
	}
}

func TestConvertRequest_UserMessageLevelMarkerGoesOnLastBlock(t *testing.T) {
	_, out := convertJSON(t, `{"model":"m","messages":[
		{"role":"user","content":[{"type":"text","text":"A"},{"type":"text","text":"B"}],"cache_control":{"type":"ephemeral"}},
		{"role":"user","content":"plain","cache_control":{"type":"ephemeral"}}]}`)
	assertNoCache(t, block(t, out, 0, 0))
	assertCache(t, block(t, out, 0, 1), "")
	assertCache(t, block(t, out, 1, 0), "")
}

func TestConvertRequest_ToolMessageMarker(t *testing.T) {
	_, out := convertJSON(t, `{"model":"m","messages":[
		{"role":"user","content":"weather?"},
		{"role":"assistant","content":"","tool_calls":[{"id":"call_1","type":"function","function":{"name":"w","arguments":"{}"}}]},
		{"role":"tool","tool_call_id":"call_1","content":"sunny","cache_control":{"type":"ephemeral"}}]}`)
	b := block(t, out, 2, 0)
	if b["type"] != "tool_result" || b["tool_use_id"] != "call_1" {
		t.Fatalf("want tool_result block, got %v", b)
	}
	assertCache(t, b, "")
}

func TestConvertRequest_AssistantMessageMarker(t *testing.T) {
	_, out := convertJSON(t, `{"model":"m","messages":[
		{"role":"user","content":"hi"},
		{"role":"assistant","content":"Sure","cache_control":{"type":"ephemeral"}},
		{"role":"user","content":"weather?"},
		{"role":"assistant","content":"","tool_calls":[{"id":"call_1","type":"function","function":{"name":"w","arguments":"{}"}}],"cache_control":{"type":"ephemeral"}}]}`)
	assertCache(t, block(t, out, 1, 0), "")
	assertNoCache(t, block(t, out, 3, 0))
	last := block(t, out, 3, 1)
	if last["type"] != "tool_use" {
		t.Fatalf("want tool_use as last block, got %v", last)
	}
	assertCache(t, last, "")
}
