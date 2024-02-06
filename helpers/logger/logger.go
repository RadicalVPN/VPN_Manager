package logger

import (
	"log"
	"os"
)

var Debug *log.Logger
var Info *log.Logger
var Warning *log.Logger
var Error *log.Logger

const flags = log.Ldate | log.Ltime

func init() {
	Debug = log.New(os.Stdout, "[DEBUG]: ", flags)
	Info = log.New(os.Stdout, "[ INFO]: ", flags)
	Warning = log.New(os.Stdout, "[ WARN]: ", flags)
	Error = log.New(os.Stderr, "[ERROR]: ", flags)
}
