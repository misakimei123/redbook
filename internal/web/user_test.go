package web

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/misakimei123/redbook/internal/domain"
	"github.com/misakimei123/redbook/internal/service"
	svcmocks "github.com/misakimei123/redbook/internal/service/mocks"
	"github.com/misakimei123/redbook/internal/web/ginadaptor"
	ginadaptormocks "github.com/misakimei123/redbook/internal/web/ginadaptor/mocks"
	"github.com/misakimei123/redbook/internal/web/jwt"
	jwtmocks "github.com/misakimei123/redbook/internal/web/jwt/mocks"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestUserHandler_SignUp(t *testing.T) {
	testCases := []struct {
		name       string
		mock       func(ctrl *gomock.Controller) (service.UserService, service.CodeService, jwt.Handler, ginadaptor.Context)
		reqBuilder func(t *testing.T) *http.Request
		wantCode   int
		wantBody   string
	}{
		{name: "注册成功",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, jwt.Handler, ginadaptor.Context) {
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().SignUp(gomock.Any(), &domain.User{
					Email:    "123@abc.net",
					Password: "hello@world123",
				}).Return(nil)
				codeService := svcmocks.NewMockCodeService(ctrl)
				handler := jwtmocks.NewMockHandler(ctrl)
				context := ginadaptormocks.NewMockContext(ctrl)
				return userService, codeService, handler, context
			},
			reqBuilder: func(t *testing.T) *http.Request {
				request, err := http.NewRequest(http.MethodPost, "/users/signup", bytes.NewReader([]byte(`{
"email":"123@abc.net",
"password":"hello@world123",
"confirmpassword":"hello@world123"
}`)))
				request.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return request
			},
			wantCode: http.StatusOK,
			wantBody: "123@abc.net signup success.",
		},
		{name: "Bind request error",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, jwt.Handler, ginadaptor.Context) {
				userService := svcmocks.NewMockUserService(ctrl)
				//userService.EXPECT().SignUp(gomock.Any(), &domain.User{
				//	Email:    "123@abc.net",
				//	Password: "hello@world123",
				//}).Return(nil)
				codeService := svcmocks.NewMockCodeService(ctrl)
				handler := jwtmocks.NewMockHandler(ctrl)
				context := ginadaptormocks.NewMockContext(ctrl)
				return userService, codeService, handler, context
			},
			reqBuilder: func(t *testing.T) *http.Request {
				request, err := http.NewRequest(http.MethodPost, "/users/signup", bytes.NewReader([]byte(`{
"email":"123@abc.net"
"password":"hello@world123"
"confirmpassword":"hello@world123"
}`)))
				request.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return request
			},
			wantCode: http.StatusBadRequest,
			wantBody: "",
		},
		{name: "email format not correct",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, jwt.Handler, ginadaptor.Context) {
				userService := svcmocks.NewMockUserService(ctrl)
				//userService.EXPECT().SignUp(gomock.Any(), &domain.User{
				//	Email:    "123@abc.net",
				//	Password: "hello@world123",
				//}).Return(nil)
				codeService := svcmocks.NewMockCodeService(ctrl)
				handler := jwtmocks.NewMockHandler(ctrl)
				context := ginadaptormocks.NewMockContext(ctrl)
				return userService, codeService, handler, context
			},
			reqBuilder: func(t *testing.T) *http.Request {
				request, err := http.NewRequest(http.MethodPost, "/users/signup", bytes.NewReader([]byte(`{
"email":"123abc.net",
"password":"hello@world123",
"confirmpassword":"hello@world123"
}`)))
				request.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return request
			},
			wantCode: http.StatusOK,
			wantBody: "email format is not correct",
		},
		{name: "password is not match",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, jwt.Handler, ginadaptor.Context) {
				userService := svcmocks.NewMockUserService(ctrl)
				//userService.EXPECT().SignUp(gomock.Any(), &domain.User{
				//	Email:    "123@abc.net",
				//	Password: "hello@world123",
				//}).Return(nil)
				codeService := svcmocks.NewMockCodeService(ctrl)
				handler := jwtmocks.NewMockHandler(ctrl)
				context := ginadaptormocks.NewMockContext(ctrl)
				return userService, codeService, handler, context
			},
			reqBuilder: func(t *testing.T) *http.Request {
				request, err := http.NewRequest(http.MethodPost, "/users/signup", bytes.NewReader([]byte(`{
"email":"123@abc.net",
"password":"hello@world123",
"confirmpassword":"hello@world1234"
}`)))
				request.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return request
			},
			wantCode: http.StatusOK,
			wantBody: "password is not same",
		},
		{name: "password format is not correct",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, jwt.Handler, ginadaptor.Context) {
				userService := svcmocks.NewMockUserService(ctrl)
				//userService.EXPECT().SignUp(gomock.Any(), &domain.User{
				//	Email:    "123@abc.net",
				//	Password: "hello@world123",
				//}).Return(nil)
				codeService := svcmocks.NewMockCodeService(ctrl)
				handler := jwtmocks.NewMockHandler(ctrl)
				context := ginadaptormocks.NewMockContext(ctrl)
				return userService, codeService, handler, context
			},
			reqBuilder: func(t *testing.T) *http.Request {
				request, err := http.NewRequest(http.MethodPost, "/users/signup", bytes.NewReader([]byte(`{
"email":"123@abc.net",
"password":"helloworld123",
"confirmpassword":"helloworld123"
}`)))
				request.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return request
			},
			wantCode: http.StatusOK,
			wantBody: "password format is not correct",
		},
		{name: "duplicate email.",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, jwt.Handler, ginadaptor.Context) {
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().SignUp(gomock.Any(), &domain.User{
					Email:    "123@abc.net",
					Password: "hello@world123",
				}).Return(service.ErrDuplicateEmail)
				codeService := svcmocks.NewMockCodeService(ctrl)
				handler := jwtmocks.NewMockHandler(ctrl)
				context := ginadaptormocks.NewMockContext(ctrl)
				return userService, codeService, handler, context
			},
			reqBuilder: func(t *testing.T) *http.Request {
				request, err := http.NewRequest(http.MethodPost, "/users/signup", bytes.NewReader([]byte(`{
"email":"123@abc.net",
"password":"hello@world123",
"confirmpassword":"hello@world123"
}`)))
				request.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return request
			},
			wantCode: http.StatusOK,
			wantBody: "duplicate email.",
		},
		{name: "system error",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, jwt.Handler, ginadaptor.Context) {
				userService := svcmocks.NewMockUserService(ctrl)
				userService.EXPECT().SignUp(gomock.Any(), &domain.User{
					Email:    "123@abc.net",
					Password: "hello@world123",
				}).Return(errors.New("db error la la la"))
				codeService := svcmocks.NewMockCodeService(ctrl)
				handler := jwtmocks.NewMockHandler(ctrl)
				context := ginadaptormocks.NewMockContext(ctrl)
				return userService, codeService, handler, context
			},
			reqBuilder: func(t *testing.T) *http.Request {
				request, err := http.NewRequest(http.MethodPost, "/users/signup", bytes.NewReader([]byte(`{
"email":"123@abc.net",
"password":"hello@world123",
"confirmpassword":"hello@world123"
}`)))
				request.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return request
			},
			wantCode: http.StatusOK,
			wantBody: "system error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			controller := gomock.NewController(t)
			defer controller.Finish()
			userService, codeService, jwtHandler, logContext := tc.mock(controller)
			handler := NewUserHandler(userService, codeService, jwtHandler, logContext)
			server := gin.Default()
			handler.RegisterRoutes(server)
			req := tc.reqBuilder(t)
			recorder := httptest.NewRecorder()
			server.ServeHTTP(recorder, req)
			assert.Equal(t, tc.wantCode, recorder.Code)
			assert.Equal(t, tc.wantBody, recorder.Body.String())
		})
	}
}

func TestEmailPattern(t *testing.T) {
	testCases := []struct {
		name  string
		email string
		match bool
	}{
		{
			name:  "正常匹配",
			email: "123@abc.net",
			match: true,
		},
		{
			name:  "不带@",
			email: "123abc.net",
			match: false,
		},
	}
	userHandler := NewUserHandler(nil, nil, nil, nil)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			match, err := userHandler.emailRexExp.MatchString(tc.email)
			require.NoError(t, err)
			assert.Equal(t, tc.match, match)
			return
		})
	}
}
