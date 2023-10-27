package loader

type NopLoader struct{}

func (n *NopLoader) LoadConfig(u string) ([]byte, error) { return nil, nil }
