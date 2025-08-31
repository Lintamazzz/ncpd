package client

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/go-resty/resty/v2"
)

var (
	restyClient *resty.Client
	once        sync.Once
)

// Get 返回全局的 resty 客户端实例（单例）
func Get() *resty.Client {
	once.Do(initClient)
	return restyClient
}

// initClient 初始化全局 resty 客户端
func initClient() {
	restyClient = resty.New()

	// 默认 Base URL
	restyClient.SetBaseURL(CurrentPlatform.DefaultAPIBaseURL)

	// 统一错误处理
	restyClient.OnAfterResponse(func(c *resty.Client, resp *resty.Response) error {
		if resp.IsError() {
			return &HTTPError{
				StatusCode: resp.StatusCode(),
				StatusText: http.StatusText(resp.StatusCode()),
				URL:        resp.Request.URL,
			}
		}
		return nil
	})
}

type HTTPError struct {
	StatusCode int
	StatusText string
	URL        string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d %s", e.StatusCode, e.StatusText)
}
