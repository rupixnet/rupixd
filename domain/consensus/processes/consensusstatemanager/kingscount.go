package consensusstatemanager

import (
	"github.com/pkg/errors"
	"github.com/rupixnet/rupixd/domain/consensus/model"
	"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
	"github.com/rupixnet/rupixd/domain/consensus/ruleerrors"
	"github.com/rupixnet/rupixd/domain/consensus/utils/constants"
)

// calculateKingsCount es el espejo de calculateMultiset para el tope de Kings:
//
//	kings(bloque) = kings(selectedParent) + Kings nacidos - Kings quemados
//
// derivado del MISMO acceptanceData que produce el UTXO set y el multiset.
// Si el resultado supera constants.MaxKings, el bloque entero es invalido:
// el King 2,101 no puede existir en ninguna cadena valida.
func (csm *consensusStateManager) calculateKingsCount(stagingArea *model.StagingArea,
	blockHash *externalapi.DomainHash,
	acceptanceData externalapi.AcceptanceData,
	blockGHOSTDAGData *externalapi.BlockGHOSTDAGData) (uint64, error) {

	if blockHash.Equal(csm.genesisHash) {
		return 0, nil // el genesis no tiene Kings: cero premine en todos los niveles
	}

	count, err := csm.kingsCountStore.Get(csm.databaseContext, stagingArea, blockGHOSTDAGData.SelectedParent())
	if err != nil {
		return 0, err
	}

	for _, blockAcceptanceData := range acceptanceData {
		for _, transactionAcceptanceData := range blockAcceptanceData.TransactionAcceptanceData {
			if !transactionAcceptanceData.IsAccepted {
				continue
			}
			tx := transactionAcceptanceData.Transaction
			for _, input := range tx.Inputs {
				if input.UTXOEntry.ScriptPublicKey().Version == constants.LevelKings {
					count-- // un King quemado (o gastado hacia... nada: los Kings solo se transfieren)
				}
			}
			for _, output := range tx.Outputs {
				if output.ScriptPublicKey.Version == constants.LevelKings {
					count++
				}
			}
		}
	}

	if count > constants.MaxKings {
		return 0, errors.Wrapf(ruleerrors.ErrKingsCapExceeded,
			"el bloque %s llevaria el conteo de Kings a %d, tope absoluto %d",
			blockHash, count, constants.MaxKings)
	}
	return count, nil
}

// countKingsInImportedUTXOSet recorre el UTXO set del pruning point importado
// y cuenta las entradas de nivel Kings. Se llama una vez por importacion.
func (csm *consensusStateManager) countKingsInImportedUTXOSet() (uint64, error) {
	iterator, err := csm.pruningStore.ImportedPruningPointUTXOIterator(csm.databaseContext)
	if err != nil {
		return 0, err
	}
	defer iterator.Close()

	count := uint64(0)
	for ok := iterator.First(); ok; ok = iterator.Next() {
		_, entry, err := iterator.Get()
		if err != nil {
			return 0, err
		}
		if entry.ScriptPublicKey().Version == constants.LevelKings {
			count++
		}
	}
	return count, nil
}
