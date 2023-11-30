package creator

import (
	"errors"
	"fmt"
	"github.com/go-kid/ioc/cmd/kioc/tpl"
	"github.com/go-kid/ioc/cmd/kioc/util"
	"os"
	"path/filepath"
)

type File struct {
	root         string
	Path         string
	RelativePath string
	Template     string
	SkipIfExists bool
	Data         map[string]interface{}
	FileName     string
	Ext          string
	PostHandler  func(file *File)
}

func (g *File) Create() error {
	root, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(root)
	g.root = root

	if err := g.init(); err != nil {
		return err
	}
	if g.Template == "" {
		return nil
	}

	if g.FileExists() && g.SkipIfExists {
		fmt.Printf("Gen %s Exists, Skip.\n", filepath.Join(g.Path, g.RelativePath, g.GetFileName()))
		return nil
	}
	if err := g.createFile(); err != nil {
		return err
	}
	if g.PostHandler != nil {
		g.PostHandler(g)
	}
	return nil
}

func (g *File) init() error {
	if g.Path == "" {
		pkgName, err := util.GetPkgName()
		if err != nil {
			return err
		}
		g.Path = pkgName
	}
	if err := util.MkDir(g.Wd()); err != nil {
		return err
	}
	if err := os.Chdir(g.Wd()); err != nil {
		return err
	}
	if g.FileName == "" {
		g.FileName = g.Template
	}
	if g.Data == nil {
		g.Data = make(map[string]interface{})
	}
	g.SetAttribute("Path", g.Path)
	g.SetAttribute("RelativePath", g.RelativePath)
	g.SetAttribute("FileName", g.FileName)
	g.SetAttribute("Ext", g.Ext)
	return nil
}

func (g *File) GetFileName() string {
	return fmt.Sprintf("%s.%s", g.FileName, g.Ext)
}

func (g *File) createFile() error {
	template := tpl.GetTemplate(g.Template, g.Data)
	fmt.Printf("Gen %s, ", filepath.Join(g.Path, g.RelativePath, g.GetFileName()))
	err := os.WriteFile(g.GetFileName(), template, 0777)
	if err != nil {
		fmt.Println("Failed.")
		return err
	}
	fmt.Println("Success.")
	return nil
}

func (g *File) Wd() string {
	return filepath.Join(g.root, g.RelativePath)
}

func (g *File) SetAttribute(key string, value interface{}) {
	if g.Data == nil {
		g.Data = make(map[string]interface{})
	}
	g.Data[key] = value
}

func (g *File) FileExists() bool {
	_, err := os.Stat(g.GetFileName())
	return !errors.Is(err, os.ErrNotExist)
}
