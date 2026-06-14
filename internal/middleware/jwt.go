package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTAuth JWT 认证中间件
func JWTAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少 Authorization header"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header 格式错误"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		}, jwt.WithIssuer("studyforge-pro"), jwt.WithAudience("studyforge-client"))

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token 无效或已过期"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token claims 解析失败"})
			c.Abort()
			return
		}

		userID, _ := claims["user_id"].(string)
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token 缺少 user_id"})
			c.Abort()
			return
		}
		username, _ := claims["username"].(string)
		c.Set("userID", userID)
		c.Set("username", username)

		c.Next()
	}
}

// GenerateToken 生成 JWT Token（含 iss/aud 防跨服务重放）
func GenerateToken(userID, username, secret string, expireHours int) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"exp":      time.Now().Add(time.Duration(expireHours) * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
		"iss":      "studyforge-pro",
		"aud":      "studyforge-client",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateToken 验证 JWT Token，返回 userID（用于 WebSocket 鉴权等非 gin 场景）
// 与 JWTAuth 保持一致的 iss/aud 校验，防止跨服务 Token 重放
func ValidateToken(tokenString, secret string) (userID string, err error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secret), nil
	}, jwt.WithIssuer("studyforge-pro"), jwt.WithAudience("studyforge-client"))
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", jwt.ErrSignatureInvalid
	}

	userID, _ = claims["user_id"].(string)
	if userID == "" {
		return "", fmt.Errorf("token 缺少 user_id")
	}
	return userID, nil
}
