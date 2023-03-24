package kid_ioc

import (
	"github.com/kid-hash/kid-ioc/registry"
	"sort"
)

func RunDebug(ops ...SettingOption) error {
	r := registry.NewRegistry()
	var testOps = []SettingOption{optionSetRegistry(r)}
	testOps = append(testOps, ops...)
	err := Run(testOps...)
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
