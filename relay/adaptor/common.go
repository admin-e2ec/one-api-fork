package adaptor

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common/client"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/relay/meta"
)

func SetupCommonRequestHeader(c *gin.Context, req *http.Request, meta *meta.Meta) {
	req.Header.Set("Content-Type", c.Request.Header.Get("Content-Type"))
	req.Header.Set("Accept", c.Request.Header.Get("Accept"))
	if meta.IsStream && c.Request.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "text/event-stream")
	}
	if config.RelayForwardRequestId {
		SetupRequestIdHeaders(c, req)
	}
}

// SetupRequestIdHeaders 把请求 id 转发给上游：X-Request-Id 用调用方的 id（没有时用 One-API 的 id），
// X-Oneapi-Request-Id 是 One-API 的 id；X-Litellm-Spend-Logs-Metadata 让 LiteLLM 把两个 id 记进 spend log。
// 两个 id 都只含 [A-Za-z0-9._:-]（One-API 的 id 是时间加数字，调用方的 id 经 helper.SanitizeClientRequestId 校验）。
func SetupRequestIdHeaders(c *gin.Context, req *http.Request) {
	oneapiId := c.GetString(helper.RequestIdKey)
	clientId := c.GetString(helper.ClientRequestIdKey)
	forwardId := clientId
	if forwardId == "" {
		forwardId = oneapiId
	}
	if forwardId != "" {
		req.Header.Set(helper.ClientRequestIdHeader, forwardId)
	}
	if oneapiId != "" {
		req.Header.Set(helper.RequestIdKey, oneapiId)
	}
	metadata := map[string]string{}
	if oneapiId != "" {
		metadata["oneapi_request_id"] = oneapiId
	}
	if clientId != "" {
		metadata["client_request_id"] = clientId
	}
	if len(metadata) == 0 {
		return
	}
	if encoded, err := json.Marshal(metadata); err == nil {
		req.Header.Set(helper.SpendLogsMetadataHeader, string(encoded))
	}
}

func DoRequestHelper(a Adaptor, c *gin.Context, meta *meta.Meta, requestBody io.Reader) (*http.Response, error) {
	fullRequestURL, err := a.GetRequestURL(meta)
	if err != nil {
		return nil, fmt.Errorf("get request url failed: %w", err)
	}
	req, err := http.NewRequest(c.Request.Method, fullRequestURL, requestBody)
	if err != nil {
		return nil, fmt.Errorf("new request failed: %w", err)
	}
	err = a.SetupRequestHeader(c, req, meta)
	if err != nil {
		return nil, fmt.Errorf("setup request header failed: %w", err)
	}
	resp, err := DoRequest(c, req)
	if err != nil {
		return nil, fmt.Errorf("do request failed: %w", err)
	}
	return resp, nil
}

func DoRequest(c *gin.Context, req *http.Request) (*http.Response, error) {
	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, errors.New("resp is nil")
	}
	_ = req.Body.Close()
	_ = c.Request.Body.Close()
	return resp, nil
}
