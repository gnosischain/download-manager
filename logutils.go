package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"
)

var l = log.New(os.Stderr, "", 0)

// TabbedLog ...
var TabbedLog = func(msg string) {
	l.Println(fmt.Sprintf("\t%s", msg))
}

// JSONLog ...
var JSONLog = func(a interface{}) {
	if a != nil {
		res, err := json.MarshalIndent(a, "", "    ")
		if err == nil {
			l.Println(fmt.Sprintf("\n"))
			l.Println(GreenBoldBrightColor(string(res)))
			l.Println(fmt.Sprintf("\n"))
		}
	}
}

// SimpleLog ...
var SimpleLog = func(msg string, a ...interface{}) {
	l.Println(fmt.Sprintf(msg, a...))
}

// InfoLog ...
var InfoLog = func(msg string, a ...interface{}) {
	l.Println(WhiteBoldBrightColor(fmt.Sprintf(msg, a...)))
}

// AsyncCallsLog ...
var AsyncCallsLog = func(msg string, a ...interface{}) {
	l.Println(MagentaBoldBrightColor(fmt.Sprintf(msg, a...)))
}

// ErrorLog ...
var ErrorLog = func(msg string, a ...interface{}) error {
	err := fmt.Sprintf(msg, a...)
	var buf [16 * 1024]byte
	stack := buf[0:runtime.Stack(buf[:], false)]
	l.Fatalln(RedBoldBrightColor(err + "\n\n" + string(stack)))
	return errors.New(err)
}

// FatalLog ...
var FatalLog = func(msg string, a ...interface{}) error {
	err := fmt.Sprintf(msg, a...)
	var buf [16 * 1024]byte
	stack := buf[0:runtime.Stack(buf[:], false)]
	l.Fatalln(RedBoldBrightColor(err + "\n\n" + string(stack)))
	return errors.New(err)
}

// SuccessLog ...
var SuccessLog = func(msg string, a ...interface{}) {
	l.Println(GreenBoldBrightColor(fmt.Sprintf(msg, a...)))
}

// WarningLog ...
var WarningLog = func(msg string, a ...interface{}) {
	l.Println(YellowBoldBrightColor(fmt.Sprintf(msg, a...)))
}
