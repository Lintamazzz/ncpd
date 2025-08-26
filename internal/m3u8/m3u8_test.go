package m3u8

import (
	"testing"

	"ncpd/internal/auth"
	"ncpd/internal/video"
)

// go test -v ./internal/m3u8/
func TestM3U8Workflow(t *testing.T) {
	videoList, err := video.GetVideoList(387)
	if err != nil {
		t.Fatalf("获取视频列表失败: %v", err)
	}

	token, err := auth.GetToken()
	if err != nil {
		t.Fatal("获取 token 失败")
	}

	for i, video := range videoList {
		sessionID, err := auth.GetSessionID(video.ContentCode, token)
		if err != nil {
			t.Logf("获取 session ID 失败: %v", err)
			continue
		}

		index, err := GetIndex(sessionID)
		if err != nil {
			t.Logf("获取 index.m3u8 失败: %v", err)
			continue
		}

		streamInfo := ParseIndexM3U8(index)
		bestQuality := GetBestQuality(streamInfo)

		if bestQuality == nil {
			t.Logf("未找到最佳画质信息")
			continue
		}

		t.Logf("  %d. %s", i+1, video.Title)
		t.Logf("   视频代码: %s", video.ContentCode)
		t.Logf("   最高画质: %s %s", bestQuality.Resolution, bestQuality.FrameRate)
		t.Logf("   下载地址: %s", bestQuality.URL)
	}
}

func TestParseIndexM3U8(t *testing.T) {
	// 测试数据
	testM3U8Content := `#EXTM3U
#EXT-X-STREAM-INF:BANDWIDTH=1000000,AVERAGE-BANDWIDTH=950000,CODECS="avc1.64001f,mp4a.40.2",RESOLUTION=1280x720,FRAME-RATE=30.000
https://example.com/stream_720p.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=2000000,AVERAGE-BANDWIDTH=1900000,CODECS="avc1.640028,mp4a.40.2",RESOLUTION=1920x1080,FRAME-RATE=30.000
https://example.com/stream_1080p.m3u8`

	streams := ParseIndexM3U8(testM3U8Content)

	if len(streams) != 2 {
		t.Fatalf("期望解析出 2 个流，实际解析出 %d 个", len(streams))
	}

	t.Logf("  解析出的流信息: ")
	for _, stream := range streams {
		t.Logf("  %s %s %d", stream.Resolution, stream.FrameRate, stream.Bandwidth)
	}
}

func TestGetBestQuality(t *testing.T) {
	streams := []StreamInfo{
		{Bandwidth: 1000000, Resolution: "1280x720", FrameRate: "30.000"},
		{Bandwidth: 2000000, Resolution: "1920x1080", FrameRate: "30.000"},
		{Bandwidth: 500000, Resolution: "854x480", FrameRate: "30.000"},
	}

	bestQuality := GetBestQuality(streams)
	if bestQuality == nil {
		t.Fatal("GetBestQuality 返回了 nil")
	}

	t.Logf("最佳画质: %s %s", bestQuality.Resolution, bestQuality.FrameRate)
}
