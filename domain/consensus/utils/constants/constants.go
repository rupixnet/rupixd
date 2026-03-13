package constants

import "math"

const (
    BlockVersion uint16 = 1
    MaxTransactionVersion uint16 = 0
    MaxScriptPublicKeyVersion uint16 = 0
    RupiaPerRupix = 100_000_000
    MaxRupia = uint64(42_000_000 * RupiaPerRupix)
    MaxTxInSequenceNum uint64 = math.MaxUint64
    SequenceLockTimeDisabled uint64 = 1 << 63
    SequenceLockTimeMask uint64 = 0x00000000ffffffff
    LockTimeThreshold = 5e11
    UnacceptedDAAScore = math.MaxUint64
    MinimumDifficultyTarget = "28269553036454149273332760011886696253239742350009903329945699220681916416"
    MinBurnPerTx uint64 = 1_000
    BurnPerByte uint64 = 10
    MaxTxPayloadSize = 100
    MaxTxOutputs = 256
    MaxTxInputs = 1000
    MinTxFeePerByte = uint64(1)
    MaxBlockMassLimit = uint64(500_000)
    TxTypeNormal = uint8(0)
    TxTypeCoinbase = uint8(1)
    TxTypeBurn = uint8(2)
    TxTypeInvalid = uint8(255)
    LevelOro uint16 = 0
    LevelDiamante uint16 = 1
    LevelPlatino uint16 = 2
    LevelRodio uint16 = 3
    LevelKings uint16 = 4
    LevelDiamanteUnlockScore uint64 = 42_000_000
    LevelPlatinoUnlockScore uint64 = 84_000_000
    LevelRodioUnlockScore uint64 = 126_000_000
    LevelKingsUnlockScore uint64 = 168_000_000
    MaxSupplyOro uint64 = 42_000_000 * RupiaPerRupix
    MaxSupplyDiamante uint64 = 2_100_000 * RupiaPerRupix
    MaxSupplyPlatino uint64 = 210_000 * RupiaPerRupix
    MaxSupplyRodio uint64 = 21_000 * RupiaPerRupix
    MaxSupplyKings uint64 = 2_100 * RupiaPerRupix
    BurnRatio uint64 = 10
)