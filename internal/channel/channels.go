package channel

import (
	"fmt"
	"strconv"

	"ncpd/internal/client"
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
	client := client.Get()

	var channelsResponse ChannelsResponse
	_, err := client.R().
		SetResult(&channelsResponse).
		Get("/content_providers/channels")

	if err != nil {
		return nil, err
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
	client := client.Get()

	var channelDomainResponse ChannelDomainResponse
	_, err := client.R().
		SetPathParam("domain", domain).
		SetResult(&channelDomainResponse).
		Get("/content_providers/channel_domain?current_site_domain={domain}")

	if err != nil {
		return nil, err
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
	client := client.Get()

	var fanclubSiteInfoResponse FanclubSiteInfoResponse
	_, err := client.R().
		SetPathParam("siteId", strconv.Itoa(siteID)).
		SetResult(&fanclubSiteInfoResponse).
		Get("/fanclub_sites/{siteId}/page_base_info")

	if err != nil {
		return nil, err
	}

	return &fanclubSiteInfoResponse.Data.FanclubSite, nil
}
