package query

type Config struct {
	Cap int
}

func DefaultConfig() Config {
	return Config{
		Cap: MaxViews,
	}
}
