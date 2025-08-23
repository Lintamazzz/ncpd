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

	// 将 videoList 转换为 JSON 格式
	jsonData, err := json.MarshalIndent(videoList, "", "  ")
	if err != nil {
		log.Fatalf("JSON 序列化失败: %v", err)
	}

	// 写入到 videoList.json 文件
	outPath := "./out/all_videos.json"
	err = os.WriteFile(outPath, jsonData, 0644)
	if err != nil {
		log.Fatalf("写入文件失败: %v", err)
	}
	fmt.Printf("\n已成功输出到 %s\n", outPath)

	// 测试打印视频详情
	fmt.Printf("\n测试打印视频详情: \n")
	videoDetails, _ := video.GetVideoDetails(387, videoList[0].ContentCode)
	jsonData, _ = json.MarshalIndent(videoDetails, "", "  ")
	fmt.Println(string(jsonData))
}
