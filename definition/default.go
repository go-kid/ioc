package definition

type PriorityComponent struct{}

func (i *PriorityComponent) Priority() {}

type WirePrimaryComponent struct{}

func (i *WirePrimaryComponent) Primary() {}

type LazyInitComponent struct{}

func (i *LazyInitComponent) LazyInit() {}
