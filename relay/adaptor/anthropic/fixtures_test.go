package anthropic

// noMarkerRequestJSON covers the main request shapes without any cache_control
// marker: string system, list system, text and image parts, tool calls and tool results.
const noMarkerRequestJSON = `{
  "model": "claude-3-5-sonnet-20241022",
  "max_tokens": 256,
  "messages": [
    {"role": "system", "content": "You are a helpful assistant."},
    {"role": "user", "content": [
      {"type": "text", "text": "Describe this image."},
      {"type": "image_url", "image_url": {"url": "data:image/png;base64,iVBORw0KGgo="}}
    ]},
    {"role": "assistant", "content": "", "tool_calls": [
      {"id": "call_1", "type": "function", "function": {"name": "get_weather", "arguments": "{\"city\":\"Paris\"}"}}
    ]},
    {"role": "tool", "tool_call_id": "call_1", "content": "sunny"},
    {"role": "user", "content": "Thanks"}
  ],
  "tools": [
    {"type": "function", "function": {"name": "get_weather", "description": "Get weather", "parameters": {"type": "object", "properties": {"city": {"type": "string"}}, "required": ["city"]}}}
  ]
}`

// noMarkerListSystemRequestJSON uses list content for the system message.
const noMarkerListSystemRequestJSON = `{
  "model": "claude-3-5-sonnet-20241022",
  "messages": [
    {"role": "system", "content": [{"type": "text", "text": "Part one. "}, {"type": "text", "text": "Part two."}]},
    {"role": "user", "content": "Hi"}
  ]
}`

// Golden outputs captured from ConvertRequest before cache_control support was added.
const noMarkerGolden = `{"model":"claude-3-5-sonnet-20241022","messages":[{"role":"user","content":[{"type":"text","text":"Describe this image."},{"type":"image","source":{"type":"base64","media_type":"image/png","data":"iVBORw0KGgo="}}]},{"role":"assistant","content":[{"type":"text"},{"type":"tool_use","id":"call_1","name":"get_weather","input":{"city":"Paris"}}]},{"role":"user","content":[{"type":"tool_result","content":"sunny","tool_use_id":"call_1"}]},{"role":"user","content":[{"type":"text","text":"Thanks"}]}],"system":"You are a helpful assistant.","max_tokens":256,"tools":[{"name":"get_weather","description":"Get weather","input_schema":{"type":"object","properties":{"city":{"type":"string"}},"required":["city"]}}],"tool_choice":{"type":"auto"}}`

const noMarkerListSystemGolden = `{"model":"claude-3-5-sonnet-20241022","messages":[{"role":"user","content":[{"type":"text","text":"Hi"}]}],"system":"Part one. Part two.","max_tokens":4096}`
