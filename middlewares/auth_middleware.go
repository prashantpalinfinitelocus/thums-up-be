package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Infinite-Locus-Product/thums_up_backend/config"
	"github.com/Infinite-Locus-Product/thums_up_backend/errors"
	"github.com/Infinite-Locus-Product/thums_up_backend/repository"
)

type JWTClaims struct {
	UserID string `json:"user_id"`
	Phone  string `json:"phone"`
	jwt.RegisteredClaims
}

func AuthMiddleware(db *gorm.DB, userRepo repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string
		cfg := config.GetConfig()

		// First, try to get token from "access_token" cookie (preferred)
		cookieToken, err := c.Cookie("access_token")
		if err == nil && cookieToken != "" {
			tokenString = cookieToken
		} else {
			// If "access_token" cookie not found, check all cookies for a valid JWT token
			cookies := c.Request.Cookies()
			for _, cookie := range cookies {
				if cookie.Value != "" {
					// Try to parse as JWT to see if it's a valid token
					token, err := jwt.ParseWithClaims(cookie.Value, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
						return []byte(cfg.JwtConfig.SecretKey), nil
					})
					if err == nil && token.Valid {
						tokenString = cookie.Value
						break
					}
				}
			}
		}

		// If no token found in cookies, check Authorization header
		if tokenString == "" {
			authHeader := c.GetHeader("Authorization")
			if authHeader == "" {
				c.JSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"error":   errors.ErrAuthHeaderRequired,
				})
				c.Abort()
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				c.JSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"error":   errors.ErrInvalidAuthHeaderFormat,
				})
				c.Abort()
				return
			}

			tokenString = parts[1]
		}

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   errors.ErrAuthHeaderRequired,
			})
			c.Abort()
			return
		}

		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JwtConfig.SecretKey), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   errors.ErrInvalidOrExpiredToken,
			})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(*JWTClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   errors.ErrInvalidTokenClaims,
			})
			c.Abort()
			return
		}

		userUUID, err := uuid.Parse(claims.UserID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   errors.ErrInvalidUserID,
			})
			c.Abort()
			return
		}

		user, err := userRepo.FindById(c.Request.Context(), db, userUUID)
		if err != nil || user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   errors.ErrUserNotFound,
			})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("phone", claims.Phone)
		c.Set("user", user)
		c.Next()
	}
}

func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string
		cfg := config.GetConfig()

		// First, try to get token from "access_token" cookie (preferred)
		cookieToken, err := c.Cookie("access_token")
		if err == nil && cookieToken != "" {
			tokenString = cookieToken
		} else {
			// If "access_token" cookie not found, check all cookies for a valid JWT token
			cookies := c.Request.Cookies()
			for _, cookie := range cookies {
				if cookie.Value != "" {
					// Try to parse as JWT to see if it's a valid token
					token, err := jwt.ParseWithClaims(cookie.Value, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
						return []byte(cfg.JwtConfig.SecretKey), nil
					})
					if err == nil && token.Valid {
						tokenString = cookie.Value
						break
					}
				}
			}
		}

		// If no token found in cookies, check Authorization header
		if tokenString == "" {
			authHeader := c.GetHeader("Authorization")
			if authHeader == "" {
				c.Next()
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				c.Next()
				return
			}

			tokenString = parts[1]
		}

		if tokenString == "" {
			c.Next()
			return
		}

		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JwtConfig.SecretKey), nil
		})

		if err == nil && token.Valid {
			if claims, ok := token.Claims.(*JWTClaims); ok {
				c.Set("user_id", claims.UserID)
				c.Set("phone", claims.Phone)
			}
		}

		c.Next()
	}
}
