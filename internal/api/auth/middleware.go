package auth

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func ExtractAuthDataMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authToken, err := c.Cookie(AuthTokenCookieName)
		if err != nil {
			c.Next()
			return
		}

		token, err := jwt.ParseWithClaims(authToken, &AuthData{}, func(token *jwt.Token) (interface{}, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.Next()
			return
		}

		claims, ok := token.Claims.(*AuthData)
		if !ok {
			c.Next()
			return
		}

		c.Set("customClaims", claims)
		c.Next()
	}
}

func GetAuthData(c *gin.Context) (*AuthData, bool) {
	v, ok := c.Get("customClaims")
	if !ok {
		return nil, false
	}

	claims, ok := v.(*AuthData)
	if !ok {
		return nil, false
	}

	return claims, true
}
