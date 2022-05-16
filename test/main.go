package main

import (
	log "github.com/Golang-Tools/loggerhelper/v2"
)

func main() {
	log.Info("test1")
	log.Set(log.WithLevel("WARN"), log.WithExtFields(log.Dict{"app": "l1"}))
	log.Info("test2")
	log.Warn("test3")
	Logger1 := log.Export()

	log.Set(log.WithLevel("Debug"), log.WithExtFields(log.Dict{"app": "l2"}))
	log.Debug("test4", log.Dict{"a": 1})
	log.Warn("test5", log.Dict{"a": 1})
	Logger2 := log.Export()

	Logger1.Debug("test logger1")
	Logger2.Debug("test logger2")
	log.Set(log.WithExtFields(log.Dict{}))
	log.Warn("test no ext fields")
}
