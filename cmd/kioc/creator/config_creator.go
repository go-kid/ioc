package creator

import (
	"gitlab.iglooinsure.com/axinan/backend/gi2/common/cmd/gi2ctl/util"
	"gitlab.iglooinsure.com/axinan/backend/gi2/common/util/config_merge"
	"gitlab.iglooinsure.com/axinan/backend/gi2/common/util/properties"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strings"
)

type Properties = map[string]interface{}

type ConfigFile struct {
	Env       string
	Config    Properties
	Overwrite bool
}

func (y *ConfigFile) Create() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	var fileNameArr = []string{"config"}
	if y.Env != "" {
		fileNameArr = append(fileNameArr, strings.ToLower(y.Env))
	}
	err = util.MkDir("resources")
	if err != nil {
		return err
	}
	var file = &File{
		RelativePath: "deployment",
		Template:     "config",
		Ext:          "yaml",
		SkipIfExists: false,
		FileName:     strings.Join(fileNameArr, "-"),
	}
	data, err := y.parseProperties()
	if err != nil {
		return err
	}

	oldData, err := y.loadOldConfig(filepath.Join(wd, "resources", file.GetFileName()))
	if err != nil {
		return err
	}
	if y.Overwrite {
		data = config_merge.MergeMap(oldData, data)
	} else {
		data = config_merge.MergeMap(data, oldData)
	}
	out, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	file.SetAttribute("Config", string(out))
	return file.Create()
}

func (y *ConfigFile) parseProperties() (Properties, error) {
	data := properties.PropMapExpand(y.Config)
	return data, nil
}

func (y *ConfigFile) loadOldConfig(fileName string) (Properties, error) {
	var oldData = make(Properties)
	if util.FileExist(fileName) {
		bytes, err := os.ReadFile(fileName)
		if err != nil {
			return nil, err
		}
		err = yaml.Unmarshal(bytes, &oldData)
		if err != nil {
			return nil, err
		}
	}
	return oldData, nil
}
