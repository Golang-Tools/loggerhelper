package main

import (
	log "github.com/Golang-Tools/loggerhelper/v2"
)

func main() {
	log.Info("test1")
	log.SetLogger(log.WithLevel("WARN"), log.WithExtFields(log.Dict{"d": 3}))
	log.Info("test2")
	log.Warn("test3")
	log.SetLogger(log.WithLevel("Debug"), log.WithExtFields(log.Dict{"e": 3}))
	log.Debug("test4", log.Dict{"a": 1})
	log.Warn("test5", log.Dict{"a": 1})
}
