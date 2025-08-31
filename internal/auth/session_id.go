package auth

import (
	"errors"
	"fmt"
	"ncpd/internal/client"
	"net/http"
)

type SessionIDResponse struct {
	Data struct {
		SessionID string `json:"session_id"`
	} `json:"data"`
}

func GetSessionID(videoID string, token string) (string, error) {
	c := client.Get()

	var sessionIDResponse SessionIDResponse

	_, err := c.R().
		SetHeader("fc_use_device", "null").
		SetHeader("Origin", fmt.Sprintf("https://%s", client.CurrentPlatform.Domain)).
		SetAuthToken(token).
		SetPathParam("videoId", videoID).
		SetBody(map[string]string{}).
		SetResult(&sessionIDResponse).
		Post("/video_pages/{videoId}/session_ids")

	if err != nil {
		var httpErr *client.HTTPError
		if errors.As(err, &httpErr) {
			if httpErr.StatusCode == http.StatusForbidden {
				return "", fmt.Errorf("状态码 %d - 会员限定内容", httpErr.StatusCode)
			}
		}
		return "", err
	}

	return sessionIDResponse.Data.SessionID, nil
}
