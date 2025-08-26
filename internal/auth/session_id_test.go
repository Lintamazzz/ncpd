package auth

import (
	"testing"
)

func TestGetSessionID(t *testing.T) {
	token, err := GetToken()
	if err != nil {
		t.Fatalf("获取 token 失败: %v", err)
	}

	videoID := "smQKzZSkFT4Fap6ERziVr26f"
	sessionID, err := GetSessionID(videoID, token)
	if err != nil {
		t.Fatalf("获取 sessionID 失败: %v", err)
	}

	if sessionID == "" {
		t.Error("获取到的 sessionID 不应为空")
	}

	t.Logf("成功获取到 sessionID: %v", sessionID)
}
