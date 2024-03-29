package syslog

import (
	"fmt"
	"log"
	"os"
	"strings"
)

type Logger interface {
	Level(lv Lv) Logger
	Pref(pref any) Logger

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
	pref     []string
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
		l.print(LvPanic, v...)
		panic(v)
	}
}

func (l *logger) Panicf(format string, v ...any) {
	if LvPanic >= l.lv {
		l.printf(LvPanic, format, v...)
		panic(fmt.Sprintf(format, v...))
	}
}

func (l *logger) Fatal(v ...any) {
	if LvFatal >= l.lv {
		l.print(LvFatal, v...)
		Fatal(v)
	}
}

func (l *logger) Fatalf(format string, v ...any) {
	if LvFatal >= l.lv {
		l.printf(LvFatal, format, v...)
		Fatal(fmt.Sprintf(format, v...))
	}
}

func (l *logger) print(lv Lv, v ...any) {
	if lv >= l.lv {
		sb := strings.Builder{}
		sb.WriteString(l.levelPre[lv])
		if len(l.pref) != 0 {
			for _, s := range l.pref {
				sb.WriteString(" " + s)
			}
		}
		l.logger.Println(append([]any{sb.String()}, v...)...)
	}
}

func (l *logger) printf(lv Lv, format string, v ...any) {
	if lv >= l.lv {
		sb := strings.Builder{}
		sb.WriteString(l.levelPre[lv])
		if len(l.pref) != 0 {
			for _, s := range l.pref {
				sb.WriteString(" " + s)
			}
		}
		l.logger.Printf(sb.String()+" "+format, v...)
	}
}

func (l *logger) Level(lv Lv) Logger {
	return New(lv)
}

func (l *logger) clone() *logger {
	return &logger{
		logger:   l.logger,
		lv:       l.lv,
		levelPre: l.levelPre,
		pref:     l.pref,
	}
}

func (l *logger) Pref(pref any) Logger {
	clone := l.clone()
	clone.pref = append(clone.pref, fmt.Sprintf("[%v]", pref))
	return clone
}
