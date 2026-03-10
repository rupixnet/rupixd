package dagtraversalmanager

import (
    "github.com/rupixnet/rupixd/domain/consensus/model"
    "github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
    "github.com/rupixnet/rupixd/infrastructure/db/database"
)

func (dtm *dagTraversalManager) DAABlockWindow(stagingArea *model.StagingArea, highHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
    return dtm.BlockWindow(stagingArea, highHash, dtm.difficultyAdjustmentWindowSize)
}

func (dtm *dagTraversalManager) BlockWindow(stagingArea *model.StagingArea, highHash *externalapi.DomainHash,
    windowSize int) ([]*externalapi.DomainHash, error) {

    windowHeap, err := dtm.blockWindowHeap(stagingArea, highHash, windowSize)
    if err != nil {
        return nil, err
    }

    window := make([]*externalapi.DomainHash, 0, len(windowHeap.impl.slice))
    for _, b := range windowHeap.impl.slice {
        window = append(window, b.Hash)
    }
    return window, nil
}

func (dtm *dagTraversalManager) blockWindowHeap(stagingArea *model.StagingArea,
    highHash *externalapi.DomainHash, windowSize int) (*sizedUpBlockHeap, error) {
    windowHeapSlice, err := dtm.windowHeapSliceStore.Get(stagingArea, highHash, windowSize)
    sliceNotCached := database.IsNotFoundError(err)
    if !sliceNotCached && err != nil {
        return nil, err
    }
    if !sliceNotCached {
        return dtm.newSizedUpHeapFromSlice(stagingArea, windowHeapSlice), nil
    }

    heap, err := dtm.calculateBlockWindowHeap(stagingArea, highHash, windowSize)
    if err != nil {
        return nil, err
    }

    if !highHash.Equal(model.VirtualBlockHash) {
        dtm.windowHeapSliceStore.Stage(stagingArea, highHash, windowSize, heap.impl.slice)
    }
    return heap, nil
}

func (dtm *dagTraversalManager) calculateBlockWindowHeap(stagingArea *model.StagingArea,
    highHash *externalapi.DomainHash, windowSize int) (*sizedUpBlockHeap, error) {

    windowHeap := dtm.newSizedUpHeap(stagingArea, windowSize)
    if highHash.Equal(dtm.genesisHash) {
        return windowHeap, nil
    }
    if windowSize == 0 {
        return windowHeap, nil
    }

    current := highHash
    currentGHOSTDAGData, err := dtm.ghostdagDataStore.Get(dtm.databaseContext, stagingArea, highHash, false)
    if err != nil {
        if database.IsNotFoundError(err) {
            return windowHeap, nil
        }
        return nil, err
    }

    _, err = dtm.daaWindowStore.DAAWindowBlock(dtm.databaseContext, stagingArea, current, 0)
    isNonTrustedBlock := database.IsNotFoundError(err)
    if !isNonTrustedBlock && err != nil {
        return nil, err
    }

    if isNonTrustedBlock && currentGHOSTDAGData.SelectedParent() != nil {
        windowHeapSlice, err := dtm.windowHeapSliceStore.Get(stagingArea, currentGHOSTDAGData.SelectedParent(), windowSize)
        selectedParentNotCached := database.IsNotFoundError(err)
        if !selectedParentNotCached && err != nil {
            return nil, err
        }
        if !selectedParentNotCached {
            windowHeap := dtm.newSizedUpHeapFromSlice(stagingArea, windowHeapSlice)
            if !currentGHOSTDAGData.SelectedParent().Equal(dtm.genesisHash) {
                selectedParentGHOSTDAGData, err := dtm.ghostdagDataStore.Get(
                    dtm.databaseContext, stagingArea, currentGHOSTDAGData.SelectedParent(), false)
                if err != nil {
                    return nil, err
                }
                _, err = dtm.tryPushMergeSet(windowHeap, currentGHOSTDAGData, selectedParentGHOSTDAGData)
                if err != nil {
                    return nil, err
                }
            }
            return windowHeap, nil
        }
    }

    for {
        if currentGHOSTDAGData.SelectedParent() == nil {
            break
        }
        if currentGHOSTDAGData.SelectedParent().Equal(dtm.genesisHash) {
            break
        }

        _, err := dtm.daaWindowStore.DAAWindowBlock(dtm.databaseContext, stagingArea, current, 0)
        currentIsNonTrustedBlock := database.IsNotFoundError(err)
        if !currentIsNonTrustedBlock && err != nil {
            return nil, err
        }

        if !currentIsNonTrustedBlock {
            for i := uint64(0); ; i++ {
                daaBlock, err := dtm.daaWindowStore.DAAWindowBlock(dtm.databaseContext, stagingArea, current, i)
                if database.IsNotFoundError(err) {
                    break
                }
                if err != nil {
                    return nil, err
                }
                _, err = windowHeap.tryPushWithGHOSTDAGData(daaBlock.Hash, daaBlock.GHOSTDAGData)
                if err != nil {
                    return nil, err
                }
            }
            break
        }

        selectedParentGHOSTDAGData, err := dtm.ghostdagDataStore.Get(
            dtm.databaseContext, stagingArea, currentGHOSTDAGData.SelectedParent(), false)
        if err != nil {
            return nil, err
        }

        done, err := dtm.tryPushMergeSet(windowHeap, currentGHOSTDAGData, selectedParentGHOSTDAGData)
        if err != nil {
            return nil, err
        }
        if done {
            break
        }

        current = currentGHOSTDAGData.SelectedParent()
        currentGHOSTDAGData = selectedParentGHOSTDAGData
    }

    return windowHeap, nil
}

func (dtm *dagTraversalManager) tryPushMergeSet(windowHeap *sizedUpBlockHeap, currentGHOSTDAGData, selectedParentGHOSTDAGData *externalapi.BlockGHOSTDAGData) (bool, error) {
    added, err := windowHeap.tryPushWithGHOSTDAGData(currentGHOSTDAGData.SelectedParent(), selectedParentGHOSTDAGData)
    if err != nil {
        return false, err
    }

    if !added {
        return true, nil
    }

    allMergeSetBlues := currentGHOSTDAGData.MergeSetBlues()
    if len(allMergeSetBlues) == 0 {
        return false, nil
    }
    mergeSetBlues := allMergeSetBlues[1:]
    for i := len(mergeSetBlues) - 1; i >= 0; i-- {
        added, err := windowHeap.tryPush(mergeSetBlues[i])
        if err != nil {
            return false, err
        }
        if !added {
            break
        }
    }

    mergeSetReds := currentGHOSTDAGData.MergeSetReds()
    for i := len(mergeSetReds) - 1; i >= 0; i-- {
        added, err := windowHeap.tryPush(mergeSetReds[i])
        if err != nil {
            return false, err
        }
        if !added {
            break
        }
    }

    return false, nil
}
