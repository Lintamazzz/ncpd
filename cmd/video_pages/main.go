package main

import (
	"encoding/json"
	"fmt"
	"log"
	"ncpd/internal/video"
	"os"
)

func main() {
	videoList, _ := video.GetVideoList(387)

	// 打印最终统计信息
	fmt.Printf("\n=== 数据获取完成 ===\n")
	fmt.Printf("总共获取到 %d 个视频\n", len(videoList))

	// 打印所有视频的详细信息
	fmt.Println("\n=== 所有视频列表 ===")
	for i, video := range videoList {
		fmt.Printf("\n%d. %s\n", i+1, video.Title)
		fmt.Printf("   视频代码: %s\n", video.ContentCode)
		fmt.Printf("   显示日期: %s\n", video.DisplayDate)
		fmt.Printf("   发布时间: %s\n", video.ReleasedAt)
		fmt.Printf("   缩略图: %s\n", video.ThumbnailURL)
		fmt.Printf("   总观看数: %d\n", video.VideoAggregateInfo.TotalViews)
		fmt.Printf("   评论数: %d\n", video.VideoAggregateInfo.NumberOfComments)
		fmt.Printf("   视频长度: %d 秒\n", video.ActiveVideoFilename.Length)

		if len(video.VideoFreePeriods) > 0 {
			fmt.Printf("   免费观看时段: %d 个\n", len(video.VideoFreePeriods))
		}
	}

	videoIndex := 36

	// 测试打印视频详情
	fmt.Printf("\n测试打印视频详情: \n")
	videoDetails, _ := video.GetVideoDetails(387, videoList[videoIndex].ContentCode)
	jsonData, _ := json.MarshalIndent(videoDetails, "", "  ")
	fmt.Println(string(jsonData))

	// 获取 comments_user_token
	commentsUserToken, _ := video.GetCommentsUserToken(videoList[videoIndex].ContentCode)
	fmt.Println(commentsUserToken)

	// 获取弹幕
	// comments, _ := video.GetComments(commentsUserToken, videoDetails.VideoCommentSetting.CommentGroupID, 0)
	// jsonData, _ = json.MarshalIndent(comments, "", "  ")
	// fmt.Println(string(jsonData))

	// 获取所有弹幕
	allComments, _ := video.GetAllComments(commentsUserToken, videoDetails.VideoCommentSetting.CommentGroupID)

	// 虽然不是和视频详情里记录的评论数完全一致，但基本非常接近
	// 实际获取到的数量和 nicochannel_comments 获取到的数量是一致的
	fmt.Printf("应有弹幕 %d 条，实际获取到 %d 条\n", videoDetails.VideoAggregateInfo.NumberOfComments, len(allComments))
	toJsonFile(allComments, "./out/all_comments.json")
}

func toJsonFile(data any, outPath string) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Fatalf("JSON 序列化失败: %v", err)
	}

	err = os.WriteFile(outPath, jsonData, 0644)
	if err != nil {
		log.Fatalf("写入文件失败: %v", err)
	}
	fmt.Printf("\n已成功输出到 %s\n", outPath)
}
