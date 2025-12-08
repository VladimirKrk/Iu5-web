package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

const (
	authorizationHeader = "Authorization"
	userCtx             = "user_id"
	roleCtx             = "is_moderator"
)
const INTERNAL_API_KEY = "Land_Cruiser_70_series_is_one_of_the_best_looking_cars_ever_lab8"

// AuthMiddleware проверяет JWT-токен и права доступа
func (h *Handler) AuthMiddleware(isModeratorOnly bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Извлекаем токен из заголовка
		tokenString := extractTokenFromHeader(c.Request)
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing or empty"})
			return
		}

		// 2. Парсим и валидируем токен
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			jwtSecret := os.Getenv("JWT_SECRET")
			if jwtSecret == "" {
				return nil, errors.New("JWT_SECRET environment variable is not set")
			}
			return []byte(jwtSecret), nil
		})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// 3. Проверяем, что токен не в черном списке (не "разлогинен")
		isBlacklisted, err := h.Repository.IsTokenBlacklisted(context.Background(), tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
		if isBlacklisted {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token has been invalidated"})
			return
		}

		// 4. Извлекаем данные (claims) из токена
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}

		// 5. Проверяем права доступа (роль модератора)
		isModerator, _ := claims[roleCtx].(bool)
		if isModeratorOnly && !isModerator {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Moderator access required"})
			return
		}

		// 6. Извлекаем ID пользователя и передаем его в контекст запроса
		userIDFloat, ok := claims[userCtx].(float64) // JWT парсит числа как float64
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID in token"})
			return
		}
		userID := uint(userIDFloat)

		c.Set(userCtx, userID)
		c.Set(roleCtx, isModerator)
		c.Next()
	}
}

// internal Auth for псевдо-авторизации
func (h *Handler) InternalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		//expectedApiKey := "Land_Cruiser_70_series_is_one_of_the_best_looking_cars_ever_lab8"
		expectedApiKey := INTERNAL_API_KEY
		apiKey := c.GetHeader("X-Internal-API-Key")

		// --- V-- ДОБАВЛЯЕМ ЛОГИ ДЛЯ ОТЛАДКИ --V ---
		fmt.Println("--- Internal Auth Middleware ---")
		fmt.Printf("Expected API Key from .env: '%s'\n", expectedApiKey)
		fmt.Printf("Received API Key from header: '%s'\n", apiKey)
		fmt.Println("-------------------------------")
		// --- ^-- КОНЕЦ БЛОКА --^ ---

		if expectedApiKey == "" {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal API key not configured on server"})
			return
		}

		if apiKey != expectedApiKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid internal API key"})
			return
		}
		c.Next()
	}
}

// extractTokenFromHeader извлекает токен из заголовка "Authorization"
func extractTokenFromHeader(r *http.Request) string {
	bearerToken := r.Header.Get(authorizationHeader)
	// Ожидаемый формат: "Bearer <token>"
	if len(strings.Split(bearerToken, " ")) == 2 {
		return strings.Split(bearerToken, " ")[1]
	}
	return ""
}

// --- Вспомогательные функции для извлечения данных из контекста ---

func GetUserID(c *gin.Context) (uint, error) {
	id, ok := c.Get(userCtx)
	if !ok {
		return 0, errors.New("user_id not found in context")
	}
	idUint, ok := id.(uint)
	if !ok {
		return 0, errors.New("user_id is of invalid type")
	}
	return idUint, nil
}

func GetUserRole(c *gin.Context) (bool, error) {
	isModerator, ok := c.Get(roleCtx)
	if !ok {
		return false, errors.New("is_moderator not found in context")
	}
	isModeratorBool, ok := isModerator.(bool)
	if !ok {
		return false, errors.New("is_moderator is of invalid type")
	}
	return isModeratorBool, nil
}

// GetTokenTTL извлекает оставшееся время жизни токена (нужна для Logout)
func GetTokenTTL(claims jwt.MapClaims) (time.Duration, error) {
	expVal, ok := claims["exp"]
	if !ok {
		return 0, errors.New("exp claim not present")
	}

	expFloat, ok := expVal.(float64)
	if !ok {
		return 0, fmt.Errorf("unsupported exp type %T", expVal)
	}
	expTime := time.Unix(int64(expFloat), 0)

	if time.Now().After(expTime) {
		return 0, errors.New("token already expired")
	}
	return time.Until(expTime), nil
}
