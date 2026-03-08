package burnmanager

import (
    "github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
    "github.com/rupixnet/rupixd/domain/consensus/ruleerrors"
    "github.com/pkg/errors"
)

// Niveles de Rupix - sistema de escasez progresiva
const (
    TokenLevelGold     = uint8(1) // Rupix Oro      L1 - minado
    TokenLevelDiamond  = uint8(2) // Rupix Diamante L2 - quemar 10 L1
    TokenLevelPlatinum = uint8(3) // Rupix Platino  L3 - quemar 10 L2
    TokenLevelRhodium  = uint8(4) // Rupix Rodio    L4 - quemar 10 L3
    TokenLevelKings    = uint8(5) // Kings Rupix    L5 - quemar 10 L4

    BurnRatio = uint64(10)
    MaxLevel  = TokenLevelKings

    BurnAddress = "rupix:qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq"
)

var MaxSupplyByLevel = map[uint8]uint64{
    TokenLevelGold:     42_000_000,
    TokenLevelDiamond:  2_100_000,
    TokenLevelPlatinum: 210_000,
    TokenLevelRhodium:  21_000,
    TokenLevelKings:    2_100,
}

var LevelNames = map[uint8]string{
    TokenLevelGold:     "Rupix Gold L1",
    TokenLevelDiamond:  "Rupix Diamante L2",
    TokenLevelPlatinum: "Rupix Platino L3",
    TokenLevelRhodium:  "Rupix Rodio L4",
    TokenLevelKings:    "Kings Rupix L5",
}

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
        return errors.Wrapf(ruleerrors.ErrNoTxInputs, "burn transaction cannot be coinbase")
    }
    if len(tx.Outputs) == 0 {
        return errors.Wrapf(ruleerrors.ErrNoTxInputs, "burn transaction has no outputs")
    }
    if len(tx.Payload) < 2 {
        return errors.Wrapf(ruleerrors.ErrInvalidPayload, "burn transaction payload too small: need sourceLevel and targetLevel")
    }
    sourceLevel := tx.Payload[0]
    targetLevel := tx.Payload[1]
    if targetLevel != sourceLevel+1 {
        return errors.Wrapf(ruleerrors.ErrInvalidPayload, "invalid level transition: can only go from level %d to %d, got %d", sourceLevel, sourceLevel+1, targetLevel)
    }
    if targetLevel > TokenLevelKings {
        return errors.Wrapf(ruleerrors.ErrInvalidPayload, "invalid target level %d: maximum level is %d", targetLevel, TokenLevelKings)
    }
    if sourceLevel < TokenLevelGold {
        return errors.Wrapf(ruleerrors.ErrInvalidPayload, "invalid source level %d: minimum level is %d", sourceLevel, TokenLevelGold)
    }
    var totalBurned uint64
    for _, out := range tx.Outputs {
        totalBurned += out.Value
    }
    if totalBurned < BurnRatio {
        return errors.Wrapf(ruleerrors.ErrInvalidPayload, "burn amount %d is less than required %d for level transition %d->%d", totalBurned, BurnRatio, sourceLevel, targetLevel)
    }
    return nil
}

func (bm *burnManager) ProcessBurnTransaction(tx *externalapi.DomainTransaction) (*BurnResult, error) {
    if err := bm.ValidateBurnTransaction(tx); err != nil {
        return nil, err
    }
    fromLevel := tx.Payload[0]
    toLevel := tx.Payload[1]
    currentMinted := bm.stats.TotalMinted[toLevel]
    maxSupply := MaxSupplyByLevel[toLevel]
    if currentMinted >= maxSupply {
        return nil, errors.Errorf("maximum supply for level %d (%s) reached: %d/%d", toLevel, LevelNames[toLevel], currentMinted, maxSupply)
    }
    bm.stats.TotalBurned[fromLevel] += BurnRatio
    bm.stats.TotalMinted[toLevel]++
    bm.stats.CurrentSupply[fromLevel] -= BurnRatio
    bm.stats.CurrentSupply[toLevel]++
    if fromLevel == TokenLevelGold {
        bm.stats.TotalRupixBurned += BurnRatio
        bm.stats.CirculatingL1 -= BurnRatio
        if bm.stats.TotalRupixMined > 0 {
            bm.stats.PercentageBurned = float64(bm.stats.TotalRupixBurned) / float64(bm.stats.TotalRupixMined) * 100
        }
    }
    return &BurnResult{BurnTx: tx, ToLevel: toLevel, MintedAmount: 1}, nil
}

func (bm *burnManager) GetBurnStats() (*BurnStats, error) {
    return bm.stats, nil
}
