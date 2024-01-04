package syslog

import (
	"log"
	"os"
)

var std = log.New(os.Stderr, "[ioc] ", log.LstdFlags)

var (
	Fatal  = std.Fatalln
	Fatalf = std.Fatalf
	Info   = std.Println
	Infof  = std.Printf
	Panic  = std.Panicln
	Panicf = std.Panicf
)
