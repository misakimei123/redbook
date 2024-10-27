package config

type DBConfig struct {
	DSN string
}

type RedisConfig struct {
	Addr     string
	DB       int
	Password string
}

type WebookConfig struct {
	DB    DBConfig
	Redis RedisConfig
}
