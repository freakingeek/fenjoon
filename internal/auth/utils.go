package auth

import (
	"errors"

	"github.com/gin-gonic/gin"
)

func GetUserIdFromContext(c *gin.Context) (uint, error) {
	token, err := ParseBearerToken(c.GetHeader("Authorization"))
	if err != nil {
		return 0, err
	}

	claims, err := ParseJWTToken(token)
	if err != nil {
		return 0, err
	}

	floatUserId, ok := claims["id"].(float64)
	if !ok {
		return 0, errors.New("invalid float number")
	}

	return uint(floatUserId), nil
}
