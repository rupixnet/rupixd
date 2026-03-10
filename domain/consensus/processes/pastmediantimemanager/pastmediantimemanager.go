package pastmediantimemanager

import (
    "sort"

    "github.com/rupixnet/rupixd/domain/consensus/utils/sorters"

    "github.com/rupixnet/rupixd/domain/consensus/model"
    "github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
    "github.com/pkg/errors"
)

type pastMedianTimeManager struct {
    timestampDeviationTolerance int
    databaseContext             model.DBReader
    dagTraversalManager         model.DAGTraversalManager
    blockHeaderStore            model.BlockHeaderStore
    ghostdagDataStore           model.GHOSTDAGDataStore
    genesisHash                 *externalapi.DomainHash
    virtualPastMedianTimeCache  int64
}

func New(timestampDeviationTolerance int,
    databaseContext model.DBReader,
    dagTraversalManager model.DAGTraversalManager,
    blockHeaderStore model.BlockHeaderStore,
    ghostdagDataStore model.GHOSTDAGDataStore,
    genesisHash *externalapi.DomainHash) model.PastMedianTimeManager {

    return &pastMedianTimeManager{
        timestampDeviationTolerance: timestampDeviationTolerance,
        databaseContext:             databaseContext,
        dagTraversalManager:         dagTraversalManager,
        blockHeaderStore:            blockHeaderStore,
        ghostdagDataStore:           ghostdagDataStore,
        genesisHash:                 genesisHash,
    }
}

func (pmtm *pastMedianTimeManager) PastMedianTime(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (int64, error) {
    if blockHash == model.VirtualBlockHash && pmtm.virtualPastMedianTimeCache != 0 {
        return pmtm.virtualPastMedianTimeCache, nil
    }
    window, err := pmtm.dagTraversalManager.BlockWindow(stagingArea, blockHash, 2*pmtm.timestampDeviationTolerance-1)
    if err != nil {
        return 0, err
    }
    if len(window) == 0 {
    header, err := pmtm.blockHeaderStore.BlockHeader(pmtm.databaseContext, stagingArea, pmtm.genesisHash)
    if err != nil {
        return 0, nil
    }
    return header.TimeInMilliseconds(), nil
}
    pastMedianTime, err := pmtm.windowMedianTimestamp(stagingArea, window)
    if err != nil {
        return 0, err
    }
    if blockHash == model.VirtualBlockHash {
        pmtm.virtualPastMedianTimeCache = pastMedianTime
    }
    return pastMedianTime, nil
}

func (pmtm *pastMedianTimeManager) windowMedianTimestamp(
    stagingArea *model.StagingArea, window []*externalapi.DomainHash) (int64, error) {

    if len(window) == 0 {
        return 0, errors.New("Cannot calculate median timestamp for an empty block window")
    }
    timestamps := make([]int64, len(window))
    for i, blockHash := range window {
        if blockHash.Equal(model.VirtualGenesisBlockHash) {
            continue
        }
        blockHeader, err := pmtm.blockHeaderStore.BlockHeader(pmtm.databaseContext, stagingArea, blockHash)
        if err != nil {
            continue
        }
        timestamps[i] = blockHeader.TimeInMilliseconds()
    }
    sort.Sort(sorters.Int64Slice(timestamps))
    return timestamps[len(timestamps)/2], nil
}

func (pmtm *pastMedianTimeManager) InvalidateVirtualPastMedianTimeCache() {
    pmtm.virtualPastMedianTimeCache = 0
}
