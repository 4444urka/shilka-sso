package tests

import (
	ssov1 "github.com/4444urka/shilka-protos/gen/go/sso"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"shilka-sso/tests/suite"
	"testing"
	"time"
)

const (
	emptyAppId     = 0
	appID          = 1
	appSecret      = "4urka"
	passDefaultLen = 10
)

// Делает запрос регистрации и логина на сервер
func TestRegisterLogin_Login_HappyPath(t *testing.T) {
	ctx, st := suite.New(t)

	username := gofakeit.Username()
	password := randomFakePassword()

	registerResponse, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Username: username,
		Password: password,
	})

	require.NoError(t, err)

	assert.NotEmpty(t, registerResponse.GetUserId())

	loginResponse, err := st.AuthClient.Login(ctx, &ssov1.LoginRequest{
		Username: username,
		Password: password,
		AppId:    appID,
	})

	require.NoError(t, err)

	loginTime := time.Now()

	token := loginResponse.GetToken()
	require.NotEmpty(t, token)

	tokenParsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(appSecret), nil
	})
	require.NoError(t, err)

	claims, ok := tokenParsed.Claims.(jwt.MapClaims)
	require.True(t, ok)

	assert.Equal(t, registerResponse.GetUserId(), int64(claims["user_id"].(float64)))
	assert.Equal(t, username, claims["username"].(string))
	assert.Equal(t, appID, int(claims["app_id"].(float64)))

	const deltaSeconds = 3
	assert.InDelta(t, loginTime.Add(st.Cfg.TokenTTL).Unix(), claims["exp"].(float64), deltaSeconds)

}

// Делает 2 запроса регистрации на сервер, должен получить ошибку
func TestRegisterLogin_DuplicatedRegistration(t *testing.T) {
	ctx, st := suite.New(t)

	username := gofakeit.Username()
	password := randomFakePassword()

	registerResponse, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Username: username,
		Password: password,
	})

	require.NoError(t, err)
	require.NotEmpty(t, registerResponse.GetUserId())

	registerResponse, err = st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Username: username,
		Password: password,
	})
	require.Error(t, err)
	assert.Empty(t, registerResponse.GetUserId())
	assert.ErrorContains(t, err, "user already exists")
}

// Отправляет запрос регистрации, а затем запрос логина на тот же username с
// неправильным паролем. Должен получить ошибку invalidCredentials.
func TestRegisterLogin_InvalidPassword(t *testing.T) {
	ctx, st := suite.New(t)

	username := gofakeit.Username()
	password := randomFakePassword()

	registerResponse, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Username: username,
		Password: password,
	})

	require.NoError(t, err)
	require.NotEmpty(t, registerResponse.GetUserId())

	loginResponse, err := st.AuthClient.Login(ctx, &ssov1.LoginRequest{
		Username: username,
		Password: randomFakePassword(),
		AppId:    appID,
	})

	require.Error(t, err)
	assert.Empty(t, loginResponse.GetToken())
	assert.ErrorContains(t, err, "invalid credentials")
}

// Делает 100 синхронных запросов о регистрации и логине на сервер
//func TestHugeAmountOfRegisterLoginResponses(t *testing.T) {
//	ctx, st := suite.New(t)
//
//	for i := 0; i < 100; i++ {
//		username := gofakeit.Username()
//		password := randomFakePassword()
//		registerResponse, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
//			Username: username,
//			Password: password,
//		})
//
//		require.NoError(t, err)
//		require.NotEmpty(t, registerResponse.GetUserId())
//
//		loginResponse, err := st.AuthClient.Login(ctx, &ssov1.LoginRequest{
//			Username: username,
//			Password: password,
//			AppId:    appID,
//		})
//
//		require.NoError(t, err)
//
//		loginTime := time.Now()
//
//		token := loginResponse.GetToken()
//		require.NotEmpty(t, token)
//
//		tokenParsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
//			return []byte(appSecret), nil
//		})
//		require.NoError(t, err)
//
//		claims, ok := tokenParsed.Claims.(jwt.MapClaims)
//		require.True(t, ok)
//
//		assert.Equal(t, registerResponse.GetUserId(), int64(claims["user_id"].(float64)))
//		assert.Equal(t, username, claims["username"].(string))
//		assert.Equal(t, appID, int(claims["app_id"].(float64)))
//
//		const deltaSeconds = 3
//		assert.InDelta(t, loginTime.Add(st.Cfg.TokenTTL).Unix(), claims["exp"].(float64), deltaSeconds)
//	}
//}

// Создаёт случайный пароль с цифрами и буквами специальными знаками длиной до 10 символов
func randomFakePassword() string {
	return gofakeit.Password(true, true, true, true, false, passDefaultLen)
}
