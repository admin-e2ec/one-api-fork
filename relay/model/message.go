package model

type Message struct {
	Role             string  `json:"role,omitempty"`
	Content          any     `json:"content,omitempty"`
	ReasoningContent any     `json:"reasoning_content,omitempty"`
	Name             *string `json:"name,omitempty"`
	ToolCalls        []Tool  `json:"tool_calls,omitempty"`
	ToolCallId       string  `json:"tool_call_id,omitempty"`
	// CacheControl is a message-level prompt caching marker (used by Anthropic-compatible channels).
	CacheControl *CacheControl `json:"cache_control,omitempty"`
}

// CacheControl is a prompt caching marker, e.g. {"type":"ephemeral","ttl":"1h"}.
type CacheControl struct {
	Type string `json:"type"`
	TTL  string `json:"ttl,omitempty"`
}

// parseCacheControl reads a cache_control value decoded into a generic map.
// It returns nil when the value is missing or has no type.
func parseCacheControl(v any) *CacheControl {
	m, ok := v.(map[string]any)
	if !ok {
		return nil
	}
	typ, _ := m["type"].(string)
	if typ == "" {
		return nil
	}
	ttl, _ := m["ttl"].(string)
	return &CacheControl{Type: typ, TTL: ttl}
}

func (m Message) IsStringContent() bool {
	_, ok := m.Content.(string)
	return ok
}

func (m Message) StringContent() string {
	content, ok := m.Content.(string)
	if ok {
		return content
	}
	contentList, ok := m.Content.([]any)
	if ok {
		var contentStr string
		for _, contentItem := range contentList {
			contentMap, ok := contentItem.(map[string]any)
			if !ok {
				continue
			}
			if contentMap["type"] == ContentTypeText {
				if subStr, ok := contentMap["text"].(string); ok {
					contentStr += subStr
				}
			}
		}
		return contentStr
	}
	return ""
}

func (m Message) ParseContent() []MessageContent {
	var contentList []MessageContent
	content, ok := m.Content.(string)
	if ok {
		contentList = append(contentList, MessageContent{
			Type: ContentTypeText,
			Text: content,
		})
		return contentList
	}
	anyList, ok := m.Content.([]any)
	if ok {
		for _, contentItem := range anyList {
			contentMap, ok := contentItem.(map[string]any)
			if !ok {
				continue
			}
			switch contentMap["type"] {
			case ContentTypeText:
				if subStr, ok := contentMap["text"].(string); ok {
					contentList = append(contentList, MessageContent{
						Type:         ContentTypeText,
						Text:         subStr,
						CacheControl: parseCacheControl(contentMap["cache_control"]),
					})
				}
			case ContentTypeImageURL:
				if subObj, ok := contentMap["image_url"].(map[string]any); ok {
					contentList = append(contentList, MessageContent{
						Type: ContentTypeImageURL,
						ImageURL: &ImageURL{
							Url: subObj["url"].(string),
						},
						CacheControl: parseCacheControl(contentMap["cache_control"]),
					})
				}
			}
		}
		return contentList
	}
	return nil
}

type ImageURL struct {
	Url    string `json:"url,omitempty"`
	Detail string `json:"detail,omitempty"`
}

type MessageContent struct {
	Type     string    `json:"type,omitempty"`
	Text     string    `json:"text"`
	ImageURL *ImageURL `json:"image_url,omitempty"`
	// CacheControl is a part-level prompt caching marker.
	CacheControl *CacheControl `json:"cache_control,omitempty"`
}
