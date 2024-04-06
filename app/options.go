package app

type Options struct {
	Addr         string `env:"ADDR" envDefault:"0.0.0.0:8080"`
	RedisConnStr string `env:"REDIS_CONN_STR,required,unset"`
}
