package kingscountstore

import (
"encoding/binary"

"github.com/rupixnet/rupixd/domain/consensus/model"
"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
"github.com/rupixnet/rupixd/domain/consensus/utils/lrucache"
"github.com/rupixnet/rupixd/util/staging"
)

var bucketName = []byte("kings-count")

// kingsCountStore guarda por bloque cuantos Kings existen en su UTXO set.
// Clon estructural de multisetStore: mismo shard, mismo cache, mismo bucket;
// el payload es un uint64 en 8 bytes big-endian.
type kingsCountStore struct {
shardID model.StagingShardID
cache   *lrucache.LRUCache
bucket  model.DBBucket
}

// New instantiates a new KingsCountStore
func New(prefixBucket model.DBBucket, cacheSize int, preallocate bool) model.KingsCountStore {
return &kingsCountStore{
shardID: staging.GenerateShardingID(),
cache:   lrucache.New(cacheSize, preallocate),
bucket:  prefixBucket.Bucket(bucketName),
}
}

// Stage stages the given count for the given blockHash
func (kcs *kingsCountStore) Stage(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash, count uint64) {
stagingShard := kcs.stagingShard(stagingArea)
stagingShard.toAdd[*blockHash] = count
}

func (kcs *kingsCountStore) IsStaged(stagingArea *model.StagingArea) bool {
return kcs.stagingShard(stagingArea).isStaged()
}

// Get gets the count associated with the given blockHash
func (kcs *kingsCountStore) Get(dbContext model.DBReader, stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (uint64, error) {
stagingShard := kcs.stagingShard(stagingArea)

if count, ok := stagingShard.toAdd[*blockHash]; ok {
return count, nil
}
if count, ok := kcs.cache.Get(blockHash); ok {
return count.(uint64), nil
}

countBytes, err := dbContext.Get(kcs.hashAsKey(blockHash))
if err != nil {
return 0, err
}
count := deserializeCount(countBytes)
kcs.cache.Add(blockHash, count)
return count, nil
}

// Delete deletes the count associated with the given blockHash
func (kcs *kingsCountStore) Delete(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) {
stagingShard := kcs.stagingShard(stagingArea)

if _, ok := stagingShard.toAdd[*blockHash]; ok {
delete(stagingShard.toAdd, *blockHash)
return
}
stagingShard.toDelete[*blockHash] = struct{}{}
}

func (kcs *kingsCountStore) hashAsKey(hash *externalapi.DomainHash) model.DBKey {
return kcs.bucket.Key(hash.ByteSlice())
}

func serializeCount(count uint64) []byte {
b := make([]byte, 8)
binary.BigEndian.PutUint64(b, count)
return b
}

func deserializeCount(b []byte) uint64 {
return binary.BigEndian.Uint64(b)
}
