package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/misakimei123/redbook/internal/integration/startup"
	"github.com/misakimei123/redbook/internal/web/result"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func TestUserHandler_SendSMSCode(t *testing.T) {
	const (
		phone   = "134123456789"
		setCode = "123456"
	)
	key := fmt.Sprintf("phone_code:login:%s", phone)
	redis := startup.InitRedis()
	server := startup.InitWebServer()
	testCases := []struct {
		name     string
		before   func(t *testing.T)
		after    func(t *testing.T)
		phone    string
		wantCode int
		wantBody result.Result
	}{
		{
			name: "send sms success",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {
				ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*5)
				defer cancelFunc()
				code, err := redis.Get(ctx, key).Result()
				assert.NoError(t, err)
				assert.True(t, len(code) == 6)
				dur, err := redis.TTL(ctx, key).Result()
				assert.NoError(t, err)
				assert.True(t, dur > time.Minute*9+time.Second+50)
				err = redis.Del(ctx, key).Err()
				assert.NoError(t, err)
			},
			phone:    phone,
			wantCode: http.StatusOK,
			wantBody: result.RetSuccess,
		},
		{
			name: "no phone number",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {

			},
			wantCode: http.StatusOK,
			wantBody: result.RetNeedPhoneNumber,
		},
		{
			name: "send sms too frequent",
			before: func(t *testing.T) {
				ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*5)
				defer cancelFunc()
				err := redis.Set(ctx, key, setCode, time.Minute*9+time.Second*50).Err()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*5)
				defer cancelFunc()
				code, err := redis.Get(ctx, key).Result()
				assert.NoError(t, err)
				assert.Equal(t, setCode, code)
				err = redis.Del(ctx, key).Err()
				assert.NoError(t, err)
			},
			phone:    phone,
			wantCode: http.StatusOK,
			wantBody: result.RetTooFrequent,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)
			request, err := http.NewRequest(http.MethodPost, "/users/loginsms/code/send",
				bytes.NewReader([]byte(fmt.Sprintf(`{"phone":"%s"}`, tc.phone))))
			request.Header.Set("Content-Type", "application/json")
			assert.NoError(t, err)
			recorder := httptest.NewRecorder()
			server.ServeHTTP(recorder, request)
			assert.Equal(t, tc.wantCode, recorder.Code)
			var res result.Result
			err = json.NewDecoder(recorder.Body).Decode(&res)
			assert.NoError(t, err)
			//json.Unmarshal([]byte(recorder.Body.String()), &res)
			assert.Equal(t, tc.wantBody, res)
		})
	}
}
