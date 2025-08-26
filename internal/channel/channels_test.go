package channel

import (
	"testing"
)

// go test -v ./internal/channel
func TestGetChannelList(t *testing.T) {
	channels, err := GetChannelList()
	if err != nil {
		t.Fatalf("获取频道列表失败: %v", err)
	}

	for i := 0; i < 10; i++ {
		channel := channels[i]
		t.Logf("频道 %d: %v", channel.ID, channel.Domain)
	}
}

func TestGetChannelByID(t *testing.T) {
	siteID := 387
	channel, err := GetChannelByID(siteID)
	if err != nil {
		t.Fatalf("获取频道失败: %v", err)
	}
	t.Logf("获取到的频道: %v", channel)
}

func TestGetChannelByDomain(t *testing.T) {
	channel, err := GetChannelByDomain("https://nicochannel.jp/sakakura-sakura")
	if err != nil {
		t.Fatalf("获取频道失败: %v", err)
	}
	t.Logf("获取到的频道: %v", channel)
}

func TestGetFanclubSiteInfo(t *testing.T) {
	siteID := 387
	fanclubInfo, err := GetFanclubSiteInfo(siteID)
	if err != nil {
		t.Fatalf("获取 fanclub site 信息失败: %v", err)
	}

	t.Logf("=== 获取到的结果 ===")
	t.Logf("Fanclub Site ID: %d", siteID)
	t.Logf("名称: %s", fanclubInfo.FanclubSiteName)
	t.Logf("描述: %s", fanclubInfo.Description)
	t.Logf("Favicon URL: %s", fanclubInfo.FaviconURL)
	t.Logf("缩略图 URL: %s", fanclubInfo.ThumbnailImageURL)

	t.Logf("=== 输出目录结构 ===")
	t.Logf("视频目录: ./out/%s/動画/", fanclubInfo.FanclubSiteName)
	t.Logf("生放送目录: ./out/%s/生放送/", fanclubInfo.FanclubSiteName)
	t.Logf("新闻目录: ./out/%s/NEWS/", fanclubInfo.FanclubSiteName)
}
