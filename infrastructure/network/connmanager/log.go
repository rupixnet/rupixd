package connmanager

import (
	"github.com/rupixnet/rupixd/infrastructure/logger"
	"github.com/rupixnet/rupixd/util/panics"
)

var log = logger.RegisterSubSystem("CMGR")
var spawn = panics.GoroutineWrapperFunc(log)

