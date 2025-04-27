package auth

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// AuthManager 用于管理 token 和 session 状态
type AuthManager struct {
	activeTokens map[string]*Session // 存储用户的活跃 Token
	mu           sync.Mutex          // 用于保护并发操作
}

// Session 用于表示一个用户的认证 session
type Session struct {
	UserID    string    // 用户 ID
	Token     string    // 认证 Token
	CreatedAt time.Time // Session 创建时间
	ExpiresAt time.Time // Session 过期时间
}

// NewAuthManager 创建一个新的 AuthManager
func NewAuthManager() *AuthManager {
	return &AuthManager{
		activeTokens: make(map[string]*Session),
	}
}

// CreateSession 创建一个新的认证 Session
func (am *AuthManager) CreateSession(userID, token string) (*Session, error) {
	am.mu.Lock()
	defer am.mu.Unlock()

	// 检查 token 是否已存在
	if _, exists := am.activeTokens[token]; exists {
		return nil, fmt.Errorf("token already exists")
	}

	// 创建新的 session
	session := &Session{
		UserID:    userID,
		Token:     token,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour), // 假设 token 有效期为 24 小时
	}

	// 保存 session
	am.activeTokens[token] = session
	log.Printf("New session created for user %s with token %s", userID, token)
	return session, nil
}

// ValidateSession 验证用户的 token 是否有效
func (am *AuthManager) ValidateSession(token string) (bool, *Session) {
	am.mu.Lock()
	defer am.mu.Unlock()

	// 查找 session
	session, exists := am.activeTokens[token]
	if !exists {
		return false, nil
	}

	// 检查 session 是否过期
	if time.Now().After(session.ExpiresAt) {
		delete(am.activeTokens, token) // 删除过期 session
		return false, nil
	}

	return true, session
}

// RefreshSession 刷新 session 的过期时间
func (am *AuthManager) RefreshSession(token string) (*Session, error) {
	am.mu.Lock()
	defer am.mu.Unlock()

	// 查找 session
	session, exists := am.activeTokens[token]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	// 刷新过期时间
	session.ExpiresAt = time.Now().Add(24 * time.Hour) // 重新设置有效期

	log.Printf("Session for user %s refreshed", session.UserID)
	return session, nil
}

// RemoveSession 移除认证 session
func (am *AuthManager) RemoveSession(token string) {
	am.mu.Lock()
	defer am.mu.Unlock()

	delete(am.activeTokens, token)
	log.Printf("Session with token %s removed", token)
}
