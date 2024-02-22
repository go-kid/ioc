package config

type Config struct {
	Val1 int
	Val2 string
	Val3 bool
	Val4 float64
	Val5 []byte
}

func (c *Config) Prefix() string {
	return "config"
}
