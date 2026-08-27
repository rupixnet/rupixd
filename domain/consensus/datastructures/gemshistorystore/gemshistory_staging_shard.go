package gemshistorystore

import (
"github.com/rupixnet/rupixd/domain/consensus/model"
"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
)

type gemsHistoryStagingShard struct {
store    *gemsHistoryStore
toAdd    map[externalapi.DomainHash]*externalapi.GemsHistory
toDelete map[externalapi.DomainHash]struct{}
}

func (ghs *gemsHistoryStore) stagingShard(stagingArea *model.StagingArea) *gemsHistoryStagingShard {
return stagingArea.GetOrCreateShard(ghs.shardID, func() model.StagingShard {
return &gemsHistoryStagingShard{
store:    ghs,
toAdd:    make(map[externalapi.DomainHash]*externalapi.GemsHistory),
toDelete: make(map[externalapi.DomainHash]struct{}),
}
}).(*gemsHistoryStagingShard)
}

func (ghss *gemsHistoryStagingShard) Commit(dbTx model.DBTransaction) error {
for hash, history := range ghss.toAdd {
err := dbTx.Put(ghss.store.hashAsKey(&hash), serializeHistory(history))
if err != nil {
return err
}
ghss.store.cache.Add(&hash, history)
}
for hash := range ghss.toDelete {
err := dbTx.Delete(ghss.store.hashAsKey(&hash))
if err != nil {
return err
}
ghss.store.cache.Remove(&hash)
}
return nil
}

func (ghss *gemsHistoryStagingShard) isStaged() bool {
return len(ghss.toAdd) != 0 || len(ghss.toDelete) != 0
}
