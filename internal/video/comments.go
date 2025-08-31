package video

import (
	"sort"
	"strconv"
	"time"

	"ncpd/internal/client"
)

type CommentsUserTokenResponse struct {
	Data struct {
		AccessToken string `json:"access_token"`
	} `json:"data"`
}

type Message struct {
	CreatedAt        time.Time `json:"created_at"`          // 创建时间
	EndTimeInSeconds *int      `json:"end_time_in_seconds"` // 置顶时长，可能为空
	GroupID          string    `json:"group_id"`            // 群组 ID
	ID               string    `json:"id"`                  // 消息 ID
	Mentions         []any     `json:"mentions"`            // 基本为空数组，具体结构不明
	Message          string    `json:"message"`             // 消息内容
	Nickname         string    `json:"nickname"`            // 昵称
	PlaybackTime     int       `json:"playback_time"`       // 弹幕出现的时间点，单位秒
	Priority         bool      `json:"priority"`            // 是否置顶
	SenderID         string    `json:"sender_id"`           // 发送者 ID
	SentAt           time.Time `json:"sent_at"`             // 发送时间
	UpdatedAt        time.Time `json:"updated_at"`          // 更新时间
}

func GetCommentsUserToken(videoID string) (string, error) {
	client := client.Get()

	var commentsUserTokenResponse CommentsUserTokenResponse
	_, err := client.R().
		SetPathParam("videoId", videoID).
		SetResult(&commentsUserTokenResponse).
		Get("/video_pages/{videoId}/comments_user_token")

	if err != nil {
		return "", err
	}

	return commentsUserTokenResponse.Data.AccessToken, nil
}

func GetComments(commentsUserToken string, groupID string, startTime int) ([]Message, error) {
	client := client.Get()

	limit := 120

	var commentsResponse []Message
	_, err := client.R().
		SetHeader("content-type", "application/json").
		SetPathParam("startTime", strconv.Itoa(startTime)).
		SetPathParam("limit", strconv.Itoa(limit)).
		SetBody(map[string]string{
			"token":    commentsUserToken,
			"group_id": groupID,
		}).
		SetResult(&commentsResponse).
		Post("https://comm-api.sheeta.com/messages.history?oldest_playback_time={startTime}&sort_direction=asc&limit={limit}&inclusive=true")

	if err != nil {
		return nil, err
	}

	return commentsResponse, nil
}

func GetAllComments(commentsUserToken string, groupId string) ([]Message, error) {
	msgSet := make(map[string]Message)
	startTime := 0
	count := 0

	for {
		// 获取一批评论
		comments, err := GetComments(commentsUserToken, groupId, startTime)
		if err != nil {
			return nil, err
		}
		count += len(comments)

		// 整个视频一条评论都没有
		if len(comments) == 0 {
			break
		}

		newCount := 0
		// 将评论添加到 set 中进行去重
		for _, comment := range comments {
			if _, ok := msgSet[comment.ID]; !ok {
				newCount++
			}
			msgSet[comment.ID] = comment
		}

		// fmt.Printf("获取到 %d 条弹幕，实际新增 %d 条，累计 %d 条\n", len(comments), newCount, count)

		// 获取最后一条评论的 playback_time 作为下一次请求的 startTime
		lastComment := comments[len(comments)-1]
		nextStartTime := lastComment.PlaybackTime

		// 因为 inclusive=true 所以要防止反复获取相同的数据
		if nextStartTime == startTime && newCount == 0 {
			break
		}

		startTime = nextStartTime
	}

	// 将 map 转换为 slice
	var allComments []Message
	for _, comment := range msgSet {
		allComments = append(allComments, comment)
	}

	// 按 PlaybackTime 排序
	sort.Slice(allComments, func(i, j int) bool {
		return allComments[i].PlaybackTime < allComments[j].PlaybackTime
	})

	// fmt.Printf("获取到 %d 条弹幕\n", count)
	// fmt.Printf("去重后 %d 条弹幕\n", len(allComments))

	return allComments, nil
}
