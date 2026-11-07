package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common/helper"
)

func RequestId() func(c *gin.Context) {
	return func(c *gin.Context) {
		id := helper.GenRequestID()
		c.Set(helper.RequestIdKey, id)
		ctx := helper.SetRequestID(c.Request.Context(), id)
		// 调用方的请求 id（X-Request-Id）：校验合格才记录，不替换 One-API 自己的 id（日志表和响应头仍用 One-API 的 id）
		if clientId := helper.SanitizeClientRequestId(c.GetHeader(helper.ClientRequestIdHeader)); clientId != "" {
			c.Set(helper.ClientRequestIdKey, clientId)
			ctx = helper.SetClientRequestId(ctx, clientId)
		}
		c.Request = c.Request.WithContext(ctx)
		c.Header(helper.RequestIdKey, id)
		c.Next()
	}
}
