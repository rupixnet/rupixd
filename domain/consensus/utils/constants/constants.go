package constants

import "math"

const (
    // BlockVersion represents the current block version
    BlockVersion uint16 = 1

    // MaxTransactionVersion is the current latest supported transaction version.
    MaxTransactionVersion uint16 = 0

    // MaxScriptPublicKeyVersion is the current latest supported public key script version.
    MaxScriptPublicKeyVersion uint16 = 0

    // RupiaPerRupix is the number of rupia in one rupix (1 RUPIX).
    RupiaPerRupix = 100_000_000

    // MaxRupia is the maximum transaction amount allowed in rupia.
    MaxRupia = uint64(42_000_000 * RupiaPerRupix)

    // MaxTxInSequenceNum is the maximum sequence number the sequence field
    // of a transaction input can be.
    MaxTxInSequenceNum uint64 = math.MaxUint64

    // SequenceLockTimeDisabled is a flag that if set on a transaction
    // input's sequence number, the sequence number will not be interpreted
    // as a relative locktime.
    SequenceLockTimeDisabled uint64 = 1 << 63

    // SequenceLockTimeMask is a mask that extracts the relative locktime
    // when masked against the transaction input sequence number.
    SequenceLockTimeMask uint64 = 0x00000000ffffffff

    // LockTimeThreshold is the number below which a lock time is
    // interpreted to be a DAA score.
    LockTimeThreshold = 5e11

    // UnacceptedDAAScore is used to for UTXOEntries that were created by
    // transactions in the mempool, or otherwise not-yet-accepted transactions.
    UnacceptedDAAScore = math.MaxUint64

    // ========================================================
    // RUPIX ANTI-SPAM PROTOCOL v1.0
    // Protecciones grabadas en el protocolo desde el dia 1
    // ========================================================

    // MaxTxPayloadSize tamano maximo del Payload en bytes
    MaxTxPayloadSize = 100

    // MaxTxOutputs maximo de outputs por transaccion
    MaxTxOutputs = 256

    // MaxTxInputs maximo de inputs por transaccion
    MaxTxInputs = 1000

    // MinTxFeePerByte fee minimo en rupias por byte
    MinTxFeePerByte = uint64(1)

    // MaxBlockMassLimit masa maxima permitida por bloque
    MaxBlockMassLimit = uint64(500_000)

    // TxTypeNormal transferencia normal de RUPIX
    TxTypeNormal = uint8(0)

    // TxTypeCoinbase recompensa del minero
    TxTypeCoinbase = uint8(1)

    // TxTypeBurn quema de tokens para subir de nivel
    TxTypeBurn = uint8(2)

    // TxTypeInvalid cualquier TX no reconocida
    TxTypeInvalid = uint8(255)
)
