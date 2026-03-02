package syslog

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
)

var lvToSlogLevel = map[Lv]slog.Level{
	LvTrace: slog.LevelDebug - 4,
	LvDebug: slog.LevelDebug,
	LvInfo:  slog.LevelInfo,
	LvWarn:  slog.LevelWarn,
	LvError: slog.LevelError,
	LvPanic: slog.LevelError + 4,
	LvFatal: slog.LevelError + 8,
}

type slogAdapter struct {
	logger *slog.Logger
	lv     Lv
	pref   []string
}

func NewSlogAdapter(handler slog.Handler) Logger {
	return &slogAdapter{
		logger: slog.New(handler),
		lv:     LvInfo,
	}
}

func (a *slogAdapter) slogLevel(lv Lv) slog.Level {
	if sl, ok := lvToSlogLevel[lv]; ok {
		return sl
	}
	return slog.LevelInfo
}

func (a *slogAdapter) enabled(lv Lv) bool {
	return lv >= a.lv
}

func (a *slogAdapter) prefixStr() string {
	if len(a.pref) == 0 {
		return ""
	}
	return strings.Join(a.pref, " ") + " "
}

func (a *slogAdapter) log(lv Lv, msg string) {
	if !a.enabled(lv) {
		return
	}
	a.logger.Log(context.Background(), a.slogLevel(lv), a.prefixStr()+msg)
}

func (a *slogAdapter) Trace(v ...any)                 { a.log(LvTrace, fmt.Sprint(v...)) }
func (a *slogAdapter) Tracef(format string, v ...any)  { a.log(LvTrace, fmt.Sprintf(format, v...)) }
func (a *slogAdapter) Debug(v ...any)                  { a.log(LvDebug, fmt.Sprint(v...)) }
func (a *slogAdapter) Debugf(format string, v ...any)  { a.log(LvDebug, fmt.Sprintf(format, v...)) }
func (a *slogAdapter) Info(v ...any)                   { a.log(LvInfo, fmt.Sprint(v...)) }
func (a *slogAdapter) Infof(format string, v ...any)   { a.log(LvInfo, fmt.Sprintf(format, v...)) }
func (a *slogAdapter) Warn(v ...any)                   { a.log(LvWarn, fmt.Sprint(v...)) }
func (a *slogAdapter) Warnf(format string, v ...any)   { a.log(LvWarn, fmt.Sprintf(format, v...)) }
func (a *slogAdapter) Error(v ...any)                  { a.log(LvError, fmt.Sprint(v...)) }
func (a *slogAdapter) Errorf(format string, v ...any)  { a.log(LvError, fmt.Sprintf(format, v...)) }

func (a *slogAdapter) Panic(v ...any) {
	if a.enabled(LvPanic) {
		msg := fmt.Sprint(v...)
		a.log(LvPanic, msg)
		panic(msg)
	}
}

func (a *slogAdapter) Panicf(format string, v ...any) {
	if a.enabled(LvPanic) {
		msg := fmt.Sprintf(format, v...)
		a.log(LvPanic, msg)
		panic(msg)
	}
}

func (a *slogAdapter) Fatal(v ...any) {
	if a.enabled(LvFatal) {
		a.log(LvFatal, fmt.Sprint(v...))
		Fatal(v)
	}
}

func (a *slogAdapter) Fatalf(format string, v ...any) {
	if a.enabled(LvFatal) {
		msg := fmt.Sprintf(format, v...)
		a.log(LvFatal, msg)
		Fatal(msg)
	}
}

func (a *slogAdapter) Level(lv Lv) Logger {
	clone := a.clone()
	clone.lv = lv
	return clone
}

func (a *slogAdapter) clone() *slogAdapter {
	return &slogAdapter{
		logger: a.logger,
		lv:     a.lv,
		pref:   append([]string{}, a.pref...),
	}
}

func (a *slogAdapter) Pref(pref any) Logger {
	clone := a.clone()
	clone.pref = append(clone.pref, fmt.Sprintf("[%v]", pref))
	return clone
}
