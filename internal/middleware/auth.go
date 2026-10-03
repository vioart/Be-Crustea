package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthRequired(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if jwtSecret == "" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "JWT secret is not configured",
			})
			c.Abort()
			return
		}

		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "authorization header is required",
			})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)

		if len(parts) != 2 ||
			!strings.EqualFold(parts[0], "Bearer") ||
			strings.TrimSpace(parts[1]) == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "invalid authorization header",
			})
			c.Abort()
			return
		}

		tokenString := strings.TrimSpace(parts[1])

		token, err := jwt.Parse(
			tokenString,
			func(token *jwt.Token) (interface{}, error) {
				// Pastikan token menggunakan algoritma yang kita harapkan.
				if token.Method != jwt.SigningMethodHS256 {
					return nil, jwt.ErrSignatureInvalid
				}

				return []byte(jwtSecret), nil
			},
		)

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "invalid or expired token",
			})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "invalid token claims",
			})
			c.Abort()
			return
		}

		// sub dibuat sebagai int64 di AuthService.
		// Saat JWT diparse kembali, angka JSON biasanya menjadi float64.
		sub, ok := claims["sub"].(float64)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "invalid user id in token",
			})
			c.Abort()
			return
		}

		userID := int64(sub)

		if userID <= 0 {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "invalid user id in token",
			})
			c.Abort()
			return
		}

		c.Set("user_id", userID)

		if email, ok := claims["email"].(string); ok {
			c.Set("user_email", email)
		}

		if name, ok := claims["name"].(string); ok {
			c.Set("user_name", name)
		}

		c.Next()
	}
}

func GetUserID(c *gin.Context) (int64, bool) {
	value, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}

	userID, ok := value.(int64)
	if !ok {
		return 0, false
	}

	return userID, true
}

func GetUserEmail(c *gin.Context) (string, bool) {
	value, exists := c.Get("user_email")
	if !exists {
		return "", false
	}

	email, ok := value.(string)
	return email, ok
}

func GetUserName(c *gin.Context) (string, bool) {
	value, exists := c.Get("user_name")
	if !exists {
		return "", false
	}

	name, ok := value.(string)
	return name, ok
}

func UserIDToString(c *gin.Context) string {
	userID, ok := GetUserID(c)
	if !ok {
		return ""
	}

	return strconv.FormatInt(userID, 10)
}
