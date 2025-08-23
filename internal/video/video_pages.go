package video

import (
	"fmt"

	"github.com/go-resty/resty/v2"
)

type VideoPagesResponse struct {
	Data struct {
		VideoPages struct {
			List  []VideoDetails `json:"list"`
			Total int            `json:"total"`
		} `json:"video_pages"`
	} `json:"data"`
}

func GetVideoList(fcSiteID int) ([]VideoDetails, error) {
	client := resty.New()

	// 这个地址返回的视频信息不全，获取更详细的信息需要使用 GetVideoDetails
	var allVideos []VideoDetails
	page := 1
	perPage := 10
	baseURL := "https://api.nicochannel.jp/fc/v2/fanclub_sites/%d/video_pages?sort=display_date&vod_type=0&per_page=%d&page=%d"

	fmt.Printf("开始获取 频道ID: %d 的所有视频信息...\n", fcSiteID)

	for {
		var response VideoPagesResponse

		url := fmt.Sprintf(baseURL, fcSiteID, perPage, page)

		resp, err := client.R().
			SetHeader("fc_use_device", "null").
			SetResult(&response).
			Get(url)

		if err != nil {
			return nil, fmt.Errorf("GetVideoList: 请求第 %d 页失败 %w", page, err)
		}

		// 打印响应状态码
		fmt.Printf("第 %d 页响应状态码: %d\n", page, resp.StatusCode())

		// 检查是否有数据
		if len(response.Data.VideoPages.List) == 0 {
			fmt.Printf("第 %d 页没有数据，停止获取\n\n", page)
			break
		}

		// 将当前页的数据添加到总列表中
		allVideos = append(allVideos, response.Data.VideoPages.List...)

		fmt.Printf("第 %d 页获取到 %d 个视频，累计 %d 个视频\n",
			page, len(response.Data.VideoPages.List), len(allVideos))

		// 如果当前页的数据少于每页数量，说明已经是最后一页
		if len(response.Data.VideoPages.List) < perPage {
			fmt.Printf("第 %d 页数据少于 %d 个，已到达最后一页\n", page, perPage)
			break
		}

		page++
	}

	return allVideos, nil
}
