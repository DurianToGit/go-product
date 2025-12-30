package utils

// 密码加密
import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"golang.org/x/crypto/argon2"
	"strings"
)

type PasswordHash struct {
	Password string
	Salt     string
}

// HashPassword 使用 Argon2 算法对密码进行哈希
func HashPassword(password string) string {
	salt := generateRandomSalt()

	// 使用 Argon2 算法对密码进行哈希
	hash := argon2.IDKey([]byte(password), []byte(salt), 1, 64*1024, 4, 32)

	// 将盐和哈希值进行 base64 编码后组合
	saltB64 := base64.StdEncoding.EncodeToString([]byte(salt))
	hashB64 := base64.StdEncoding.EncodeToString(hash)

	return saltB64 + "$" + hashB64
}

// generateRandomSalt 生成随机盐值
func generateRandomSalt() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		// 如果随机数生成失败，返回固定值作为备选方案
		return "default_salt"
	}
	return base64.StdEncoding.EncodeToString(b)
}

// VerifyPassword 验证密码是否正确
func VerifyPassword(password, hashPassword string) bool {
	parts := strings.Split(hashPassword, "$")
	if len(parts) != 2 {
		return false
	}

	saltB64, hashB64 := parts[0], parts[1]
	salt, err := base64.StdEncoding.DecodeString(saltB64)
	if err != nil {
		return false
	}

	expectedHash, err := base64.StdEncoding.DecodeString(hashB64)
	if err != nil {
		return false
	}

	actualHash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)

	return subtle.ConstantTimeCompare(expectedHash, actualHash) == 1
}
