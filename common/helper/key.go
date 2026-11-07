package helper

const (
	RequestIdKey = "X-Oneapi-Request-Id"
	// ClientRequestIdKey 是调用方经请求头 X-Request-Id 传来的请求 id（校验后）在 gin / context 里的键
	ClientRequestIdKey = "X-Client-Request-Id"
	// ClientRequestIdHeader 是调用方传入请求 id 的请求头；RELAY_FORWARD_REQUEST_ID=true 时也用这个头转发给上游
	ClientRequestIdHeader = "X-Request-Id"
	// SpendLogsMetadataHeader 是 LiteLLM 读取的请求头：其中的 JSON 写进 LiteLLM spend log 的 metadata.spend_logs_metadata
	SpendLogsMetadataHeader = "X-Litellm-Spend-Logs-Metadata"
	// ChannelIdHeader 是响应头：本次请求最后选中的渠道 id（EXPOSE_CHANNEL_ID_HEADER=true 时返回）
	ChannelIdHeader = "X-Oneapi-Channel-Id"
)
