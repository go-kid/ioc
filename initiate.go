package ioc

import (
	"github.com/go-kid/ioc/app"
	"github.com/go-kid/ioc/factory/post_processors/definition_registry_post_processors"
)

func init() {
	app.Settings(
		app.SetComponents(
			&definition_registry_post_processors.PropTagScanProcessor{},
			&definition_registry_post_processors.ValueTagScanProcessor{},
			&definition_registry_post_processors.WireTagScanProcessor{},
		),
	)
}
