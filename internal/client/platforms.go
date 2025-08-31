package client

import (
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
)

type Platform struct {
	Name              string
	Domain            string
	DefaultAPIBaseURL string
	TemplateFile      string
}

// 支持的平台列表
var SupportedPlatforms = []Platform{
	{Name: "Nicochannel+", Domain: "nicochannel.jp", DefaultAPIBaseURL: "https://api.nicochannel.jp/fc", TemplateFile: "assets/template_white_bg.html"},
	{Name: "QloveR", Domain: "qlover.jp", DefaultAPIBaseURL: "https://api.qlover.jp/fc", TemplateFile: "assets/template_black_bg.html"},
}

// 当前选择的平台
var CurrentPlatform *Platform = &SupportedPlatforms[0] // 默认设置为 Nicochannel+

func SetCurrentPlatform(platform *Platform) {
	CurrentPlatform = platform
}

type SiteSettings struct {
	PlatformID     string `json:"platform_id"`
	FanclubSiteID  string `json:"fanclub_site_id"`
	FanclubGroupID string `json:"fanclub_group_id"`
	APIBaseURL     string `json:"api_base_url"`
}

// GetAPIBaseURL 根据平台获取 API base URL
func GetAPIBaseURL(platform *Platform) (string, error) {
	var settings SiteSettings
	resp, err := resty.New().R().
		SetResult(&settings).
		SetPathParam("domain", platform.Domain).
		Get("https://{domain}/site/settings.json")

	if err != nil {
		return "", err
	}

	if resp.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("状态码 %d", resp.StatusCode())
	}

	if settings.APIBaseURL == "" {
		return "", fmt.Errorf("api_base_url 为空")
	}

	return settings.APIBaseURL, nil
}

// InitClientWithPlatform 根据平台设置 Resty 客户端的 Base URL
func InitClientWithPlatform(platform *Platform) error {
	apiBaseURL, err := GetAPIBaseURL(platform)
	if err != nil {
		return fmt.Errorf("获取 API base URL 失败: %w", err)
	}

	// 确保客户端已初始化
	Get()

	// 设置新的BaseURL
	restyClient.SetBaseURL(apiBaseURL)

	// 设置当前平台
	SetCurrentPlatform(platform)

	return nil
}
