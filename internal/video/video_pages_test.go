package video

import (
	"testing"
)

// go test -v ./internal/video -run TestGetVideoList
func TestGetVideoList(t *testing.T) {
	fcSiteID := 387

	t.Log("🔍 正在获取视频列表...")
	videoList, err := GetVideoList(fcSiteID)
	if err != nil {
		t.Fatalf("获取视频列表失败: %v", err)
	}

	if len(videoList) == 0 {
		t.Log("获取到的视频列表为空")
		return
	}

	t.Logf("✅ 成功获取到 %d 个视频", len(videoList))

	// 输出视频列表信息
	t.Log("📋 视频列表:")
	for i, video := range videoList {
		t.Logf("  %d. [%s] %s", i+1, video.ContentCode, video.Title)
		t.Logf("     显示日期: %s", video.DisplayDate)
		t.Logf("     发布时间: %s", video.ReleasedAt)
		t.Logf("     总观看数: %d", video.VideoAggregateInfo.TotalViews)
		t.Logf("     评论数: %d", video.VideoAggregateInfo.NumberOfComments)
		t.Logf("     视频长度: %d 秒", video.ActiveVideoFilename.Length)
	}
}
