package util

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

type GoCmd struct {
}

func (g *GoCmd) ModInit(pkg string) {
	err := g.Exec("mod", "init", pkg)
	if err != nil {
		fmt.Println(err)
	}
}

func (g *GoCmd) ModTidy() {
	err := g.Exec("mod", "tidy")
	if err != nil {
		fmt.Println(err)
	}
}

func (g *GoCmd) Fmt(paths ...string) {
	var err error
	if len(paths) > 0 {
		err = g.ExecWithoutOutput("fmt", paths[0])
	} else {
		err = g.ExecWithoutOutput("fmt")
	}
	if err != nil {
		fmt.Println(err)
	}
}

func (g *GoCmd) FmtAll(dir ...string) {
	wd, _ := os.Getwd()
	var paths []string
	filepath.WalkDir(wd, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			for _, key := range dir {
				if strings.HasSuffix(path, key) {
					paths = append(paths, path)
				}
			}
		}
		return err
	})
	wg := sync.WaitGroup{}
	wg.Add(len(paths))
	for _, path := range paths {
		go func(path string) {
			g.Fmt(path)
			wg.Done()
		}(path)
	}
	wg.Wait()
}

func (g *GoCmd) Run(file string) {
	err := g.Exec("build")
	if err != nil {
		fmt.Println(err)
		return
	}
	output, err := exec.Command("./" + file).CombinedOutput()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(output))
}

func (g *GoCmd) Get(pkg string) {
	err := g.Exec("get", pkg)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func (g *GoCmd) Install(pkg string) {
	err := g.Exec("get", "-u", pkg)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = g.Exec("install", pkg)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func (g *GoCmd) List() (*GoList, error) {
	cmd := exec.Command("go", "list", "-m", "-json")
	var output []byte
	err := PipeOutput(cmd, func(line []byte) {
		output = append(output, line...)
	})
	if err != nil {
		return nil, err
	}
	l := &GoList{}
	err = json.Unmarshal(output, l)
	if err != nil {
		return nil, err
	}
	return l, nil
}

func (g *GoCmd) Exec(args ...string) error {
	cmd := exec.Command("go", args...)
	return PipeOutput(cmd, func(line []byte) {
		fmt.Println(string(line))
	})
}

func (g *GoCmd) ExecWithoutOutput(args ...string) error {
	cmd := exec.Command("go", args...)
	return PipeOutput(cmd, func(line []byte) {})
}

type GoList struct {
	Path      string `json:"Path"`
	Main      bool   `json:"Main"`
	Dir       string `json:"Dir"`
	GoMod     string `json:"GoMod"`
	GoVersion string `json:"GoVersion"`
}
