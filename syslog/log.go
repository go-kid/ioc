package syslog

import "log"

func Fatal(v ...any) {
	log.Fatal(append([]any{"[ioc]"}, v...)...)
}

func Fatalf(format string, v ...any) {
	log.Fatalf("[ioc] "+format, v...)
}

func Info(v ...any) {
	log.Println(append([]any{"[ioc]"}, v...)...)
}

func Infof(format string, v ...any) {
	log.Printf("[ioc] "+format, v...)
}

func Panic(v ...any) {
	log.Panic(append([]any{"[ioc]"}, v...)...)
}

func Panicf(format string, v ...any) {
	log.Panicf("[ioc] "+format, v...)
}
