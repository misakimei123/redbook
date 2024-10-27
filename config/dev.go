//go:build !k8s

package config

var Config = WebookConfig{
	DB: DBConfig{
		DSN: "root:ahyang@tcp(192.168.252.128:30306)/webook_dev",
	},
	Redis: RedisConfig{
		Addr:     "192.168.252.128:31379",
		Password: "",
		DB:       1,
	},
}
