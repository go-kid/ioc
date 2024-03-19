package loader

type RawLoader []byte

func NewRawLoader(raw []byte) RawLoader {
	return RawLoader(raw)
}

func (r RawLoader) LoadConfig() ([]byte, error) {
	return r, nil
}
