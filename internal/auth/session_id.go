package auth

import (
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
)

type SessionIDResponse struct {
	Data struct {
		SessionID string `json:"session_id"`
	} `json:"data"`
}

func GetSessionID(videoID string, token string) (string, error) {
	client := resty.New()

	var sessionIDResponse SessionIDResponse

	baseURL := "https://api.nicochannel.jp/fc/video_pages/%s/session_ids"
	URL := fmt.Sprintf(baseURL, videoID)

	resp, err := client.R().
		SetHeader("fc_use_device", "null").
		SetHeader("Origin", "https://nicochannel.jp").
		SetAuthToken(token).
		SetBody(map[string]string{}).
		SetResult(&sessionIDResponse).
		Post(URL)

	if err != nil {
		return "", fmt.Errorf("GetSessionID: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		if resp.StatusCode() == http.StatusForbidden {
			return "", fmt.Errorf("状态码 %d - 会员限定内容", resp.StatusCode())
		}
		return "", fmt.Errorf("状态码 %d", resp.StatusCode())
	}

	return sessionIDResponse.Data.SessionID, nil
}
