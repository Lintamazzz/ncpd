package main

import (
	"fmt"
	"log"
	"ncpd/internal/auth"
	"ncpd/internal/m3u8"
	"ncpd/internal/video"
)

func main() {
	videoList, _ := video.GetVideoList(387)
	token, err := auth.GetToken()
	if err != nil {
		log.Fatal("获取 token 失败")
	}

	for i, video := range videoList {
		sessionID, _ := auth.GetSessionID(video.ContentCode, token)
		index, _ := m3u8.GetIndex(sessionID)
		streamInfo := m3u8.ParseIndexM3U8(index)
		bestQuality := m3u8.GetBestQuality(streamInfo)

		fmt.Printf("\n%d. %s\n", i+1, video.Title)
		fmt.Printf("   视频代码: %s\n", video.ContentCode)
		fmt.Printf("   最高画质: %s %s\n", bestQuality.Resolution, bestQuality.FrameRate)
		fmt.Printf("   下载地址: %s\n", bestQuality.URL)
	}
}
