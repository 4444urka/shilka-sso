package jwt

import (
	"github.com/golang-jwt/jwt/v5"
	"shilka-sso/internal/domain/models"
	"time"
)

// NewToken Создаёт  JWT токен, который хранит в себе инофрмацию о пользователе
func NewToken(user models.User, app models.App, duration time.Duration) (string, error) {

	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = user.Id
	claims["username"] = user.Username
	claims["exp"] = time.Now().Add(duration).Unix()
	claims["app_id"] = app.Id

	tokenString, err := token.SignedString([]byte(app.Secret))

	if err != nil {
		return "", err
	}

	return tokenString, nil
}
