package main

import (
	"fmt"
	"log"

	"ncpd/internal/channel"
)

func main() {
	// 测试获取 fanclub site 信息（使用示例中的 site ID 387）
	siteID := 387
	fanclubInfo, err := channel.GetFanclubSiteInfo(siteID)
	if err != nil {
		log.Fatalf("获取 fanclub site 信息失败: %v", err)
	}

	fmt.Printf("Fanclub Site ID: %d\n", siteID)
	fmt.Printf("名称: %s\n", fanclubInfo.FanclubSiteName)
	fmt.Printf("描述: %s\n", fanclubInfo.Description)
	fmt.Printf("Favicon URL: %s\n", fanclubInfo.FaviconURL)
	fmt.Printf("缩略图 URL: %s\n", fanclubInfo.ThumbnailImageURL)

	// 测试新的目录结构
	fmt.Printf("\n=== 新的目录结构示例 ===\n")
	fmt.Printf("视频目录: ./out/%s/動画/\n", fanclubInfo.FanclubSiteName)
	fmt.Printf("生放送目录: ./out/%s/生放送/\n", fanclubInfo.FanclubSiteName)
	fmt.Printf("新闻目录: ./out/%s/NEWS/\n", fanclubInfo.FanclubSiteName)
}
