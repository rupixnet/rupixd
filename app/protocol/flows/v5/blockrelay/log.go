package blockrelay

import (
	"github.com/rupixnet/rupixd/infrastructure/logger"
	"github.com/rupixnet/rupixd/util/panics"
)

var log = logger.RegisterSubSystem("PROT")
var spawn = panics.GoroutineWrapperFunc(log)

