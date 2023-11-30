package util

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
)

func Error(msg string, args ...interface{}) {
	fmt.Printf("err: "+msg, args...)
	os.Exit(1)
}

func MkDir(dir string) error {
	return os.MkdirAll(dir, 0777)
	//if FileExist(dir) {
	//	return nil
	//}
	//if err := os.Mkdir(dir, 0754); err != nil {
	//	return err
	//}
	//return nil
}

func FileExist(file string) bool {
	_, err := os.Stat(file)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func GetPkgName() (string, error) {
	g := &GoCmd{}
	list, err := g.List()
	if err != nil {
		return "", err
	}
	return list.Path, nil
}

func FindGoModFile() (string, error) {
	g := &GoCmd{}
	list, err := g.List()
	if err != nil {
		return "", err
	}
	return list.GoMod, nil
}

func PipeOutput(cmd *exec.Cmd, pipe func(line []byte)) error {
	out, _ := cmd.StdoutPipe()
	defer out.Close()
	cmd.Stderr = cmd.Stdout
	if err := cmd.Start(); err != nil {
		panic(err)
	}
	reader := bufio.NewReader(out)
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			if err.Error() != "EOF" {
				return err
			}
			break
		}
		pipe(line)
	}
	return nil
}
