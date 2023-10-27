package loader

type RawLoader struct{}

func (r *RawLoader) LoadConfig(u string) ([]byte, error) {
	return []byte(u), nil
}
