package consensusstatemanager

import (
	"fmt"
	"github.com/rupixnet/rupixd/domain/consensus/model"
	"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
	"github.com/rupixnet/rupixd/domain/consensus/utils/utxo"
	"github.com/rupixnet/rupixd/infrastructure/db/database"
	"github.com/rupixnet/rupixd/infrastructure/logger"
)

func (csm *consensusStateManager) AddBlock(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash, updateVirtual bool) (
	*externalapi.SelectedChainPath, externalapi.UTXODiff, *model.UTXODiffReversalData, error) {

	onEnd := logger.LogAndMeasureExecutionTime(log, "csm.AddBlock")
	defer onEnd()

	var reversalData *model.UTXODiffReversalData
	if updateVirtual {
		log.Debugf("Resolving whether the block %s is the next virtual selected parent", blockHash)
		isCandidateToBeNextVirtualSelectedParent, err := csm.isCandidateToBeNextVirtualSelectedParent(stagingArea, blockHash)
		if err != nil {
			return nil, nil, nil, err
		}

		if isCandidateToBeNextVirtualSelectedParent {
			log.Debugf("Block %s is candidate to be the next virtual selected parent. Resolving whether it violates finality", blockHash)
			isViolatingFinality, shouldNotify, err := csm.isViolatingFinality(stagingArea, blockHash)
if err != nil {
    fmt.Printf("FAIL isViolatingFinality: %T :: %+v\n", err, err)
    return nil, nil, nil, err
}

			if shouldNotify {
				log.Warnf("Finality Violation Detected! Block %s violates finality!", blockHash)
			}

			if !isViolatingFinality {
				log.Debugf("Block %s doesn't violate finality. Resolving its block status", blockHash)
				var blockStatus externalapi.BlockStatus
				blockStatus, reversalData, err = csm.resolveBlockStatus(stagingArea, blockHash, true)
if err != nil {
    fmt.Printf("FAIL resolveBlockStatus: %T :: %+v\n", err, err)
    return nil, nil, nil, err
}
				log.Debugf("Block %s resolved to status `%s`", blockHash, blockStatus)
			}
		} else {
			log.Debugf("Block %s is not the next virtual selected parent, therefore its status remains `%s`", blockHash, externalapi.StatusUTXOPendingVerification)
		}
	}

	log.Debugf("Adding block %s to the DAG tips", blockHash)
	newTips, err := csm.addTip(stagingArea, blockHash)
if err != nil {
    fmt.Printf("FAIL addTip: %T :: %+v\n", err, err)
    return nil, nil, nil, err
}
	log.Debugf("After adding %s, the amount of new tips are %d", blockHash, len(newTips))

	if !updateVirtual {
		return &externalapi.SelectedChainPath{}, utxo.NewUTXODiff(), nil, nil
	}

	log.Debugf("Updating the virtual with the new tips")
    selectedParentChainChanges, virtualUTXODiff, err := csm.updateVirtual(stagingArea, blockHash, newTips)
    if err != nil {
    fmt.Printf("FAIL updateVirtual: %T :: %+v\n", err, err)
    return nil, nil, nil, err
    }
	return selectedParentChainChanges, virtualUTXODiff, reversalData, nil
}

func (csm *consensusStateManager) isCandidateToBeNextVirtualSelectedParent(
	stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (bool, error) {

	log.Tracef("isCandidateToBeNextVirtualSelectedParent start for block %s", blockHash)
	defer log.Tracef("isCandidateToBeNextVirtualSelectedParent end for block %s", blockHash)

	if blockHash == nil {
		return false, nil
	}

	if blockHash.Equal(model.VirtualGenesisBlockHash) {
		return false, nil
	}

	if blockHash.Equal(csm.genesisHash) {
		log.Debugf("Block %s is the genesis block, therefore it is the selected parent by definition", blockHash)
		return true, nil
	}

	virtualGhostdagData, err := csm.ghostdagDataStore.Get(csm.databaseContext, stagingArea, model.VirtualBlockHash, false)
	if err != nil {
		// RUPIX FIX: primer arranque — VirtualBlock aun no existe en DB
		if database.IsNotFoundError(err) {
			return true, nil
		}
		return false, err
	}

	if virtualGhostdagData.SelectedParent() == nil || virtualGhostdagData.SelectedParent().Equal(model.VirtualGenesisBlockHash) {
		return true, nil
	}

	log.Debugf("Selecting the next selected parent between the block %s the current selected parent %s", blockHash, virtualGhostdagData.SelectedParent())
	nextVirtualSelectedParent, err := csm.ghostdagManager.ChooseSelectedParent(
		stagingArea, virtualGhostdagData.SelectedParent(), blockHash)
	if err != nil {
		return false, err
	}
	log.Debugf("The next selected parent is: %s", nextVirtualSelectedParent)

	return blockHash.Equal(nextVirtualSelectedParent), nil
}

func (csm *consensusStateManager) addTip(stagingArea *model.StagingArea, newTipHash *externalapi.DomainHash) (newTips []*externalapi.DomainHash, err error) {
	log.Tracef("addTip start for new tip %s", newTipHash)
	defer log.Tracef("addTip end for new tip %s", newTipHash)

	log.Debugf("Calculating the new tips for new tip %s", newTipHash)
	newTips, err = csm.calculateNewTips(stagingArea, newTipHash)
	if err != nil {
		return nil, err
	}

	csm.consensusStateStore.StageTips(stagingArea, newTips)
	log.Debugf("Staged the new tips, len: %d", len(newTips))

	return newTips, nil
}

func (csm *consensusStateManager) calculateNewTips(
	stagingArea *model.StagingArea, newTipHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {

	log.Tracef("calculateNewTips start for new tip %s", newTipHash)
	defer log.Tracef("calculateNewTips end for new tip %s", newTipHash)

	if newTipHash.Equal(csm.genesisHash) {
		log.Debugf("The new tip is the genesis block, therefore it is the only tip by definition")
		return []*externalapi.DomainHash{newTipHash}, nil
	}

	currentTips, err := csm.consensusStateStore.Tips(stagingArea, csm.databaseContext)
	if err != nil {
		return nil, err
	}
	log.Debugf("The number of tips is: %d", len(currentTips))
	log.Tracef("The current tips are: %s", currentTips)

	newTipParents, err := csm.dagTopologyManager.Parents(stagingArea, newTipHash)
	if err != nil {
		return nil, err
	}
	log.Debugf("The parents of the new tip are: %s", newTipParents)

	newTips := []*externalapi.DomainHash{newTipHash}

	for _, currentTip := range currentTips {
		isCurrentTipInNewTipParents := false
		for _, newTipParent := range newTipParents {
			if currentTip.Equal(newTipParent) {
				isCurrentTipInNewTipParents = true
				break
			}
		}
		if !isCurrentTipInNewTipParents {
			newTips = append(newTips, currentTip)
		}
	}
	log.Debugf("The new number of tips is: %d", len(newTips))
	log.Tracef("The new tips are: %s", newTips)

	return newTips, nil
}
