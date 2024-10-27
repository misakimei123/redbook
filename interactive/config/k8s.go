//go:build k8s

package config

var Config = WebookConfig{
	DB: DBConfig{
		DSN: "root:ahyang@tcp(mysql:3306)/webook",
	},
	Redis: RedisConfig{
		Addr:     "redis:6379",
		Password: "",
		DB:       1,
	},
}
