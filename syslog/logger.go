package syslog

import (
	"log"
	"os"
)

type Logger interface {
	Level(lv Lv) Logger

	Trace(v ...any)
	Tracef(format string, v ...any)
	Debug(v ...any)
	Debugf(format string, v ...any)
	Info(v ...any)
	Infof(format string, v ...any)
	Warn(v ...any)
	Warnf(format string, v ...any)
	Error(v ...any)
	Errorf(format string, v ...any)
	Panic(v ...any)
	Panicf(format string, v ...any)
	Fatal(v ...any)
	Fatalf(format string, v ...any)
}

type logger struct {
	logger   *log.Logger
	lv       Lv
	levelPre []string
}

func New(lv Lv) Logger {
	var lp = make([]string, LvFatal+1)
	for i := LvTrace; i <= LvFatal; i++ {
		lp[i] = clCode[lvColor[i]] + lvPrefix[i] + reset
	}
	return &logger{
		lv:       lv,
		logger:   log.New(os.Stderr, "[ioc] ", log.LstdFlags),
		levelPre: lp,
	}
}

func (l *logger) print(lv Lv, v ...any) {
	if lv >= l.lv {
		l.logger.Println(append([]any{l.levelPre[lv]}, v...)...)
	}
}

func (l *logger) printf(lv Lv, format string, v ...any) {
	if lv >= l.lv {
		l.logger.Printf(l.levelPre[lv]+" "+format, v...)
	}
}

func (l *logger) Trace(v ...any) {
	l.print(LvTrace, v...)
}

func (l *logger) Tracef(format string, v ...any) {
	l.printf(LvTrace, format, v...)
}

func (l *logger) Debug(v ...any) {
	l.print(LvDebug, v...)
}

func (l *logger) Debugf(format string, v ...any) {
	l.printf(LvDebug, format, v...)
}

func (l *logger) Info(v ...any) {
	l.print(LvInfo, v...)
}

func (l *logger) Infof(format string, v ...any) {
	l.printf(LvInfo, format, v...)
}

func (l *logger) Warn(v ...any) {
	l.print(LvWarn, v...)
}

func (l *logger) Warnf(format string, v ...any) {
	l.printf(LvWarn, format, v...)
}

func (l *logger) Error(v ...any) {
	l.print(LvError, v...)
}

func (l *logger) Errorf(format string, v ...any) {
	l.printf(LvError, format, v...)
}

func (l *logger) Panic(v ...any) {
	if LvPanic >= l.lv {
		l.logger.Panicln(append([]any{l.levelPre[LvPanic]}, v...)...)
	}
}

func (l *logger) Panicf(format string, v ...any) {
	if LvPanic >= l.lv {
		l.logger.Panicf(l.levelPre[LvPanic]+" "+format, v...)
	}
}

func (l *logger) Fatal(v ...any) {
	if LvFatal >= l.lv {
		l.logger.Fatalln(append([]any{l.levelPre[LvFatal]}, v...)...)
	}
}

func (l *logger) Fatalf(format string, v ...any) {
	if LvFatal >= l.lv {
		l.logger.Fatalf(l.levelPre[LvFatal]+" "+format, v...)
	}
}

func (l *logger) Level(lv Lv) Logger {
	return New(lv)
}
