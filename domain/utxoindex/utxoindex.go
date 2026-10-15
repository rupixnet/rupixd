package utxoindex

import (
	"github.com/rupixnet/rupixd/domain"
	"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
	"github.com/rupixnet/rupixd/infrastructure/db/database"
	"github.com/rupixnet/rupixd/infrastructure/logger"
	"sync"
)

// UTXOIndex maintains an index between transaction scriptPublicKeys
// and UTXOs
type UTXOIndex struct {
	domain domain.Domain
	store  *utxoIndexStore

	mutex sync.Mutex

	// Rupix: conteo de gemas vivas por nivel (0 Gold, 1 Diamante, 2 Platino, 3 Rodio,
	// 4 Kings). Se llena recorriendo el indice una vez (gemCountsListo) y desde ahi se
	// mantiene en cada Update: suma lo anadido, resta lo quitado. Siempre al dia con el
	// indice, nunca una foto vieja (14-oct-2026: una cache de 30 s hizo que el explorador
	// mostrara "quemados 21" donde eran 20).
	gemCounts      [5]uint64
	gemCountsListo bool
}

// New creates a new UTXO index.
//
// NOTE: While this is called no new blocks can be added to the consensus.
func New(domain domain.Domain, database database.Database) (*UTXOIndex, error) {
	utxoIndex := &UTXOIndex{
		domain: domain,
		store:  newUTXOIndexStore(database),
	}
	isSynced, err := utxoIndex.isSynced()
	if err != nil {
		return nil, err
	}

	///Has check is for migration to circulating supply, can be removed eventually.
	hasCirculatingSupplyKey, err := utxoIndex.store.database.Has(circulatingSupplyKey)
	if err != nil {
		return nil, err
	}

	if !isSynced || !hasCirculatingSupplyKey {

		err := utxoIndex.Reset()
		if err != nil {
			return nil, err
		}
	}

	return utxoIndex, nil
}

// Reset deletes the whole UTXO index and resyncs it from consensus.
func (ui *UTXOIndex) Reset() error {
	ui.mutex.Lock()
	defer ui.mutex.Unlock()

	log.Infof("Starting UTXO index reset")

	err := ui.store.deleteAll()
	if err != nil {
		return err
	}

	virtualInfo, err := ui.domain.Consensus().GetVirtualInfo()
	if err != nil {
		return err
	}

	err = ui.store.initializeCirculatingSompiSupply() //At this point the database is empty, so the sole purpose of this call is to initialize the circulating supply key
	if err != nil {
		return err
	}

	var fromOutpoint *externalapi.DomainOutpoint
	for {
		const step = 1000
		virtualUTXOs, err := ui.domain.Consensus().GetVirtualUTXOs(virtualInfo.ParentHashes, fromOutpoint, step)
		if err != nil {
			return err
		}

		err = ui.store.addAndCommitOutpointsWithoutTransaction(virtualUTXOs)
		if err != nil {
			return err
		}

		if len(virtualUTXOs) < step {
			break
		}

		fromOutpoint = virtualUTXOs[len(virtualUTXOs)-1].Outpoint
	}

	// This has to be done last to mark that the reset went smoothly and no reset has to be called next time.
	err = ui.store.updateAndCommitVirtualParentsWithoutTransaction(virtualInfo.ParentHashes)
	if err != nil {
		return err
	}

	ui.gemCountsListo = false // el indice cambio entero; se vuelve a contar al pedirlo
	log.Infof("Finished UTXO index reset")
	return nil
}

func (ui *UTXOIndex) isSynced() (bool, error) {
	utxoIndexVirtualParents, err := ui.store.getVirtualParents()
	if err != nil {
		if database.IsNotFoundError(err) {
			return false, nil
		}
		return false, err
	}

	virtualInfo, err := ui.domain.Consensus().GetVirtualInfo()
	if err != nil {
		return false, err
	}

	return externalapi.HashesEqual(virtualInfo.ParentHashes, utxoIndexVirtualParents), nil
}

// Update updates the UTXO index with the given DAG selected parent chain changes
func (ui *UTXOIndex) Update(virtualChangeSet *externalapi.VirtualChangeSet) (*UTXOChanges, error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "UTXOIndex.Update")
	defer onEnd()

	ui.mutex.Lock()
	defer ui.mutex.Unlock()

	log.Tracef("Updating UTXO index with VirtualUTXODiff: %+v", virtualChangeSet.VirtualUTXODiff)
	err := ui.removeUTXOs(virtualChangeSet.VirtualUTXODiff.ToRemove())
	if err != nil {
		return nil, err
	}

	err = ui.addUTXOs(virtualChangeSet.VirtualUTXODiff.ToAdd())
	if err != nil {
		return nil, err
	}

	ui.store.updateVirtualParents(virtualChangeSet.VirtualParents)

	added, removed, _ := ui.store.stagedData()
	utxoIndexChanges := &UTXOChanges{
		Added:   added,
		Removed: removed,
	}

	err = ui.store.commit()
	if err != nil {
		return nil, err
	}
	ui.ajustarGemCounts(utxoIndexChanges)

	log.Tracef("UTXO index updated with the UTXOChanged: %+v", utxoIndexChanges)
	return utxoIndexChanges, nil
}

// ajustarGemCounts (Rupix) aplica al conteo de gemas vivas lo que acaba de entrar y
// salir del indice. Solo si el conteo ya esta listo; si no, lo hara el recorrido completo.
// Se llama con ui.mutex tomado.
func (ui *UTXOIndex) ajustarGemCounts(cambios *UTXOChanges) {
	if !ui.gemCountsListo {
		return
	}
	for _, pares := range cambios.Added {
		for _, entry := range pares {
			if v := entry.ScriptPublicKey().Version; int(v) < len(ui.gemCounts) {
				ui.gemCounts[v]++
			}
		}
	}
	for _, pares := range cambios.Removed {
		for _, entry := range pares {
			if v := entry.ScriptPublicKey().Version; int(v) < len(ui.gemCounts) {
				if ui.gemCounts[v] == 0 {
					// No deberia pasar: quitar lo que no se conto. Se vuelve a contar.
					log.Warnf("gemCounts: nivel %d en 0 al quitar un UTXO; se recontara", v)
					ui.gemCountsListo = false
					return
				}
				ui.gemCounts[v]--
			}
		}
	}
}

func (ui *UTXOIndex) addUTXOs(toAdd externalapi.UTXOCollection) error {
	iterator := toAdd.Iterator()
	defer iterator.Close()
	for ok := iterator.First(); ok; ok = iterator.Next() {
		outpoint, entry, err := iterator.Get()
		if err != nil {
			return err
		}

		log.Tracef("Adding outpoint %s to UTXO index", outpoint)
		err = ui.store.add(entry.ScriptPublicKey(), outpoint, entry)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ui *UTXOIndex) removeUTXOs(toRemove externalapi.UTXOCollection) error {
	iterator := toRemove.Iterator()
	defer iterator.Close()
	for ok := iterator.First(); ok; ok = iterator.Next() {
		outpoint, entry, err := iterator.Get()
		if err != nil {
			return err
		}

		log.Tracef("Removing outpoint %s from UTXO index", outpoint)
		err = ui.store.remove(entry.ScriptPublicKey(), outpoint, entry)
		if err != nil {
			return err
		}
	}
	return nil
}

// UTXOs returns all the UTXOs for the given scriptPublicKey
func (ui *UTXOIndex) UTXOs(scriptPublicKey *externalapi.ScriptPublicKey) (UTXOOutpointEntryPairs, error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "UTXOIndex.UTXOs")
	defer onEnd()

	ui.mutex.Lock()
	defer ui.mutex.Unlock()

	return ui.store.getUTXOOutpointEntryPairs(scriptPublicKey)
}

// GetCirculatingRupiaSupply returns the current circulating supply of rupias in the network
// GetGemCounts (Rupix) devuelve cuantos UTXOs de cada nivel viven ahora mismo
// (indice 0 = Gold, 1 Diamante, 2 Platino, 3 Rodio, 4 Kings), exactamente los del
// indice en este instante. El recorrido completo ocurre una sola vez (o tras un Reset);
// despues cada Update lo mantiene al dia.
func (ui *UTXOIndex) GetGemCounts() ([5]uint64, error) {
	ui.mutex.Lock()
	defer ui.mutex.Unlock()
	if ui.gemCountsListo {
		return ui.gemCounts, nil
	}
	conteo, err := ui.store.countGems()
	if err != nil {
		return conteo, err
	}
	ui.gemCounts, ui.gemCountsListo = conteo, true
	return conteo, nil
}

func (ui *UTXOIndex) GetCirculatingRupiaSupply() (uint64, error) {

	ui.mutex.Lock()
	defer ui.mutex.Unlock()

	return ui.store.getCirculatingSompiSupply()
}
