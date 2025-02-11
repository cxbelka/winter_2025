package config

type Config struct {
	DB        *DBcfg   `envconfig:"DATABASE"`
	HTTP      *HTTPcfg `envconfig:"SERVER"`
	JWTSecret string   `envconfig:"JWT_SECRET"`
}

type DBcfg struct {
	Port   int    `envconfig:"PORT"`
	User   string `envconfig:"USER"`
	Pass   string `envconfig:"PASSWORD"`
	DBName string `envconfig:"NAME"`
	Host   string `envconfig:"HOST"`
}

type HTTPcfg struct {
	Port int `envconfig:"PORT"`
}
