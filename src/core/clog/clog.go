/*
Copyright (C) 2018 by Martin Langlotz aka stackshadow

This file is part of gopilot, an rewrite of the copilot-project in go

gopilot is free software: you can redistribute it and/or modify
it under the terms of the GNU Lesser General Public License as published by
the Free Software Foundation, version 3 of this License

gopilot is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Lesser General Public License for more details.

You should have received a copy of the GNU Lesser General Public License
along with gopilot.  If not, see <http://www.gnu.org/licenses/>.
*/

package clog

import (
	"flag"
	"fmt"
)

var isDebug bool

type Logger struct {
	prefix string
}

func ParseCmdLine() {
	flag.BoolVar(&isDebug, "v", false, "Enable debug")
}

func Init() {

}

func New(prefix string) Logger {
	newLog := Logger{
		prefix: prefix,
	}
	return newLog
}

func EnableDebug() {
	isDebug = true
}

func (logging *Logger) Debug(group, message string) {
	if isDebug == true {
		fmt.Printf("[DEBUG] [%s] [%s] %s\n", logging.prefix, group, message)
	}
}

func (logging *Logger) Info(group, message string) {
	fmt.Printf("[INFO] [%s] [%s] %s\n", logging.prefix, group, message)
}

func (logging *Logger) Error(group, message string) {
	fmt.Printf("[ERROR] [%s] [%s] %s\n", logging.prefix, group, message)
}
