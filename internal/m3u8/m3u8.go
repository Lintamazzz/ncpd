package m3u8

import (
	"regexp"
	"strconv"
	"strings"

	"ncpd/internal/client"
)

// StreamInfo 表示单个流的详细信息
type StreamInfo struct {
	Bandwidth        int    `json:"bandwidth"`
	AverageBandwidth int    `json:"average_bandwidth"`
	Codecs           string `json:"codecs"`
	Resolution       string `json:"resolution"`
	FrameRate        string `json:"frame_rate"`
	URL              string `json:"url"`
}

// ParseIndexM3U8 解析 index.m3u8 文件内容，提取所有流信息
func ParseIndexM3U8(content string) []StreamInfo {
	lines := strings.Split(content, "\n")
	var streams []StreamInfo
	var currentStream StreamInfo

	// 匹配EXT-X-STREAM-INF行的正则表达式
	streamInfRegex := regexp.MustCompile(`BANDWIDTH=(\d+),AVERAGE-BANDWIDTH=(\d+),CODECS="([^"]+)",RESOLUTION=(\d+x\d+),FRAME-RATE=([\d.]+)`)

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// 检查是否是EXT-X-STREAM-INF行
		if strings.HasPrefix(line, "#EXT-X-STREAM-INF:") {
			matches := streamInfRegex.FindStringSubmatch(line)
			if len(matches) == 6 {
				bandwidth, _ := strconv.Atoi(matches[1])
				avgBandwidth, _ := strconv.Atoi(matches[2])

				currentStream = StreamInfo{
					Bandwidth:        bandwidth,
					AverageBandwidth: avgBandwidth,
					Codecs:           matches[3],
					Resolution:       matches[4],
					FrameRate:        matches[5],
				}
			}
		} else if strings.HasPrefix(line, "http") && currentStream.Resolution != "" {
			// 这是URL行，且前面有流信息
			currentStream.URL = line
			streams = append(streams, currentStream)
			currentStream = StreamInfo{} // 重置
		}
	}

	return streams
}

// GetBestQuality 从流列表中获取最佳画质（最高分辨率）
func GetBestQuality(streams []StreamInfo) *StreamInfo {
	if len(streams) == 0 {
		return nil
	}

	bestStream := streams[0]
	for _, stream := range streams {
		if stream.Bandwidth > bestStream.Bandwidth {
			bestStream = stream
		}
	}
	return &bestStream
}

// GetIndex 获取index.m3u8文件内容
func GetIndex(sessionID string) (string, error) {
	client := client.Get()

	resp, err := client.R().
		SetPathParam("sessionId", sessionID).
		Get("https://hls-auth.cloud.stream.co.jp/auth/index.m3u8?session_id={sessionId}")

	if err != nil {
		return "", err
	}

	return resp.String(), nil
}
