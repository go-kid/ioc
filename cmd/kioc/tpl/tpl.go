package tpl

import (
	"bytes"
	"embed"
	"text/template"
)

//go:embed *.tpl
var tpls embed.FS

func GetTemplate(key string, data interface{}) []byte {
	content, err := tpls.ReadFile(key + ".tpl")
	if err != nil {
		panic("read template " + key + "failed: " + err.Error())
	}
	buffer := bytes.NewBuffer(nil)
	tpl, err := template.New(key).Parse(string(content))
	if err != nil {
		panic("parse template " + key + " failed: " + err.Error())
	}
	err = tpl.Execute(buffer, data)
	if err != nil {
		panic("execute template " + key + " failed: " + err.Error())
	}
	return buffer.Bytes()
}
