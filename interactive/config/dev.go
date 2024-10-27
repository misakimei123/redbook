//go:build !k8s

package config

var Config = WebookConfig{
	DB: DBConfig{
		DSN: "root:root@tcp(127.0.0.1:3308)/redbook",
	},
	Redis: RedisConfig{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       1,
	},
}
