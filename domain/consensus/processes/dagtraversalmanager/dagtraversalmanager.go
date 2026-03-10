package dagtraversalmanager

import (
	"github.com/rupixnet/rupixd/domain/consensus/model"
	"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
	"github.com/pkg/errors"
)

type dagTraversalManager struct {
	databaseContext model.DBReader

	dagTopologyManager             model.DAGTopologyManager
	ghostdagManager                model.GHOSTDAGManager
	ghostdagDataStore              model.GHOSTDAGDataStore
	reachabilityManager            model.ReachabilityManager
	daaWindowStore                 model.BlocksWithTrustedDataDAAWindowStore
	genesisHash                    *externalapi.DomainHash
	difficultyAdjustmentWindowSize int
	windowHeapSliceStore           model.WindowHeapSliceStore
}

func New(
	databaseContext model.DBReader,
	dagTopologyManager model.DAGTopologyManager,
	ghostdagDataStore model.GHOSTDAGDataStore,
	reachabilityManager model.ReachabilityManager,
	ghostdagManager model.GHOSTDAGManager,
	daaWindowStore model.BlocksWithTrustedDataDAAWindowStore,
	windowHeapSliceStore model.WindowHeapSliceStore,
	genesisHash *externalapi.DomainHash,
	difficultyAdjustmentWindowSize int) model.DAGTraversalManager {
	return &dagTraversalManager{
		databaseContext:                databaseContext,
		dagTopologyManager:             dagTopologyManager,
		ghostdagDataStore:              ghostdagDataStore,
		reachabilityManager:            reachabilityManager,
		ghostdagManager:                ghostdagManager,
		daaWindowStore:                 daaWindowStore,
		genesisHash:                    genesisHash,
		difficultyAdjustmentWindowSize: difficultyAdjustmentWindowSize,
		windowHeapSliceStore:           windowHeapSliceStore,
	}
}

func (dtm *dagTraversalManager) LowestChainBlockAboveOrEqualToBlueScore(stagingArea *model.StagingArea, highHash *externalapi.DomainHash, blueScore uint64) (*externalapi.DomainHash, error) {
	highBlockGHOSTDAGData, err := dtm.ghostdagDataStore.Get(dtm.databaseContext, stagingArea, highHash, false)
	if err != nil {
		return nil, err
	}

	if highBlockGHOSTDAGData.BlueScore() < blueScore {
		return nil, errors.Errorf("the given blue score %d is higher than block %s blue score of %d",
			blueScore, highHash, highBlockGHOSTDAGData.BlueScore())
	}

	currentHash := highHash
	currentBlockGHOSTDAGData := highBlockGHOSTDAGData

	for !currentHash.Equal(dtm.genesisHash) {
		selectedParentBlockGHOSTDAGData, err := dtm.ghostdagDataStore.Get(dtm.databaseContext, stagingArea,
			currentBlockGHOSTDAGData.SelectedParent(), false)
		if err != nil {
			return nil, err
		}

		if selectedParentBlockGHOSTDAGData.BlueScore() < blueScore {
			break
		}
		currentHash = currentBlockGHOSTDAGData.SelectedParent()
		currentBlockGHOSTDAGData = selectedParentBlockGHOSTDAGData
	}

	return currentHash, nil
}

func (dtm *dagTraversalManager) CalculateChainPath(stagingArea *model.StagingArea,
	fromBlockHash, toBlockHash *externalapi.DomainHash) (*externalapi.SelectedChainPath, error) {

	var removed []*externalapi.DomainHash
	current := fromBlockHash
	for {
        if current == nil || current.Equal(model.VirtualGenesisBlockHash) {
            break
        }
		isCurrentInTheSelectedParentChainOfNewVirtualSelectedParent, err :=
			dtm.dagTopologyManager.IsInSelectedParentChainOf(stagingArea, current, toBlockHash)
		if err != nil {
			return nil, err
		}
		if isCurrentInTheSelectedParentChainOfNewVirtualSelectedParent {
			break
		}
		removed = append(removed, current)
		currentGHOSTDAGData, err := dtm.ghostdagDataStore.Get(dtm.databaseContext, stagingArea, current, false)
		if err != nil {
			return nil, err
		}
		current = currentGHOSTDAGData.SelectedParent()
		if current == nil || current.Equal(model.VirtualGenesisBlockHash) {
			break
		}
	}
	commonAncestor := current

	var added []*externalapi.DomainHash
	current = toBlockHash
	for current != nil && !current.Equal(commonAncestor) {
		added = append(added, current)
		currentGHOSTDAGData, err := dtm.ghostdagDataStore.Get(dtm.databaseContext, stagingArea, current, false)
		if err != nil {
			return nil, err
		}
		current = currentGHOSTDAGData.SelectedParent()
	}

	for i, j := 0, len(added)-1; i < j; i, j = i+1, j-1 {
		added[i], added[j] = added[j], added[i]
	}

	return &externalapi.SelectedChainPath{
		Added:   added,
		Removed: removed,
	}, nil
}

