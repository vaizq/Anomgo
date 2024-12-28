package log

import (
	"log"
	"os"
)

var Debug *log.Logger
var Info *log.Logger
var Error *log.Logger

func Init() {
	Debug = log.New(os.Stdout, "DEBUG\t", log.Lshortfile|log.Ldate|log.Ltime)
	Info = log.New(os.Stdout, "INFO\t", log.Lshortfile|log.Ldate|log.Ltime)
	Error = log.New(os.Stderr, "ERROR\t", log.Lshortfile|log.Ldate|log.Ltime)
}
