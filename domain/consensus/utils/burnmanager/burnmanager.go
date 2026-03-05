package burnmanager

import (
    "github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
    "github.com/pkg/errors"
)

const (
    TokenLevelGold       = uint8(1)
    TokenLevelDiamond    = uint8(2)
    TokenLevelPlatinum   = uint8(3)
    TokenLevelRhodium    = uint8(4)
    TokenLevelCalifornio = uint8(5)
    BurnRatio            = uint64(10)
    MaxLevel             = TokenLevelCalifornio
    BurnAddress          = "rupix:burnrupixburnrupixburnrupixburnrupixburnrupix0"
)

var MaxSupplyByLevel = map[uint8]uint64{
    TokenLevelGold:       42_000_000,
    TokenLevelDiamond:    2_100_000,
    TokenLevelPlatinum:   210_000,
    TokenLevelRhodium:    21_000,
    TokenLevelCalifornio: 2_100,
}

var LevelNames = map[uint8]string{
    TokenLevelGold:       "Oro (Gold)",
    TokenLevelDiamond:    "Diamante (Diamond)",
    TokenLevelPlatinum:   "Platino (Platinum)",
    TokenLevelRhodium:    "Rodio (Rhodium)",
    TokenLevelCalifornio: "Californio (Californium)",
}

var BurnPayloadPrefix = []byte("RUPIX_BURN_V1")

type BurnResult struct {
    BurnTx           *externalapi.DomainTransaction
    ToLevel          uint8
    MintedAmount     uint64
    RecipientAddress string
}

type BurnStats struct {
    TotalMined       map[uint8]uint64
    TotalBurned      map[uint8]uint64
    TotalMinted      map[uint8]uint64
    CurrentSupply    map[uint8]uint64
    MaxSupply        map[uint8]uint64
    TotalRupixBurned uint64
    TotalRupixMined  uint64
    CirculatingL1    uint64
    PercentageBurned float64
}

type Manager interface {
    ValidateBurnTransaction(tx *externalapi.DomainTransaction) error
    ProcessBurnTransaction(tx *externalapi.DomainTransaction) (*BurnResult, error)
    GetBurnStats() (*BurnStats, error)
    IsBurnAddress(address string) bool
}

type burnManager struct {
    stats *BurnStats
}

func New() Manager {
    return &burnManager{
        stats: &BurnStats{
            TotalMined:    make(map[uint8]uint64),
            TotalBurned:   make(map[uint8]uint64),
            TotalMinted:   make(map[uint8]uint64),
            CurrentSupply: make(map[uint8]uint64),
            MaxSupply:     MaxSupplyByLevel,
        },
    }
}

func (bm *burnManager) IsBurnAddress(address string) bool {
    return address == BurnAddress
}

func (bm *burnManager) ValidateBurnTransaction(tx *externalapi.DomainTransaction) error {
    if len(tx.Inputs) == 0 {
        return errors.New("burn transaction cannot be coinbase")
    }
    if len(tx.Payload) < len(BurnPayloadPrefix) {
        return errors.New("burn transaction missing payload prefix")
    }
    for i, b := range BurnPayloadPrefix {
        if tx.Payload[i] != b {
            return errors.New("invalid burn transaction payload prefix")
        }
    }
    return nil
}

func (bm *burnManager) ProcessBurnTransaction(tx *externalapi.DomainTransaction) (*BurnResult, error) {
    if err := bm.ValidateBurnTransaction(tx); err != nil {
        return nil, err
    }
    if len(tx.Payload) < len(BurnPayloadPrefix)+1 {
        return nil, errors.New("burn payload too short")
    }
    fromLevel := tx.Payload[len(BurnPayloadPrefix)]
    if fromLevel < TokenLevelGold || fromLevel >= MaxLevel {
        return nil, errors.Errorf("invalid burn level %d", fromLevel)
    }
    toLevel := fromLevel + 1
    currentMinted := bm.stats.TotalMinted[toLevel]
    maxSupply := MaxSupplyByLevel[toLevel]
    if currentMinted >= maxSupply {
        return nil, errors.Errorf("maximum supply for level %d (%s) reached: %d/%d",
            toLevel, LevelNames[toLevel], currentMinted, maxSupply)
    }
    bm.stats.TotalBurned[fromLevel] += BurnRatio
    bm.stats.TotalMinted[toLevel]++
    bm.stats.CurrentSupply[fromLevel] -= BurnRatio
    bm.stats.CurrentSupply[toLevel]++
    if fromLevel == TokenLevelGold {
        bm.stats.TotalRupixBurned += BurnRatio
        bm.stats.CirculatingL1 -= BurnRatio
        if bm.stats.TotalRupixMined > 0 {
            bm.stats.PercentageBurned = float64(bm.stats.TotalRupixBurned) /
                float64(bm.stats.TotalRupixMined) * 100
        }
    }
    return &BurnResult{BurnTx: tx, ToLevel: toLevel, MintedAmount: 1}, nil
}

func (bm *burnManager) GetBurnStats() (*BurnStats, error) {
    return bm.stats, nil
}
