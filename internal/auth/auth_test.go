package auth

import (
	"log"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	_, err := GetToken()
	if err != nil {
		log.Fatalf("获取 token 失败: %v", err)
	}
	code := m.Run()
	os.Exit(code)
}

// go test ./internal/auth -v
func TestGetToken(t *testing.T) {
	token1 := getToken(t)
	t.Logf("获取到的 token: %s", token1[:50])

	token2 := getToken(t)
	t.Logf("获取到的 token: %s", token2[:50])

	if token1 != token2 {
		t.Fatalf("两次获取的 token 不同")
	} else {
		t.Logf("两次获取的 token 相同")
	}
}

func getToken(t *testing.T) string {
	token, err := GetToken()
	if err != nil {
		t.Fatalf("获取 token 失败: %v", err)
	}
	return token
}

func TestGetExpireTime(t *testing.T) {
	expireTime := GetExpireTime()
	t.Logf("获取到的 expire_time: %s", expireTime.Format("2006-01-02 15:04:05"))
}

func TestGetRefreshToken(t *testing.T) {
	refreshToken := GetRefreshToken()
	t.Logf("获取到的 refresh_token: %s", refreshToken)
}
