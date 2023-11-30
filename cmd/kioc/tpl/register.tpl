// {{ .GenerateTitle }}
package {{.Pkg}}

import (
	"github.com/go-kid/ioc"
	{{range .Imports}}"{{.}}"
	{{end}}
)

func init() {
	ioc.Register(
    {{- range .Components}}
	    {{if eq .Kind "type"}}new({{.Pkg}}.{{.Name}}),{{- end}}
	    {{- if eq .Kind "func"}}{{.Pkg}}.{{.Name}}(),{{- end}}
    {{- end}}
	)
}
