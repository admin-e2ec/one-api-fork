package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common/helper"
)

func SetUpLogger(server *gin.Engine) {
	server.Use(gin.LoggerWithFormatter(formatAccessLog))
}

// formatAccessLog 输出访问日志：时间、One-API 请求 id、状态、耗时、来源 IP、方法和路径；
// 调用方传了合格的 X-Request-Id 时在行尾加 client_request_id。不记请求体、请求头和查询参数以外的内容。
func formatAccessLog(param gin.LogFormatterParams) string {
	requestID, _ := param.Keys[helper.RequestIdKey].(string)
	clientRequestID, _ := param.Keys[helper.ClientRequestIdKey].(string)
	suffix := ""
	if clientRequestID != "" {
		suffix = " | client_request_id=" + clientRequestID
	}
	return fmt.Sprintf("[GIN] %s | %s | %3d | %13v | %15s | %7s %s%s\n",
		param.TimeStamp.Format("2006/01/02 - 15:04:05"),
		requestID,
		param.StatusCode,
		param.Latency,
		param.ClientIP,
		param.Method,
		param.Path,
		suffix,
	)
}
