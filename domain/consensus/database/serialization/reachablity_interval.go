package serialization

import (
	"github.com/rupixnet/rupixd/domain/consensus/model"
)

func reachablityIntervalToDBReachablityInterval(reachabilityInterval *model.ReachabilityInterval) *DbReachabilityInterval {
	if reachabilityInterval == nil {
		return &DbReachabilityInterval{Start: 0, End: 0}
	}
	return &DbReachabilityInterval{
		Start: reachabilityInterval.Start,
		End:   reachabilityInterval.End,
	}
}

func dbReachablityIntervalToReachablityInterval(dbReachabilityInterval *DbReachabilityInterval) *model.ReachabilityInterval {
	return &model.ReachabilityInterval{
		Start: dbReachabilityInterval.Start,
		End:   dbReachabilityInterval.End,
	}
}

