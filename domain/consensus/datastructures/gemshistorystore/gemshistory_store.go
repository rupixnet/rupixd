package gemshistorystore

import (
"encoding/binary"

"github.com/rupixnet/rupixd/domain/consensus/model"
"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
"github.com/rupixnet/rupixd/domain/consensus/utils/lrucache"
"github.com/rupixnet/rupixd/util/staging"
)

var bucketName = []byte("gems-history")

// gemsHistoryStore guarda por bloque los conteos historicos de gemas
// (Diamante, Platino, Rodio). Clon estructural del kingsCountStore; el payload
// son 3 uint64 en 24 bytes big-endian.
type gemsHistoryStore struct {
shardID model.StagingShardID
cache   *lrucache.LRUCache
bucket  model.DBBucket
}

// New instantiates a new GemsHistoryStore
func New(prefixBucket model.DBBucket, cacheSize int, preallocate bool) model.GemsHistoryStore {
return &gemsHistoryStore{
shardID: staging.GenerateShardingID(),
cache:   lrucache.New(cacheSize, preallocate),
bucket:  prefixBucket.Bucket(bucketName),
}
}

// Stage stages the given history for the given blockHash
func (ghs *gemsHistoryStore) Stage(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash, history *externalapi.GemsHistory) {
stagingShard := ghs.stagingShard(stagingArea)
stagingShard.toAdd[*blockHash] = history.Clone()
}

func (ghs *gemsHistoryStore) IsStaged(stagingArea *model.StagingArea) bool {
return ghs.stagingShard(stagingArea).isStaged()
}

// Get gets the history associated with the given blockHash
func (ghs *gemsHistoryStore) Get(dbContext model.DBReader, stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (*externalapi.GemsHistory, error) {
stagingShard := ghs.stagingShard(stagingArea)
if history, ok := stagingShard.toAdd[*blockHash]; ok {
return history.Clone(), nil
}
if history, ok := ghs.cache.Get(blockHash); ok {
return history.(*externalapi.GemsHistory).Clone(), nil
}
historyBytes, err := dbContext.Get(ghs.hashAsKey(blockHash))
if err != nil {
return nil, err
}
history := deserializeHistory(historyBytes)
ghs.cache.Add(blockHash, history)
return history.Clone(), nil
}

// Delete deletes the history associated with the given blockHash
func (ghs *gemsHistoryStore) Delete(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) {
stagingShard := ghs.stagingShard(stagingArea)
if _, ok := stagingShard.toAdd[*blockHash]; ok {
delete(stagingShard.toAdd, *blockHash)
return
}
stagingShard.toDelete[*blockHash] = struct{}{}
}

func (ghs *gemsHistoryStore) hashAsKey(hash *externalapi.DomainHash) model.DBKey {
return ghs.bucket.Key(hash.ByteSlice())
}

func serializeHistory(history *externalapi.GemsHistory) []byte {
b := make([]byte, 24)
binary.BigEndian.PutUint64(b[0:8], history.Diamante)
binary.BigEndian.PutUint64(b[8:16], history.Platino)
binary.BigEndian.PutUint64(b[16:24], history.Rodio)
return b
}

func deserializeHistory(b []byte) *externalapi.GemsHistory {
return &externalapi.GemsHistory{
Diamante: binary.BigEndian.Uint64(b[0:8]),
Platino:  binary.BigEndian.Uint64(b[8:16]),
Rodio:    binary.BigEndian.Uint64(b[16:24]),
}
}
