package anthropic

import "github.com/songquanpeng/one-api/relay/model"

// Prompt caching helpers.
// https://docs.anthropic.com/en/docs/build-with-claude/prompt-caching

func copyCacheControl(cc *model.CacheControl) *model.CacheControl {
	if cc == nil {
		return nil
	}
	c := *cc
	return &c
}

// hasCacheControl reports whether the message or any of its content parts carries a marker.
func hasCacheControl(message model.Message) bool {
	if message.CacheControl != nil {
		return true
	}
	if message.IsStringContent() {
		return false
	}
	for _, part := range message.ParseContent() {
		if part.CacheControl != nil {
			return true
		}
	}
	return false
}

// withLastBlockCacheControl returns a copy of blocks whose last block carries cc.
// A message-level marker applies to the end of that message.
func withLastBlockCacheControl(blocks []Content, cc *model.CacheControl) []Content {
	if cc == nil || len(blocks) == 0 {
		return blocks
	}
	out := make([]Content, len(blocks))
	copy(out, blocks)
	out[len(out)-1].CacheControl = copyCacheControl(cc)
	return out
}

// convertSystem converts an OpenAI system message to the Anthropic system field.
// Without markers it keeps the plain string form; with markers it returns text blocks.
// It returns nil when the system prompt is empty.
func convertSystem(message model.Message) any {
	if !hasCacheControl(message) {
		if text := message.StringContent(); text != "" {
			return text
		}
		return nil
	}
	blocks := make([]Content, 0)
	for _, part := range message.ParseContent() {
		if part.Type != model.ContentTypeText || part.Text == "" {
			continue
		}
		blocks = append(blocks, Content{
			Type:         "text",
			Text:         part.Text,
			CacheControl: copyCacheControl(part.CacheControl),
		})
	}
	if len(blocks) == 0 {
		return nil
	}
	return withLastBlockCacheControl(blocks, message.CacheControl)
}

func nonNegative(v int) int {
	if v < 0 {
		return 0
	}
	return v
}

// withCacheUsage adds Anthropic cache counters to an OpenAI usage.
// Following the OpenAI convention, prompt_tokens includes cached tokens.
// Usage without cache tokens is returned unchanged.
func withCacheUsage(usage model.Usage, cacheRead, cacheCreation int) model.Usage {
	cacheRead = nonNegative(cacheRead)
	cacheCreation = nonNegative(cacheCreation)
	if cacheRead == 0 && cacheCreation == 0 {
		return usage
	}
	usage.PromptTokens += cacheRead + cacheCreation
	usage.CacheReadInputTokens = cacheRead
	usage.CacheCreationInputTokens = cacheCreation
	usage.PromptTokensDetails = &model.PromptTokensDetails{CachedTokens: cacheRead}
	return usage
}

// ConvertUsage maps a complete Anthropic usage object to OpenAI usage.
func ConvertUsage(u Usage) model.Usage {
	usage := withCacheUsage(model.Usage{
		PromptTokens:     u.InputTokens,
		CompletionTokens: u.OutputTokens,
	}, u.CacheReadInputTokens, u.CacheCreationInputTokens)
	usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
	return usage
}
