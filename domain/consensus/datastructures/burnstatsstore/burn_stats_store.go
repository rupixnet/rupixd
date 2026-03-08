package burnstatsstore

import (
	"encoding/json"

	"github.com/rupixnet/rupixd/domain/consensus/model"
)

var bucketName = []byte("burn-stats")

type BurnStatsStore struct {
	bucket model.DBBucket
}

type PersistedBurnStats struct {
	TotalBurned      map[uint8]uint64
	TotalMinted      map[uint8]uint64
	CurrentSupply    map[uint8]uint64
	TotalRupixBurned uint64
	TotalRupixMined  uint64
	CirculatingL1    uint64
	PercentageBurned float64
}

func New(prefixBucket model.DBBucket) *BurnStatsStore {
	return &BurnStatsStore{
		bucket: prefixBucket.Bucket(bucketName),
	}
}

var statsKey = []byte("global")

func (s *BurnStatsStore) Save(dbContext model.DBWriter, stats *PersistedBurnStats) error {
	data, err := json.Marshal(stats)
	if err != nil {
		return err
	}
	return dbContext.Put(s.bucket.Key(statsKey), data)
}

func (s *BurnStatsStore) Load(dbContext model.DBReader) (*PersistedBurnStats, error) {
	data, err := dbContext.Get(s.bucket.Key(statsKey))
	if err != nil {
		// Primera vez — devolver stats vacios
		return &PersistedBurnStats{
			TotalBurned:   make(map[uint8]uint64),
			TotalMinted:   make(map[uint8]uint64),
			CurrentSupply: make(map[uint8]uint64),
		}, nil
	}
	var stats PersistedBurnStats
	if err := json.Unmarshal(data, &stats); err != nil {
		return nil, err
	}
	return &stats, nil
}