package scanner

type policy struct {
	tag     string
	handler ExtTagHandler
}

func DefaultScanPolicy(tag string, handler ExtTagHandler) ScanPolicy {
	return &policy{
		tag:     tag,
		handler: handler,
	}
}

func (p *policy) Tag() string {
	return p.tag
}

func (p *policy) ExtHandler() ExtTagHandler {
	return p.handler
}
