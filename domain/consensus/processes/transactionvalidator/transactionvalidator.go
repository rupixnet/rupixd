package transactionvalidator

import (
	"github.com/rupixnet/rupixd/domain/consensus/model"
	"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
	"github.com/rupixnet/rupixd/domain/consensus/utils/txscript"
	"github.com/rupixnet/rupixd/util/txmass"
)

const sigCacheSize = 10_000

type transactionValidator struct {
	blockCoinbaseMaturity                   uint64
	databaseContext                         model.DBReader
	pastMedianTimeManager                   model.PastMedianTimeManager
	ghostdagDataStore                       model.GHOSTDAGDataStore
	daaBlocksStore                          model.DAABlocksStore
	enableNonNativeSubnetworks              bool
	maxCoinbasePayloadLength                uint64
	ghostdagK                               externalapi.KType
	coinbasePayloadScriptPublicKeyMaxLength uint8
	sigCache                                *txscript.SigCache
	sigCacheECDSA                           *txscript.SigCacheECDSA
	txMassCalculator                        *txmass.Calculator
	// Rupix: unlock scores por red
	levelDiamanteUnlockScore uint64
	levelPlatinoUnlockScore  uint64
	levelRodioUnlockScore    uint64
	levelKingsUnlockScore    uint64
}

func New(blockCoinbaseMaturity uint64,
	enableNonNativeSubnetworks bool,
	maxCoinbasePayloadLength uint64,
	ghostdagK externalapi.KType,
	coinbasePayloadScriptPublicKeyMaxLength uint8,
	databaseContext model.DBReader,
	pastMedianTimeManager model.PastMedianTimeManager,
	ghostdagDataStore model.GHOSTDAGDataStore,
	daaBlocksStore model.DAABlocksStore,
	txMassCalculator *txmass.Calculator,
	levelDiamanteUnlockScore uint64,
	levelPlatinoUnlockScore uint64,
	levelRodioUnlockScore uint64,
	levelKingsUnlockScore uint64) model.TransactionValidator {
	return &transactionValidator{
		blockCoinbaseMaturity:                   blockCoinbaseMaturity,
		enableNonNativeSubnetworks:              enableNonNativeSubnetworks,
		maxCoinbasePayloadLength:                maxCoinbasePayloadLength,
		ghostdagK:                               ghostdagK,
		coinbasePayloadScriptPublicKeyMaxLength: coinbasePayloadScriptPublicKeyMaxLength,
		databaseContext:                         databaseContext,
		pastMedianTimeManager:                   pastMedianTimeManager,
		ghostdagDataStore:                       ghostdagDataStore,
		daaBlocksStore:                          daaBlocksStore,
		sigCache:                                txscript.NewSigCache(sigCacheSize),
		sigCacheECDSA:                           txscript.NewSigCacheECDSA(sigCacheSize),
		txMassCalculator:                        txMassCalculator,
		levelDiamanteUnlockScore:                levelDiamanteUnlockScore,
		levelPlatinoUnlockScore:                 levelPlatinoUnlockScore,
		levelRodioUnlockScore:                   levelRodioUnlockScore,
		levelKingsUnlockScore:                   levelKingsUnlockScore,
	}
}