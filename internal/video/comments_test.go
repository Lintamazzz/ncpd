package video

import (
	"testing"
)

// go test -v ./internal/video -run TestGetAllComments
func TestGetAllComments(t *testing.T) {
	contentCode := "smQKzZSkFT4Fap6ERziVr26f"
	fcSiteID := 387

	// 1. 获取视频详情以获取评论组ID
	t.Log("🔍 正在获取视频详情以获取评论组ID...")
	videoDetails, err := GetVideoDetails(fcSiteID, contentCode)
	if err != nil {
		t.Fatalf("获取视频详情失败: %v", err)
	}

	if videoDetails.VideoCommentSetting == nil {
		t.Fatal("视频评论设置为空")
	}

	commentGroupID := videoDetails.VideoCommentSetting.CommentGroupID
	if commentGroupID == "" {
		t.Fatal("评论组ID为空")
	}

	t.Logf("✅ 获取到评论组ID: %s", commentGroupID)

	// 2. 获取 comments_user_token
	t.Log("🔍 正在获取 comments_user_token...")
	commentsUserToken, err := GetCommentsUserToken(contentCode)
	if err != nil {
		t.Fatalf("获取 comments_user_token 失败: %v", err)
	}

	if commentsUserToken == "" {
		t.Fatal("comments_user_token 为空")
	}

	t.Logf("✅ 获取到 comments_user_token: %s", commentsUserToken)

	// 3. 获取所有评论
	t.Log("🔍 正在获取所有评论...")
	allComments, err := GetAllComments(commentsUserToken, commentGroupID)
	if err != nil {
		t.Fatalf("获取评论失败: %v", err)
	}

	t.Logf("✅ 成功获取到 %d 条评论", len(allComments))

	// 4. 验证评论数据
	if len(allComments) == 0 {
		t.Log("没有获取到任何评论")
		return
	}

	// 输出评论统计信息
	t.Logf("📊 评论统计:")
	t.Logf("  应有评论数: %d", videoDetails.VideoAggregateInfo.NumberOfComments)
	t.Logf("  实际获取数: %d", len(allComments))

	// 输出前几条评论的详细信息
	t.Logf("📝 前5条评论预览:")
	for i, comment := range allComments {
		if i >= 5 {
			break
		}
		t.Logf("  %d. [%s] %s", i+1, comment.SenderID, comment.Message)
	}
}
