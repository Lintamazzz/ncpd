package auth

import (
	"fmt"
	"log"
	"sync"
	"time"

	"ncpd/config"
	"ncpd/internal/client"

	"github.com/go-resty/resty/v2"
)

type OAuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
	Scope        string `json:"scope"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// tokenManager 管理 OAuth token 的获取和自动刷新
type tokenManager struct {
	mu           sync.RWMutex
	accessToken  string
	refreshToken string
	expiresAt    time.Time
	client       *resty.Client
	config       *config.Config
}

var (
	instance *tokenManager
	once     sync.Once
)

// Singleton
func getTokenManager() *tokenManager {
	once.Do(func() {
		instance = &tokenManager{
			client: client.Get(),
			config: config.Load(),
		}
	})
	return instance
}

// 获取有效的 Access token，如果过期会自动刷新
func (tm *tokenManager) getToken() (string, error) {
	// 检查 token 是否有效，使用读锁
	tm.mu.RLock()
	if tm.isTokenValid() {
		token := tm.accessToken
		tm.mu.RUnlock()
		return token, nil
	}
	tm.mu.RUnlock()

	// 需要刷新 token，使用写锁
	return tm.refreshOAuthToken()
}

// 检查 token 是否有效
func (tm *tokenManager) isTokenValid() bool {
	// 提前 30s 刷新 token
	return tm.accessToken != "" && time.Now().Add(30*time.Second).Before(tm.expiresAt)
}

// 刷新 token
func (tm *tokenManager) refreshOAuthToken() (string, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// 双重检查，防止并发刷新
	if tm.isTokenValid() {
		return tm.accessToken, nil
	}

	log.Println("正在刷新 OAuth token...")

	var oauthResp OAuthResponse

	// 获取 refresh token，优先使用存储的，否则使用配置中的
	var refreshToken string
	if tm.refreshToken != "" {
		refreshToken = tm.refreshToken
	} else {
		refreshToken = tm.config.NicoRefreshToken
	}

	_, err := tm.client.R().
		SetFormData(map[string]string{
			"client_id":     tm.config.NicoClientID,
			"redirect_uri":  fmt.Sprintf("https://%s/login/login-redirect", client.CurrentPlatform.Domain),
			"grant_type":    "refresh_token",
			"refresh_token": refreshToken,
		}).
		SetResult(&oauthResp).
		SetPathParam("platformDomain", client.CurrentPlatform.Domain).
		Post("https://auth.{platformDomain}/oauth/token")

	if err != nil {
		return "", err
	}

	// 保存新的 token 信息
	tm.accessToken = oauthResp.AccessToken
	tm.refreshToken = oauthResp.RefreshToken
	tm.expiresAt = time.Now().Add(time.Duration(oauthResp.ExpiresIn) * time.Second)

	log.Printf("刷新 token 成功，过期时间: %s", tm.expiresAt.Format("2006-01-02 15:04:05"))
	log.Printf("新的 refresh_token: %s", tm.refreshToken)

	return tm.accessToken, nil
}

// 获取当前的 refresh token
func (tm *tokenManager) getRefreshToken() string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	// 先从内存中取，没有再从配置中取
	if tm.refreshToken != "" {
		return tm.refreshToken
	} else {
		return tm.config.NicoRefreshToken
	}
}

// 获取 token 过期时间
func (tm *tokenManager) getTokenExpiry() time.Time {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.expiresAt
}

// 提供全局函数供其他文件使用
// 获取 token，如果过期会自动刷新
func GetToken() (string, error) {
	return getTokenManager().getToken()
}

// 获取 token 过期时间
func GetExpireTime() time.Time {
	return getTokenManager().getTokenExpiry()
}

// 获取 refresh token
func GetRefreshToken() string {
	return getTokenManager().getRefreshToken()
}
