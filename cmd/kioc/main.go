package main

import (
	"github.com/go-kid/ioc/cmd/kioc/cmd"
)

func main() {
	if err := cmd.Root.Execute(); err != nil {
		panic(err)
	}
}
