package helper

import "context"

// MaxClientRequestIdLength 是调用方请求 id 的最大长度；UUID 是 36 个字符
const MaxClientRequestIdLength = 128

type clientRequestIdCtxKey struct{}

// SanitizeClientRequestId 校验调用方传入的请求 id：1–128 个字符，只允许字母、数字和 . _ : -。
// 不合格时返回空串（当作没有传），不截断、不改写，避免把日志注入用的换行、空格、引号写进日志和上游请求头。
func SanitizeClientRequestId(raw string) string {
	if len(raw) == 0 || len(raw) > MaxClientRequestIdLength {
		return ""
	}
	for i := 0; i < len(raw); i++ {
		if !isClientRequestIdChar(raw[i]) {
			return ""
		}
	}
	return raw
}

func isClientRequestIdChar(b byte) bool {
	switch {
	case b >= 'a' && b <= 'z', b >= 'A' && b <= 'Z', b >= '0' && b <= '9':
		return true
	case b == '.', b == '_', b == ':', b == '-':
		return true
	}
	return false
}

func SetClientRequestId(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, clientRequestIdCtxKey{}, id)
}

// GetClientRequestId 返回 context 里调用方的请求 id；没有时返回空串
func GetClientRequestId(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	id, _ := ctx.Value(clientRequestIdCtxKey{}).(string)
	return id
}
