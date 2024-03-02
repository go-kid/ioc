package binder

import (
	"bytes"
	"fmt"
	"github.com/go-kid/ioc/configure"
	"text/template"
)

type ExpressionBinder struct {
	*ViperBinder
}

func NewExpressionBinder(configType string) configure.Binder {
	return &ExpressionBinder{
		NewViperBinder(configType).(*ViperBinder),
	}
}

func (e *ExpressionBinder) SetConfig(c []byte) error {
	err := e.ViperBinder.SetConfig(c)
	if err != nil {
		return err
	}

	tpl, err := template.New("").Parse(string(c))
	if err != nil {
		return fmt.Errorf("parse config template failed: %v", err)
	}
	var m = map[string]any{}
	err = e.Viper.Unmarshal(&m)
	if err != nil {
		return err
	}
	buffer := bytes.NewBuffer(nil)
	err = tpl.Execute(buffer, m)
	if err != nil {
		return fmt.Errorf("execute config template failed: %v", err)
	}
	return e.ViperBinder.SetConfig(buffer.Bytes())
}
