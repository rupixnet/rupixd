package standalone

import (
	"github.com/rupixnet/rupixd/infrastructure/logger"
	"github.com/rupixnet/rupixd/util/panics"
)

var log = logger.RegisterSubSystem("NTAR")
var spawn = panics.GoroutineWrapperFunc(log)

