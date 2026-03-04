package main

import (
	"github.com/rupixnet/rupixd/infrastructure/logger"
	"github.com/rupixnet/rupixd/util/panics"
)

var (
	backendLog = logger.NewBackend()
	log        = backendLog.Logger("MNJS")
	spawn      = panics.GoroutineWrapperFunc(log)
)

