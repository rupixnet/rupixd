package addresslevelstore

import (
    "encoding/binary"
    "github.com/rupixnet/rupixd/domain/consensus/model"
    "github.com/rupixnet/rupixd/util/staging"
    "sync"
)

var bucketName = []byte("address-levels")

// addressLevelStore almacena el nivel L1-L5 de cada direccion
type AddressLevelStore struct {
    shardID model.StagingShardID
    cache   map[string]uint8
    mu      sync.RWMutex
    bucket  model.DBBucket
}

// New crea una nueva instancia del AddressLevelStore
func New(prefixBucket model.DBBucket, cacheSize int, preallocate bool) *AddressLevelStore {
    var cache map[string]uint8
    if preallocate {
        cache = make(map[string]uint8, cacheSize)
    } else {
        cache = make(map[string]uint8)
    }
    return &AddressLevelStore{
        shardID: staging.GenerateShardingID(),
        cache:   cache,
        bucket:  prefixBucket.Bucket(bucketName),
    }
}

// Stage guarda el nivel de una direccion en cache
func (als *AddressLevelStore) Stage(address string, level uint8) {
    als.mu.Lock()
    defer als.mu.Unlock()
    als.cache[address] = level
}

// Get obtiene el nivel de una direccion
func (als *AddressLevelStore) Get(dbContext model.DBReader, address string) (uint8, error) {
    als.mu.RLock()
    if level, ok := als.cache[address]; ok {
        als.mu.RUnlock()
        return level, nil
    }
    als.mu.RUnlock()

    key := als.bucket.Key([]byte(address))
    levelBytes, err := dbContext.Get(key)
    if err != nil {
        return 1, nil // default L1
    }
    level := uint8(binary.LittleEndian.Uint16(levelBytes))

    als.mu.Lock()
    als.cache[address] = level
    als.mu.Unlock()

    return level, nil
}

// Set guarda el nivel de una direccion en la DB
func (als *AddressLevelStore) Set(dbContext model.DBWriter, address string, level uint8) error {
    key := als.bucket.Key([]byte(address))
    levelBytes := make([]byte, 2)
    binary.LittleEndian.PutUint16(levelBytes, uint16(level))

    als.mu.Lock()
    als.cache[address] = level
    als.mu.Unlock()

    return dbContext.Put(key, levelBytes)
}

// Exists verifica si una direccion tiene nivel registrado
func (als *AddressLevelStore) Exists(dbContext model.DBReader, address string) (bool, error) {
    als.mu.RLock()
    _, ok := als.cache[address]
    als.mu.RUnlock()
    if ok {
        return true, nil
    }
    key := als.bucket.Key([]byte(address))
    return dbContext.Has(key)
}
