package consensusstatemanager

import (
	"github.com/rupixnet/rupixd/domain/consensus/model"
	"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
	"github.com/rupixnet/rupixd/domain/consensus/utils/consensushashing"
	"github.com/rupixnet/rupixd/domain/consensus/utils/multiset"
	"github.com/rupixnet/rupixd/domain/consensus/utils/utxo"
	"github.com/rupixnet/rupixd/infrastructure/db/database"
)

func (csm *consensusStateManager) calculateMultiset(stagingArea *model.StagingArea,
	blockHash *externalapi.DomainHash,
	acceptanceData externalapi.AcceptanceData,
	blockGHOSTDAGData *externalapi.BlockGHOSTDAGData,
	daaScore uint64) (model.Multiset, error) {

	if blockHash.Equal(csm.genesisHash) {
		return csm.multisetStore.Get(csm.databaseContext, stagingArea, blockHash)
	}

	// RUPIX FIX: selectedParent may be nil or VirtualGenesis on first startup
	var ms model.Multiset
	selectedParent := blockGHOSTDAGData.SelectedParent()
	if selectedParent == nil || selectedParent.Equal(model.VirtualGenesisBlockHash) {
		ms = multiset.New()
	} else {
		var err error
		ms, err = csm.multisetStore.Get(csm.databaseContext, stagingArea, selectedParent)
		if err != nil {
			if database.IsNotFoundError(err) {
				ms = multiset.New()
			} else {
				return nil, err
			}
		}
	}

	for _, blockAcceptanceData := range acceptanceData {
		for i, transactionAcceptanceData := range blockAcceptanceData.TransactionAcceptanceData {
			transaction := transactionAcceptanceData.Transaction
			transactionID := consensushashing.TransactionID(transaction)
			if !transactionAcceptanceData.IsAccepted {
				log.Tracef("Skipping transaction %s because it was not accepted", transactionID)
				continue
			}

			isCoinbase := i == 0
			err := addTransactionToMultiset(ms, transaction, daaScore, isCoinbase)
			if err != nil {
				return nil, err
			}
		}
	}

	return ms, nil
}

func addTransactionToMultiset(ms model.Multiset, transaction *externalapi.DomainTransaction,
	blockDAAScore uint64, isCoinbase bool) error {

	transactionID := consensushashing.TransactionID(transaction)
	log.Tracef("addTransactionToMultiset start for transaction %s", transactionID)
	defer log.Tracef("addTransactionToMultiset end for transaction %s", transactionID)

	for _, input := range transaction.Inputs {
		err := removeUTXOFromMultiset(ms, input.UTXOEntry, &input.PreviousOutpoint)
		if err != nil {
			return err
		}
	}

	for i, output := range transaction.Outputs {
		outpoint := &externalapi.DomainOutpoint{
			TransactionID: *transactionID,
			Index:         uint32(i),
		}
		utxoEntry := utxo.NewUTXOEntry(output.Value, output.ScriptPublicKey, isCoinbase, blockDAAScore)
		err := addUTXOToMultiset(ms, utxoEntry, outpoint)
		if err != nil {
			return err
		}
	}

	return nil
}

func addUTXOToMultiset(ms model.Multiset, entry externalapi.UTXOEntry,
	outpoint *externalapi.DomainOutpoint) error {

	serializedUTXO, err := utxo.SerializeUTXO(entry, outpoint)
	if err != nil {
		return err
	}
	ms.Add(serializedUTXO)
	return nil
}

func removeUTXOFromMultiset(ms model.Multiset, entry externalapi.UTXOEntry,
	outpoint *externalapi.DomainOutpoint) error {

	serializedUTXO, err := utxo.SerializeUTXO(entry, outpoint)
	if err != nil {
		return err
	}
	ms.Remove(serializedUTXO)
	return nil
}
