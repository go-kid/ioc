package ioc

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
)

type setting struct {
	registry *registry
}

type SettingOption func(s *setting)

func Run(cfgPath string, ops ...SettingOption) error {
	var s = &setting{
		registry: _registry,
	}
	for _, op := range ops {
		op(s)
	}
	err := initConfig(cfgPath)
	if err != nil {
		return fmt.Errorf("initConfig failed: %v", err)
	}
	initComponentPostProcessors(s.registry)
	for _, meta := range s.registry.GetComponents() {
		err := initialize(meta, s.registry)
		if err != nil {
			return fmt.Errorf("initialize failed: %v", err)
		}
	}

	err = callRunners(s.registry)
	if err != nil {
		return fmt.Errorf("runners failed: %v", err)
	}
	return nil
}

func RunTest(t *testing.T, cfgPath string, ops ...SettingOption) {
	var testOps = []SettingOption{
		setRegistry(NewRegistry()),
	}
	testOps = append(testOps, ops...)
	err := Run(cfgPath, testOps...)
	if t != nil {
		assert.NoError(t, err)
	}
}

func RunDebug(cfgPath string, ops ...SettingOption) error {
	r := NewRegistry()
	var testOps = []SettingOption{setRegistry(r)}
	testOps = append(testOps, ops...)
	err := Run(cfgPath, testOps...)
	if err != nil {
		return err
	}
	metas := r.GetComponents()
	sort.Slice(metas, func(i, j int) bool {
		if len(metas[i].DependsBy) != len(metas[j].DependsBy) {
			return len(metas[i].DependsBy) > len(metas[j].DependsBy)
		}
		return metas[i].ID() < metas[j].ID()
	})
	for _, m := range metas {
		m.Describe()
	}
	return nil
}

/*
setRegistry
used for unit testing
*/
func setRegistry(r *registry) SettingOption {
	return func(s *setting) {
		s.registry = r
	}
}

func SetComponents(c ...interface{}) SettingOption {
	return func(s *setting) {
		s.registry.Register(c...)
	}
}
