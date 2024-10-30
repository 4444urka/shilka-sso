package tests

import (
	ssov1 "github.com/4444urka/shilka-protos/gen/go/sso"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"shilka-sso/tests/suite"
	"sync"
	"testing"
	"time"
)

// Делает 100 синхронных запросов о регистрации и логине на сервер
func TestHugeAmountOfRegisterLoginResponses(t *testing.T) {
	ctx, st := suite.New(t)

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

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
		}()
	}

	wg.Wait()
}
