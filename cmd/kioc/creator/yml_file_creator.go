package creator

type YamlFile struct {
	Pkg          string
	Template     string
	SkipIfExists bool
	Data         map[string]interface{}
	FileName     string
}

func (y *YamlFile) Create() error {
	var file = &File{
		RelativePath: y.Pkg,
		Template:     y.Template,
		Ext:          "yaml",
		SkipIfExists: y.SkipIfExists,
		Data:         y.Data,
		FileName:     y.FileName,
	}
	return file.Create()
}
