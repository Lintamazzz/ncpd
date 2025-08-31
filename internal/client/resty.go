package client

import (
	"fmt"
	"log"
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

	// 设置 API Base URL
	baseURL, err := getAPIBaseURL()
	if err != nil {
		log.Printf("获取站点设置失败: %v，使用默认值", err)
		baseURL = "https://api.nicochannel.jp/fc"
	}
	restyClient.SetBaseURL(baseURL)
	log.Printf("设置 API Base URL: %s", baseURL)

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

type SiteSettings struct {
	PlatformID     string `json:"platform_id"`
	FanclubSiteID  string `json:"fanclub_site_id"`
	FanclubGroupID string `json:"fanclub_group_id"`
	APIBaseURL     string `json:"api_base_url"`
}

func getAPIBaseURL() (string, error) {
	var settings SiteSettings
	resp, err := resty.New().R().
		SetResult(&settings).
		Get("https://nicochannel.jp/site/settings.json")

	if err != nil {
		return "", err
	}

	if resp.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("状态码 %d", resp.StatusCode())
	}

	return settings.APIBaseURL, nil
}
