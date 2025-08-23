package video

import (
	"fmt"
	"strconv"

	"github.com/go-resty/resty/v2"
)

type VideoDetailsResponse struct {
	Data struct {
		VideoPage *VideoDetails `json:"video_page"`
	} `json:"data"`
}

type VideoDetails struct {
	ActiveVideoFilename  *ActiveVideoFilename `json:"active_video_filename"`
	ContentCode          string               `json:"content_code"`
	Description          string               `json:"description"`
	DisplayDate          string               `json:"display_date"`
	LiveFinishedAt       *string              `json:"live_finished_at"`
	LiveScheduledEndAt   *string              `json:"live_scheduled_end_at"`
	LiveScheduledStartAt *string              `json:"live_scheduled_start_at"`
	LiveStartedAt        *string              `json:"live_started_at"`
	ReleasedAt           string               `json:"released_at"`
	StartWithFreePartFlg bool                 `json:"start_with_free_part_flg"`
	ThumbnailURL         string               `json:"thumbnail_url"`
	Title                string               `json:"title"`
	VideoAggregateInfo   *VideoAggregateInfo  `json:"video_aggregate_info"`
	VideoCommentSetting  *VideoCommentSetting `json:"video_comment_setting"`
	VideoFreePeriods     []VideoFreePeriod    `json:"video_free_periods"`
	VideoQuestionnaires  []VideoQuestionnaire `json:"video_questionnaires"`
	VideoStream          *VideoStream         `json:"video_stream"`
}

type ActiveVideoFilename struct {
	ID                int                `json:"id"`
	Length            int                `json:"length"`
	VideoFilenameType *VideoFilenameType `json:"video_filename_type"`
}

type VideoFilenameType struct {
	ID    int    `json:"id"`
	Value string `json:"value"`
}

type VideoAggregateInfo struct {
	ID               int `json:"id"`
	NumberOfComments int `json:"number_of_comments"`
	TotalViews       int `json:"total_views"`
}

type VideoCommentSetting struct {
	CommentGroupID string `json:"comment_group_id"`
}

type VideoFreePeriod struct {
	ElapsedEndedTime   int    `json:"elapsed_ended_time"`
	ElapsedStartedTime int    `json:"elapsed_started_time"`
	EndAt              string `json:"end_at"`
	ID                 int    `json:"id"`
	StartedAt          string `json:"started_at"`
}

type VideoQuestionnaire struct {
	ElapsedDeadlineTime       int                        `json:"elapsed_deadline_time"`
	ElapsedHideResultTime     int                        `json:"elapsed_hide_result_time"`
	ElapsedResultTime         int                        `json:"elapsed_result_time"`
	ElapsedShowTime           int                        `json:"elapsed_show_time"`
	ID                        int                        `json:"id"`
	Question                  string                     `json:"question"`
	VideoQuestionnaireOptions []VideoQuestionnaireOption `json:"video_questionnaire_options"`
}

type VideoQuestionnaireOption struct {
	ID                       int                      `json:"id"`
	Text                     string                   `json:"text"`
	VideoQuestionnaireResult VideoQuestionnaireResult `json:"video_questionnaire_result"`
}

type VideoQuestionnaireResult struct {
	Percentage int `json:"percentage"`
}

type VideoStream struct {
	AuthenticatedURL string `json:"authenticated_url"`
}

func GetVideoDetails(fcSiteID int, contentCode string) (*VideoDetails, error) {
	client := resty.New()

	var response VideoDetailsResponse

	baseURL := "https://api.nicochannel.jp/fc/video_pages/%s"
	url := fmt.Sprintf(baseURL, contentCode)

	resp, err := client.R().
		SetHeader("fc_site_id", strconv.Itoa(fcSiteID)).
		SetHeader("fc_use_device", "null").
		SetResult(&response).
		Get(url)

	if err != nil {
		return nil, fmt.Errorf("GetVideoDetails: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("GetVideoDetails: 状态码 %d", resp.StatusCode())
	}

	return response.Data.VideoPage, nil
}
