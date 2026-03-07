package addresslevelstore

import (
	"encoding/binary"
	"github.com/rupixnet/rupixd/domain/consensus/model"
	"github.com/rupixnet/rupixd/domain/consensus/utils/lrucache"
	"github.com/rupixnet/rupixd/util/staging"
)

var bucketName = []byte("address-levels")

// addressLevelStore almacena el nivel L1-L5 de cada direccion
type addressLevelStore struct {
	shardID model.StagingShardID
	cache   *lrucache.LRUCache
	bucket  model.DBBucket
}

// New crea una nueva instancia del AddressLevelStore
func New(prefixBucket model.DBBucket, cacheSize int, preallocate bool) *addressLevelStore {
	return &addressLevelStore{
		shardID: staging.GenerateShardingID(),
		cache:   lrucache.New(cacheSize, preallocate),
		bucket:  prefixBucket.Bucket(bucketName),
	}
}

// Stage guarda el nivel de una direccion en staging
func (als *addressLevelStore) Stage(address string, level uint8) {
	als.cache.Add(address, level)
}

// Get obtiene el nivel de una direccion
func (als *addressLevelStore) Get(dbContext model.DBReader, address string) (uint8, error) {
	// Buscar en cache primero
	if level, ok := als.cache.Get(address); ok {
		return level.(uint8), nil
	}

	// Buscar en DB
	key := als.bucket.Key([]byte(address))
	levelBytes, err := dbContext.Get(key)
	if err != nil {
		return 1, nil // default L1
	}

	level := uint8(binary.LittleEndian.Uint16(levelBytes))
	als.cache.Add(address, level)
	return level, nil
}

// Set guarda el nivel de una direccion en la DB
func (als *addressLevelStore) Set(dbContext model.DBWriter, address string, level uint8) error {
	key := als.bucket.Key([]byte(address))
	levelBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(levelBytes, uint16(level))
	als.cache.Add(address, level)
	return dbContext.Put(key, levelBytes)
}

// Exists verifica si una direccion tiene nivel registrado
func (als *addressLevelStore) Exists(dbContext model.DBReader, address string) (bool, error) {
	if als.cache.Has(address) {
		return true, nil
	}
	key := als.bucket.Key([]byte(address))
	return dbContext.Has(key)
}