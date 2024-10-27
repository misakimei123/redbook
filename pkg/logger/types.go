package logger

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

func example() {
	var l Logger
	l.Info("user wechat id %d", 123)
}

type LoggerV1 interface {
	Debug(msg string, args ...Field)
	Info(msg string, args ...Field)
	Warn(msg string, args ...Field)
	Error(msg string, args ...Field)
}

type Field struct {
	Key string
	Val any
}

func exampleV1() {
	var l LoggerV1
	l.Info("user wechat ", Field{Key: "id", Val: 1})
}

// 不建议
type LoggerV2 interface {
	//参数必须偶数个，以k v...顺序传递
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}
