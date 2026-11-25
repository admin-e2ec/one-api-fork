package model

import (
	"encoding/json"
	"testing"
)

func TestParseContent_CacheControl(t *testing.T) {
	var m Message
	raw := `{"role":"user","cache_control":{"type":"ephemeral","ttl":"1h"},"content":[
		{"type":"text","text":"a","cache_control":{"type":"ephemeral","ttl":"5m"}},
		{"type":"image_url","image_url":{"url":"data:image/png;base64,AA=="},"cache_control":{"type":"ephemeral"}},
		{"type":"text","text":"b","cache_control":{"ttl":"1h"}},
		{"type":"text","text":"c","cache_control":"bad"}]}`
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		t.Fatal(err)
	}
	if m.CacheControl == nil || m.CacheControl.Type != "ephemeral" || m.CacheControl.TTL != "1h" {
		t.Fatalf("message cache_control = %+v", m.CacheControl)
	}
	parts := m.ParseContent()
	if len(parts) != 4 {
		t.Fatalf("want 4 parts, got %d", len(parts))
	}
	if cc := parts[0].CacheControl; cc == nil || cc.Type != "ephemeral" || cc.TTL != "5m" {
		t.Fatalf("text part cache_control = %+v", cc)
	}
	if cc := parts[1].CacheControl; cc == nil || cc.Type != "ephemeral" || cc.TTL != "" {
		t.Fatalf("image part cache_control = %+v", cc)
	}
	if parts[2].CacheControl != nil || parts[3].CacheControl != nil {
		t.Fatal("invalid cache_control values should be ignored")
	}
}

func TestMessageMarshal_NoCacheControlUnchanged(t *testing.T) {
	m := Message{Role: "user", Content: "hi"}
	b, _ := json.Marshal(m)
	if string(b) != `{"role":"user","content":"hi"}` {
		t.Fatalf("unexpected json: %s", b)
	}
	u, _ := json.Marshal(Usage{PromptTokens: 1, CompletionTokens: 2, TotalTokens: 3})
	if string(u) != `{"prompt_tokens":1,"completion_tokens":2,"total_tokens":3}` {
		t.Fatalf("unexpected usage json: %s", u)
	}
}
