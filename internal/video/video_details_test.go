package video

import (
	"encoding/json"
	"testing"
)

// go test -v ./internal/video -run TestGetVideoDetails
func TestGetVideoDetails(t *testing.T) {
	fcSiteID := 387
	contentCode := "smQKzZSkFT4Fap6ERziVr26f"

	t.Log("🔍 正在获取视频详细信息...")
	videoDetails, err := GetVideoDetails(fcSiteID, contentCode)
	if err != nil {
		t.Fatalf("获取视频详情失败: %v", err)
	}

	// 输出视频详情信息
	t.Logf("✅ 成功获取视频详情:")
	t.Logf("  标题: %s", videoDetails.Title)
	t.Logf("  视频代码: %s", videoDetails.ContentCode)
	t.Logf("  显示日期: %s", videoDetails.DisplayDate)
	t.Logf("  发布时间: %s", videoDetails.ReleasedAt)
	t.Logf("  总观看数: %d", videoDetails.VideoAggregateInfo.TotalViews)
	t.Logf("  评论数: %d", videoDetails.VideoAggregateInfo.NumberOfComments)
	t.Logf("  视频长度: %d 秒", videoDetails.ActiveVideoFilename.Length)

	// 输出评论设置信息
	if videoDetails.VideoCommentSetting != nil {
		t.Logf("  评论组ID: %s", videoDetails.VideoCommentSetting.CommentGroupID)
	}

	// 输出完整的JSON格式（用于调试）
	jsonData, _ := json.MarshalIndent(videoDetails, "", "  ")
	t.Logf("📄 完整视频详情JSON:")
	t.Log(string(jsonData))
}
