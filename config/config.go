package config

type Config struct {
	AdminUsername string
	AdminPassword string
}

func LoadConfig() (*Config, error) {
	return &Config{
		AdminUsername: "admin",
		AdminPassword: "admin",
	}, nil
}
