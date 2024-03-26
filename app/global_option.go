package app

var globalOptions []SettingOption

func Settings(ops ...SettingOption) {
	globalOptions = append(globalOptions, ops...)
}
