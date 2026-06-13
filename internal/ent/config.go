package ent

type Config struct {
	Cap     int
	FreeCap int
}

func DefaultConfig() Config {
	return Config{
		Cap:     1000,
		FreeCap: 1024,
	}
}
