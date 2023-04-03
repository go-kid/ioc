package kid_ioc

import (
	"github.com/kidhat/kid-ioc/registry"
	"github.com/stretchr/testify/assert"
	"testing"
)

func RunTest(t *testing.T, ops ...SettingOption) {
	var testOps = []SettingOption{
		optionSetRegistry(registry.NewRegistry()),
	}
	testOps = append(testOps, ops...)
	err := Run(testOps...)
	if t != nil {
		assert.NoError(t, err)
	}
}

/*
setRegistry
used for unit testing
*/
func optionSetRegistry(r *registry.Registry) SettingOption {
	return func(s *setting) {
		s.registry = r
	}
}
