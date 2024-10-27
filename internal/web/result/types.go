package result

type Result struct {
	Code int    `json:"code"`
	Msg  string `json:"message"`
	Data any    `json:"data"`
}

var (
	RetSuccess = Result{
		Msg: "success",
	}
	RetLoginSuccess = Result{
		Msg: "login success",
	}
	RetDataNotReady = Result{
		Msg: "data is not ready now, please retry later",
	}
	RetSystemError = Result{
		Code: UserInternalServerError,
		Msg:  "system error",
	}
	RetNeedPhoneNumber = Result{
		Code: UserInvalidInput,
		Msg:  "please input phone number",
	}
	RetTooFrequent = Result{
		Code: 4,
		Msg:  "send too frequent, please retry later",
	}
	RetVerifyFail = Result{
		Code: 4,
		Msg:  "verify fail",
	}
	RetAuthCodeError = Result{
		Msg:  "授权码有误",
		Code: 4,
	}
	RetIllegalRequest = Result{
		Msg:  "illegal request",
		Code: 4,
	}
	RetIllegalNumberFormat = Result{
		Msg:  "illegal number format",
		Code: UserInvalidInput,
	}
)

const (
	// UserInvalidInput 统一的用户模块的输入错误
	UserInvalidInput = 401001
	// UserInvalidOrPassword 用户名错误或者密码不对
	UserInvalidOrPassword = 401002
	// UserDuplicateEmail 用户邮箱冲突
	UserDuplicateEmail = 401003
	// UserInternalServerError 统一的用户模块的系统错误
	UserInternalServerError = 501001
)

const (
	// ArticleInvalidInput 文章模块的统一的错误码
	ArticleInvalidInput        = 402001
	ArticleInternalServerError = 502001
)
