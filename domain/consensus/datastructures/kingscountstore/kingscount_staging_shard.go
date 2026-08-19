package kingscountstore

import (
"github.com/rupixnet/rupixd/domain/consensus/model"
"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
)

type kingsCountStagingShard struct {
store    *kingsCountStore
toAdd    map[externalapi.DomainHash]uint64
toDelete map[externalapi.DomainHash]struct{}
}

func (kcs *kingsCountStore) stagingShard(stagingArea *model.StagingArea) *kingsCountStagingShard {
return stagingArea.GetOrCreateShard(kcs.shardID, func() model.StagingShard {
return &kingsCountStagingShard{
store:    kcs,
toAdd:    make(map[externalapi.DomainHash]uint64),
toDelete: make(map[externalapi.DomainHash]struct{}),
}
}).(*kingsCountStagingShard)
}

func (kcss *kingsCountStagingShard) Commit(dbTx model.DBTransaction) error {
for hash, count := range kcss.toAdd {
err := dbTx.Put(kcss.store.hashAsKey(&hash), serializeCount(count))
if err != nil {
return err
}
kcss.store.cache.Add(&hash, count)
}
for hash := range kcss.toDelete {
err := dbTx.Delete(kcss.store.hashAsKey(&hash))
if err != nil {
return err
}
kcss.store.cache.Remove(&hash)
}
return nil
}

func (kcss *kingsCountStagingShard) isStaged() bool {
return len(kcss.toAdd) != 0 || len(kcss.toDelete) != 0
}
