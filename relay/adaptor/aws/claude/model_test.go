package aws

import (
	"encoding/json"
	"testing"

	"github.com/jinzhu/copier"
	"github.com/songquanpeng/one-api/relay/adaptor/anthropic"
	"github.com/songquanpeng/one-api/relay/model"
)

func marshalAWSRequest(t *testing.T, claudeReq *anthropic.Request) string {
	t.Helper()
	awsReq := &Request{AnthropicVersion: "bedrock-2023-05-31"}
	if err := copier.Copy(awsReq, claudeReq); err != nil {
		t.Fatalf("copy request: %v", err)
	}
	b, err := json.Marshal(awsReq)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	return string(b)
}

func TestRequestCopy_SystemString(t *testing.T) {
	claudeReq := &anthropic.Request{Model: "m", MaxTokens: 10, System: "sys"}
	want := `{"anthropic_version":"bedrock-2023-05-31","messages":null,"system":"sys","max_tokens":10}`
	if got := marshalAWSRequest(t, claudeReq); got != want {
		t.Fatalf("got %s\nwant %s", got, want)
	}
}

func TestRequestCopy_NoSystem(t *testing.T) {
	claudeReq := &anthropic.Request{Model: "m", MaxTokens: 10}
	want := `{"anthropic_version":"bedrock-2023-05-31","messages":null,"max_tokens":10}`
	if got := marshalAWSRequest(t, claudeReq); got != want {
		t.Fatalf("got %s\nwant %s", got, want)
	}
}

func TestRequestCopy_SystemBlocks(t *testing.T) {
	claudeReq := &anthropic.Request{Model: "m", MaxTokens: 10, System: []anthropic.Content{
		{Type: "text", Text: "sys", CacheControl: &model.CacheControl{Type: "ephemeral"}},
	}}
	want := `{"anthropic_version":"bedrock-2023-05-31","messages":null,"system":[{"type":"text","text":"sys","cache_control":{"type":"ephemeral"}}],"max_tokens":10}`
	if got := marshalAWSRequest(t, claudeReq); got != want {
		t.Fatalf("got %s\nwant %s", got, want)
	}
}
