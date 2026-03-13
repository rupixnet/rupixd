package consensusstatemanager

import (
        "fmt"
        "github.com/rupixnet/rupixd/domain/consensus/model"
        "github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
        "github.com/rupixnet/rupixd/infrastructure/logger"
        "github.com/rupixnet/rupixd/domain/consensus/utils/utxo"
        "github.com/rupixnet/rupixd/infrastructure/db/database"
)

func (csm *consensusStateManager) updateVirtual(stagingArea *model.StagingArea, newBlockHash *externalapi.DomainHash,
	tips []*externalapi.DomainHash) (*externalapi.SelectedChainPath, externalapi.UTXODiff, error) {

	onEnd := logger.LogAndMeasureExecutionTime(log, "updateVirtual")
	defer onEnd()

	log.Debugf("updateVirtual start for block %s", newBlockHash)

	log.Debugf("Saving a reference to the GHOSTDAG data of the old virtual")
	var oldVirtualSelectedParent *externalapi.DomainHash
	if !newBlockHash.Equal(csm.genesisHash) {
		oldVirtualGHOSTDAGData, err := csm.ghostdagDataStore.Get(csm.databaseContext, stagingArea, model.VirtualBlockHash, false)
		if err != nil {
			return nil, nil, err
		}
		oldVirtualSelectedParent = oldVirtualGHOSTDAGData.SelectedParent()
	}

	log.Debugf("Picking virtual parents from tips len: %d", len(tips))
	virtualParents, err := csm.pickVirtualParents(stagingArea, tips)
	if err != nil {
		return nil, nil, err
	}
	log.Debugf("Picked virtual parents: %s", virtualParents)

	virtualUTXODiff, err := csm.updateVirtualWithParents(stagingArea, virtualParents)
	if err != nil {
		return nil, nil, err
	}

	log.Debugf("Calculating selected parent chain changes")
	var selectedParentChainChanges *externalapi.SelectedChainPath
	if !newBlockHash.Equal(csm.genesisHash) {
		newVirtualGHOSTDAGData, err := csm.ghostdagDataStore.Get(csm.databaseContext, stagingArea, model.VirtualBlockHash, false)
		if err != nil {
			return nil, nil, err
		}
		newVirtualSelectedParent := newVirtualGHOSTDAGData.SelectedParent()
		selectedParentChainChanges, err = csm.dagTraversalManager.
			CalculateChainPath(stagingArea, oldVirtualSelectedParent, newVirtualSelectedParent)
		if err != nil {
			return nil, nil, err
		}
		log.Debugf("Selected parent chain changes: %d blocks were removed and %d blocks were added",
			len(selectedParentChainChanges.Removed), len(selectedParentChainChanges.Added))
	}

	return selectedParentChainChanges, virtualUTXODiff, nil
}

func (csm *consensusStateManager) updateVirtualWithParents(
	stagingArea *model.StagingArea, virtualParents []*externalapi.DomainHash) (externalapi.UTXODiff, error) {
	err := csm.dagTopologyManager.SetParents(stagingArea, model.VirtualBlockHash, virtualParents)
	if err != nil {
		return nil, err
	}
	log.Debugf("Set new parents for the virtual block hash")

	fmt.Printf("VIRTUAL PARENTS before GHOSTDAG: %d parents: %v\n", len(virtualParents), virtualParents)
    for _, vp := range virtualParents {
        if gd, e := csm.ghostdagDataStore.Get(csm.databaseContext, stagingArea, vp, false); e == nil {
            fmt.Printf("  parent %s ghostdag: sp=%v blues=%d reds=%d blueScore=%d\n", vp, gd.SelectedParent(), len(gd.MergeSetBlues()), len(gd.MergeSetReds()), gd.BlueScore())
        } else {
            fmt.Printf("  parent %s ghostdag ERROR: %v\n", vp, e)
        }
    }
	err = csm.ghostdagManager.GHOSTDAG(stagingArea, model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}
if vGhostdag, e := csm.ghostdagDataStore.Get(csm.databaseContext, stagingArea, model.VirtualBlockHash, false); e == nil {
		fmt.Printf("VIRTUAL AFTER GHOSTDAG: sp=%s blues=%d\n", vGhostdag.SelectedParent(), len(vGhostdag.MergeSetBlues()))
	}
// This is needed for `csm.CalculatePastUTXOAndAcceptanceData`
	_, err = csm.difficultyManager.StageDAADataAndReturnRequiredDifficulty(stagingArea, model.VirtualBlockHash, false)
	if err != nil {
		return nil, err
	}

	// DEBUG
	if vGhostdag, e := csm.ghostdagDataStore.Get(csm.databaseContext, stagingArea, model.VirtualBlockHash, false); e == nil {
		fmt.Printf("VIRTUAL sp=%s blues=%d reds=%d\n", vGhostdag.SelectedParent(), len(vGhostdag.MergeSetBlues()), len(vGhostdag.MergeSetReds()))
	}

	log.Debugf("Calculating past UTXO, acceptance data, and multiset for the new virtual block")
	virtualUTXODiff, virtualAcceptanceData, virtualMultiset, err :=
		csm.CalculatePastUTXOAndAcceptanceData(stagingArea, model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}
	fmt.Printf("UTXO VIRTUAL DIFF toAdd=%d toRemove=%d\n",
    virtualUTXODiff.ToAdd().Len(), virtualUTXODiff.ToRemove().Len())
    log.Infof("Calculated the past UTXO of the new virtual. "+
    "Diff toAdd length: %d, toRemove length: %d",
    virtualUTXODiff.ToAdd().Len(), virtualUTXODiff.ToRemove().Len())

	log.Debugf("Staging new acceptance data for the virtual block")
	csm.acceptanceDataStore.Stage(stagingArea, model.VirtualBlockHash, virtualAcceptanceData)

	log.Debugf("Staging new multiset for the virtual block")
	csm.multisetStore.Stage(stagingArea, model.VirtualBlockHash, virtualMultiset)

	log.Debugf("Staging new UTXO diff for the virtual block")
	csm.consensusStateStore.StageVirtualUTXODiff(stagingArea, virtualUTXODiff)

	log.Debugf("Updating the selected tip's utxo-diff")
	err = csm.updateSelectedTipUTXODiff(stagingArea, virtualUTXODiff)
	if err != nil {
		return nil, err
	}

	return virtualUTXODiff, nil
}

func (csm *consensusStateManager) updateSelectedTipUTXODiff(
	stagingArea *model.StagingArea, virtualUTXODiff externalapi.UTXODiff) error {

	onEnd := logger.LogAndMeasureExecutionTime(log, "updateSelectedTipUTXODiff")
	defer onEnd()

	selectedTip, err := csm.virtualSelectedParent(stagingArea)
if err != nil {
    return err
}
if selectedTip == nil {
    return nil
}

	log.Debugf("Calculating new UTXO diff for virtual diff parent %s", selectedTip)
	selectedTipUTXODiff, err := csm.utxoDiffStore.UTXODiff(csm.databaseContext, stagingArea, selectedTip)
if err != nil {
    if database.IsNotFoundError(err) {
        selectedTipUTXODiff = utxo.NewUTXODiff()
    } else {
        return err
    }
}
	newDiff, err := virtualUTXODiff.DiffFrom(selectedTipUTXODiff)
	if err != nil {
		return err
	}

	log.Debugf("Staging new UTXO diff for virtual diff parent %s", selectedTip)
	csm.stageDiff(stagingArea, selectedTip, newDiff, nil)

	return nil
}

