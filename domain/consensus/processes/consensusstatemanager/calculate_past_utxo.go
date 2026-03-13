package consensusstatemanager

import (

	"github.com/rupixnet/rupixd/domain/consensus/utils/consensushashing"
	"github.com/rupixnet/rupixd/domain/consensus/utils/utxo"
	"github.com/rupixnet/rupixd/infrastructure/db/database"
	"github.com/rupixnet/rupixd/infrastructure/logger"
	"github.com/pkg/errors"

	"github.com/rupixnet/rupixd/domain/consensus/model"
	"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
	"github.com/rupixnet/rupixd/domain/consensus/ruleerrors"
	"github.com/rupixnet/rupixd/domain/consensus/utils/transactionhelper"
)

func (csm *consensusStateManager) CalculatePastUTXOAndAcceptanceData(stagingArea *model.StagingArea,
	blockHash *externalapi.DomainHash) (externalapi.UTXODiff, externalapi.AcceptanceData, model.Multiset, error) {

	onEnd := logger.LogAndMeasureExecutionTime(log, "CalculatePastUTXOAndAcceptanceData")
	defer onEnd()


	if blockHash.Equal(csm.genesisHash) {
    return csm.calculatePastUTXOAndAcceptanceDataWithSelectedParentUTXO(stagingArea, blockHash, utxo.NewUTXODiff())
}

	blockGHOSTDAGData, err := csm.ghostdagDataStore.Get(csm.databaseContext, stagingArea, blockHash, false)
	if err != nil {
		return nil, nil, nil, err
	}

	selectedParent := blockGHOSTDAGData.SelectedParent()
    if selectedParent == nil || selectedParent.Equal(model.VirtualGenesisBlockHash) {
        return csm.calculatePastUTXOAndAcceptanceDataWithSelectedParentUTXO(stagingArea, blockHash, utxo.NewUTXODiff())
    }
    selectedParentPastUTXO, err := csm.restorePastUTXO(stagingArea, selectedParent)
	if err != nil {
		return nil, nil, nil, err
	}

	return csm.calculatePastUTXOAndAcceptanceDataWithSelectedParentUTXO(stagingArea, blockHash, selectedParentPastUTXO)
}

func (csm *consensusStateManager) calculatePastUTXOAndAcceptanceDataWithSelectedParentUTXO(stagingArea *model.StagingArea,
	blockHash *externalapi.DomainHash, selectedParentPastUTXO externalapi.UTXODiff) (
	externalapi.UTXODiff, externalapi.AcceptanceData, model.Multiset, error) {

	blockGHOSTDAGData, err := csm.ghostdagDataStore.Get(csm.databaseContext, stagingArea, blockHash, false)
	if err != nil {
		return nil, nil, nil, err
	}

	daaScore, err := csm.daaBlocksStore.DAAScore(csm.databaseContext, stagingArea, blockHash)
	if err != nil {
		if database.IsNotFoundError(err) {
			daaScore = 0
		} else {
			return nil, nil, nil, err
		}
	}

	acceptanceData, utxoDiff, err := csm.applyMergeSetBlocks(stagingArea, blockHash, selectedParentPastUTXO, daaScore)
	if err != nil {
		return nil, nil, nil, err
	}

	multiset, err := csm.calculateMultiset(stagingArea, blockHash, acceptanceData, blockGHOSTDAGData, daaScore)
	if err != nil {
		return nil, nil, nil, err
	}

	return utxoDiff.ToImmutable(), acceptanceData, multiset, nil
}

func (csm *consensusStateManager) restorePastUTXO(
	stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (externalapi.UTXODiff, error) {

	onEnd := logger.LogAndMeasureExecutionTime(log, "restorePastUTXO")
	defer onEnd()

	var utxoDiffs []externalapi.UTXODiff
	nextBlockHash := blockHash
	for {
		// Guard: if hash is nil, no more chain to traverse
		if nextBlockHash == nil {
			break
		}

		utxoDiff, err := csm.utxoDiffStore.UTXODiff(csm.databaseContext, stagingArea, nextBlockHash)
		if err != nil {
			if database.IsNotFoundError(err) {
				// No UTXODiff in DB for this block - end of known chain
				break
			}
			return nil, err
		}
		utxoDiffs = append(utxoDiffs, utxoDiff)

		exists, err := csm.utxoDiffStore.HasUTXODiffChild(csm.databaseContext, stagingArea, nextBlockHash)
		if err != nil {
			return nil, err
		}
		if !exists {
			break
		}

		nextBlockHash, err = csm.utxoDiffStore.UTXODiffChild(csm.databaseContext, stagingArea, nextBlockHash)
		if err != nil {
			return nil, err
		}
		if nextBlockHash == nil {
			break
		}
	}

	accumulatedDiff := utxo.NewMutableUTXODiff()
	for i := len(utxoDiffs) - 1; i >= 0; i-- {
		err := accumulatedDiff.WithDiffInPlace(utxoDiffs[i])
		if err != nil {
			return nil, err
		}
	}

	return accumulatedDiff.ToImmutable(), nil
}

func (csm *consensusStateManager) applyMergeSetBlocks(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash,
	selectedParentPastUTXODiff externalapi.UTXODiff, daaScore uint64) (
	externalapi.AcceptanceData, externalapi.MutableUTXODiff, error) {

	mergeSetHashes, err := csm.ghostdagManager.GetSortedMergeSet(stagingArea, blockHash)
	if err != nil {
		return nil, nil, err
	}

	mergeSetBlocks, err := csm.blockStore.Blocks(csm.databaseContext, stagingArea, mergeSetHashes)
	if err != nil {
		return nil, nil, err
	}

	selectedParentMedianTime, err := csm.pastMedianTimeManager.PastMedianTime(stagingArea, blockHash)
	if err != nil {
		return nil, nil, err
	}

	multiblockAcceptanceData := make(externalapi.AcceptanceData, len(mergeSetBlocks))
	if selectedParentPastUTXODiff == nil {
		selectedParentPastUTXODiff = utxo.NewUTXODiff()
	}
	accumulatedUTXODiff := selectedParentPastUTXODiff.CloneMutable()
	accumulatedMass := uint64(0)

	for i, mergeSetBlock := range mergeSetBlocks {
		mergeSetBlockHash := consensushashing.BlockHash(mergeSetBlock)
		blockAcceptanceData := &externalapi.BlockAcceptanceData{
			BlockHash:                 mergeSetBlockHash,
			TransactionAcceptanceData: make([]*externalapi.TransactionAcceptanceData, len(mergeSetBlock.Transactions)),
		}
		isSelectedParent := i == 0

		for j, transaction := range mergeSetBlock.Transactions {
			var isAccepted bool
			transactionID := consensushashing.TransactionID(transaction)

			isAccepted, accumulatedMass, err = csm.maybeAcceptTransaction(stagingArea, transaction, blockHash,
				isSelectedParent, accumulatedUTXODiff, accumulatedMass, selectedParentMedianTime, daaScore)
			if err != nil {
				return nil, nil, err
			}

			var transactionInputUTXOEntries []externalapi.UTXOEntry
			if isAccepted {
				transactionInputUTXOEntries = make([]externalapi.UTXOEntry, len(transaction.Inputs))
				for k, input := range transaction.Inputs {
					transactionInputUTXOEntries[k] = input.UTXOEntry
				}
			}

			blockAcceptanceData.TransactionAcceptanceData[j] = &externalapi.TransactionAcceptanceData{
				Transaction:                 transaction,
				Fee:                         transaction.Fee,
				IsAccepted:                  isAccepted,
				TransactionInputUTXOEntries: transactionInputUTXOEntries,
			}
			_ = transactionID
		}
		multiblockAcceptanceData[i] = blockAcceptanceData
	}

	return multiblockAcceptanceData, accumulatedUTXODiff, nil
}

func (csm *consensusStateManager) maybeAcceptTransaction(stagingArea *model.StagingArea,
	transaction *externalapi.DomainTransaction, blockHash *externalapi.DomainHash, isSelectedParent bool,
	accumulatedUTXODiff externalapi.MutableUTXODiff, accumulatedMassBefore uint64, selectedParentPastMedianTime int64,
	blockDAAScore uint64) (isAccepted bool, accumulatedMassAfter uint64, err error) {

	err = csm.populateTransactionWithUTXOEntriesFromVirtualOrDiff(stagingArea, transaction, accumulatedUTXODiff.ToImmutable())
	if err != nil {
		if !errors.As(err, &(ruleerrors.RuleError{})) {
			return false, 0, err
		}
		return false, accumulatedMassBefore, nil
	}

	if transactionhelper.IsCoinBase(transaction) {
		if !isSelectedParent {
			return false, accumulatedMassBefore, nil
		}
	} else {
		err = csm.transactionValidator.ValidateTransactionInContextAndPopulateFee(
			stagingArea, transaction, blockHash)
		if err != nil {
			if !errors.As(err, &(ruleerrors.RuleError{})) {
				return false, 0, err
			}
			return false, accumulatedMassBefore, nil
		}
	}

	err = accumulatedUTXODiff.AddTransaction(transaction, blockDAAScore)
	if err != nil {
		return false, 0, err
	}

	return true, accumulatedMassAfter, nil
}

func (csm *consensusStateManager) RestorePastUTXOSetIterator(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (
	externalapi.ReadOnlyUTXOSetIterator, error) {

	onEnd := logger.LogAndMeasureExecutionTime(log, "RestorePastUTXOSetIterator")
	defer onEnd()

	blockStatus, _, err := csm.resolveBlockStatus(stagingArea, blockHash, true)
	if err != nil {
		return nil, err
	}
	if blockStatus != externalapi.StatusUTXOValid {
		return nil, errors.Errorf(
			"block %s, has status '%s', and therefore can't restore it's UTXO set. Only blocks with status '%s' can be restored.",
			blockHash, blockStatus, externalapi.StatusUTXOValid)
	}

	blockDiff, err := csm.restorePastUTXO(stagingArea, blockHash)
	if err != nil {
		return nil, err
	}

	virtualUTXOSetIterator, err := csm.consensusStateStore.VirtualUTXOSetIterator(csm.databaseContext, stagingArea)
	if err != nil {
		return nil, err
	}

	return utxo.IteratorWithDiff(virtualUTXOSetIterator, blockDiff)
}







