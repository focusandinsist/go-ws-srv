package auth

import (
	"errors"
	"fmt"
	"time"
)

// Auth 用于存储认证信息（如 JWT Secret）
type Auth struct {
	// 可以存储密钥、过期时间等
	SecretKey string
}

// NewAuth 创建一个新的认证实例
func NewAuth() *Auth {
	return &Auth{
		SecretKey: "mysecretkey", // 示例密钥，实际使用时需要更加安全的做法
	}
}

// ValidateToken 验证 token 是否有效
// 这里只是示例，实际应该解密并验证 token
func (a *Auth) ValidateToken(token string) (bool, error) {
	// 在实际应用中，你可以解密 token 或与数据库对比
	if token == "" {
		return false, errors.New("token is empty")
	}

	// 假设 token 有效期为 24 小时
	expirationTime := time.Now().Add(24 * time.Hour)
	if time.Now().After(expirationTime) {
		return false, fmt.Errorf("token expired")
	}

	// 假设 token 校验成功
	return true, nil
}
