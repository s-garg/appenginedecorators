package core

import (
	"time"
	"appengine"
)

func DebugMsg(ctx appengine.Context, msg string) {
	ctx.Debugf("[DEBUG]: %s : %s", logTime(), msg)
}

func InfoMsg(ctx appengine.Context, msg string) {
	ctx.Infof("[INFO]: %s : %s", logTime(), msg)
}

func WarningMsg(ctx appengine.Context, msg string) {
	ctx.Warningf("[WARNING]: %s : %s", logTime(), msg)
}

func ErrorMsg(ctx appengine.Context, msg string) {
	ctx.Errorf("[ERROR]: %s : %s", logTime(), msg)
}

// returns the current time in RFC3339 format
func logTime() string {
	return time.Now().Format(time.RFC3339)
}
