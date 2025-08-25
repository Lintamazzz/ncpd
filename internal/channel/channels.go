package channel

import (
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
)

type ChannelsResponse struct {
	Data struct {
		ContentProviders      []ContentProvider `json:"content_providers"`
		ParentFanclubSiteName *string           `json:"parent_fanclub_site_name"`
		ParentUseNfcApp       *bool             `json:"parent_use_nfc_app"`
	} `json:"data"`
}

type ChannelDomainResponse struct {
	Data struct {
		ContentProviders *ContentProvider `json:"content_providers"`
	} `json:"data"`
}

type ContentProvider struct {
	Domain      string      `json:"domain"`
	FanclubSite FanclubSite `json:"fanclub_site"`
	ID          int         `json:"id"`
}

type FanclubSite struct {
	ID int `json:"id"`
}

// 获取频道列表
func GetChannels() (*ChannelsResponse, error) {
	client := resty.New()

	var channelsResponse ChannelsResponse
	resp, err := client.R().
		SetResult(&channelsResponse).
		Get("https://api.nicochannel.jp/fc/content_providers/channels")

	if err != nil {
		return nil, fmt.Errorf("channel.GetChannels: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("channel.GetChannels: 状态码 %d", resp.StatusCode())
	}

	return &channelsResponse, nil
}

// 获取频道列表（简化版本，只返回 ContentProvider 数组）
func GetChannelList() ([]ContentProvider, error) {
	response, err := GetChannels()
	if err != nil {
		return nil, err
	}

	return response.Data.ContentProviders, nil
}

// 根据 FC Site ID 查找特定频道
func GetChannelByID(id int) (*ContentProvider, error) {
	channels, err := GetChannelList()
	if err != nil {
		return nil, err
	}

	for _, channel := range channels {
		if channel.FanclubSite.ID == id {
			return &channel, nil
		}
	}

	return nil, fmt.Errorf("channel.GetChannelByID: 未找到 ID 为 %d 的频道", id)
}

// 根据域名查找特定频道的 ID
func GetChannelByDomain(domain string) (*ContentProvider, error) {
	client := resty.New()

	baseURL := "https://api.nicochannel.jp/fc/content_providers/channel_domain?current_site_domain=%s"
	URL := fmt.Sprintf(baseURL, domain)

	var channelDomainResponse ChannelDomainResponse
	resp, err := client.R().
		SetResult(&channelDomainResponse).
		Get(URL)

	if err != nil {
		return nil, fmt.Errorf("channel.GetChannelByDomain: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("channel.GetChannelByDomain: 状态码 %d", resp.StatusCode())
	}

	if channelDomainResponse.Data.ContentProviders == nil {
		return nil, fmt.Errorf("channel.GetChannelByDomain: 未找到域名为 %s 的频道", domain)
	}

	return channelDomainResponse.Data.ContentProviders, nil
}

// FanclubSiteInfoResponse 表示 fanclub site 信息的 API 响应
type FanclubSiteInfoResponse struct {
	Data struct {
		FanclubSite FanclubSiteInfo `json:"fanclub_site"`
	} `json:"data"`
}

// FanclubSiteInfo 表示 fanclub site 的详细信息
type FanclubSiteInfo struct {
	Description       string `json:"description"`
	FanclubSiteName   string `json:"fanclub_site_name"`
	FaviconURL        string `json:"favicon_url"`
	ThumbnailImageURL string `json:"thumbnail_image_url"`
}

// GetFanclubSiteInfo 根据 fc site id 获取 fanclub 信息
func GetFanclubSiteInfo(siteID int) (*FanclubSiteInfo, error) {
	client := resty.New()

	baseURL := "https://api.nicochannel.jp/fc/fanclub_sites/%d/page_base_info"
	URL := fmt.Sprintf(baseURL, siteID)

	var fanclubSiteInfoResponse FanclubSiteInfoResponse
	resp, err := client.R().
		SetResult(&fanclubSiteInfoResponse).
		Get(URL)

	if err != nil {
		return nil, fmt.Errorf("channel.GetFanclubSiteInfo: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("channel.GetFanclubSiteInfo: 状态码 %d", resp.StatusCode())
	}

	return &fanclubSiteInfoResponse.Data.FanclubSite, nil
}
