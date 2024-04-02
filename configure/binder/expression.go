package binder

import (
	"bytes"
	"github.com/pkg/errors"
	"text/template"
)

type ExpressionBinder struct {
	*ViperBinder
}

func NewExpressionBinder(configType string) *ExpressionBinder {
	return &ExpressionBinder{
		NewViperBinder(configType),
	}
}

func (e *ExpressionBinder) SetConfig(c []byte) error {
	err := e.ViperBinder.SetConfig(c)
	if err != nil {
		return errors.Wrapf(err, "viper binder set config: %s", string(c))
	}

	tpl, err := template.New("").Parse(string(c))
	if err != nil {
		return errors.Wrapf(err, "parse config template failed: %s", string(c))
	}
	var m = map[string]any{}
	err = e.Viper.Unmarshal(&m)
	if err != nil {
		return err
	}
	buffer := bytes.NewBuffer(nil)
	err = tpl.Execute(buffer, m)
	if err != nil {
		return errors.Wrapf(err, "execute config template failed: %s", string(c))
	}
	err = e.ViperBinder.SetConfig(buffer.Bytes())
	if err != nil {
		return errors.Wrapf(err, "viper binder set config: %s", buffer.String())
	}
	return nil
}

func (e *ExpressionBinder) Set(path string, val any) {
	//todo: fill tpl logic
	e.ViperBinder.Set(path, val)
}
