package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/model"
)

func withExposeChannelId(t *testing.T, v bool) {
	t.Helper()
	old := config.ExposeChannelIdHeader
	config.ExposeChannelIdHeader = v
	t.Cleanup(func() { config.ExposeChannelIdHeader = old })
}

func newTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/v1/chat/completions", nil)
	return c, w
}

func TestChannelIdHeaderReflectsLastSelectedChannel(t *testing.T) {
	withExposeChannelId(t, true)
	c, w := newTestContext()
	SetupContextForSelectedChannel(c, &model.Channel{Id: 3, Name: "litellm", Key: "k"}, "deepseek-v4-flash")
	SetupContextForSelectedChannel(c, &model.Channel{Id: 7, Name: "retry", Key: "k"}, "deepseek-v4-flash") // 重试换渠道
	if got := w.Header().Get(helper.ChannelIdHeader); got != "7" {
		t.Fatalf("%s = %q, want 7", helper.ChannelIdHeader, got)
	}
}

func TestChannelIdHeaderOffByDefault(t *testing.T) {
	withExposeChannelId(t, false)
	c, w := newTestContext()
	SetupContextForSelectedChannel(c, &model.Channel{Id: 3, Name: "litellm", Key: "k"}, "deepseek-v4-flash")
	if got := w.Header().Get(helper.ChannelIdHeader); got != "" {
		t.Fatalf("%s should be absent, got %q", helper.ChannelIdHeader, got)
	}
}
