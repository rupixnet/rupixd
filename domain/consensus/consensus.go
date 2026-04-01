package consensus

import (
	"math"
	"math/big"
	"sync"

	"github.com/rupixnet/rupixd/domain/consensus/datastructures/addresslevelstore"
	"github.com/rupixnet/rupixd/domain/consensus/database"
	"github.com/rupixnet/rupixd/domain/consensus/model"
	"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
	"github.com/rupixnet/rupixd/domain/consensus/ruleerrors"
	"github.com/rupixnet/rupixd/domain/consensus/utils/reachabilitydata"
	"github.com/rupixnet/rupixd/infrastructure/logger"
	"github.com/rupixnet/rupixd/util/mstime"
	"github.com/rupixnet/rupixd/util/staging"
	"github.com/pkg/errors"
)

type consensus struct {
	lock            *sync.Mutex
	databaseContext model.DBManager

	genesisBlock *externalapi.DomainBlock
	genesisHash  *externalapi.DomainHash

	expectedDAAWindowDurationInMilliseconds int64

	blockProcessor        model.BlockProcessor
	blockBuilder          model.BlockBuilder
	consensusStateManager model.ConsensusStateManager
	transactionValidator  model.TransactionValidator
	syncManager           model.SyncManager
	pastMedianTimeManager model.PastMedianTimeManager
	blockValidator        model.BlockValidator
	coinbaseManager       model.CoinbaseManager
	dagTopologyManagers   []model.DAGTopologyManager
	dagTraversalManager   model.DAGTraversalManager
	difficultyManager     model.DifficultyManager
	ghostdagManagers      []model.GHOSTDAGManager
	headerTipsManager     model.HeadersSelectedTipManager
	mergeDepthManager     model.MergeDepthManager
	pruningManager        model.PruningManager
	reachabilityManager   model.ReachabilityManager
	finalityManager       model.FinalityManager
	pruningProofManager   model.PruningProofManager

	acceptanceDataStore                 model.AcceptanceDataStore
	blockStore                          model.BlockStore
	blockHeaderStore                    model.BlockHeaderStore
	pruningStore                        model.PruningStore
	ghostdagDataStores                  []model.GHOSTDAGDataStore
	blockRelationStores                 []model.BlockRelationStore
	blockStatusStore                    model.BlockStatusStore
	consensusStateStore                 model.ConsensusStateStore
	headersSelectedTipStore             model.HeaderSelectedTipStore
	multisetStore                       model.MultisetStore
	reachabilityDataStore               model.ReachabilityDataStore
	utxoDiffStore                       model.UTXODiffStore
	finalityStore                       model.FinalityStore
	headersSelectedChainStore           model.HeadersSelectedChainStore
	daaBlocksStore                      model.DAABlocksStore
	blocksWithTrustedDataDAAWindowStore model.BlocksWithTrustedDataDAAWindowStore

	addressLevelStore   *addresslevelstore.AddressLevelStore
	consensusEventsChan chan externalapi.ConsensusEvent
	virtualNotUpdated   bool
}

const virtualResolveChunk = 100

func (s *consensus) ValidateAndInsertBlockAsTrusted(block *externalapi.DomainBlock, updateVirtual bool) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	_, err := s.validateAndInsertBlockAsTrustedNoLock(block, updateVirtual)
	return err
}

func (s *consensus) validateAndInsertBlockAsTrustedNoLock(block *externalapi.DomainBlock, updateVirtual bool) (*externalapi.VirtualChangeSet, error) {
	virtualChangeSet, blockStatus, err := s.blockProcessor.ValidateAndInsertBlockAsTrusted(block, updateVirtual)
	if err != nil {
		return nil, err
	}
	if !updateVirtual && blockStatus != externalapi.StatusHeaderOnly {
		s.virtualNotUpdated = true
	}
	err = s.sendBlockAddedEvent(block, blockStatus)
	if err != nil {
		return nil, err
	}
	return virtualChangeSet, nil
}

func (s *consensus) ValidateAndInsertBlockWithTrustedData(block *externalapi.BlockWithTrustedData, validateUTXO bool) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	_, _, err := s.blockProcessor.ValidateAndInsertBlockWithTrustedData(block, validateUTXO)
	if err != nil {
		return err
	}
	return nil
}

// Init initializes consensus
func (s *consensus) Init(skipAddingGenesis bool) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	onEnd := logger.LogAndMeasureExecutionTime(log, "Init")
	defer onEnd()

	stagingArea := model.NewStagingArea()

	exists, err := s.blockStatusStore.Exists(s.databaseContext, stagingArea, model.VirtualGenesisBlockHash)
	if err != nil {
		return err
	}

	if !exists {
		s.blockStatusStore.Stage(stagingArea, model.VirtualGenesisBlockHash, externalapi.StatusUTXOValid)
		err = s.reachabilityManager.Init(stagingArea)
		if err != nil {
			return err
		}

		for _, dagTopologyManager := range s.dagTopologyManagers {
			err = dagTopologyManager.SetParents(stagingArea, model.VirtualGenesisBlockHash, nil)
			if err != nil {
				return err
			}
		}

		s.consensusStateStore.StageTips(stagingArea, []*externalapi.DomainHash{model.VirtualGenesisBlockHash})
		for _, dagTopologyManager := range s.dagTopologyManagers {
			err = dagTopologyManager.SetParents(stagingArea, model.VirtualBlockHash, []*externalapi.DomainHash{model.VirtualGenesisBlockHash})
			if err != nil {
				return err
			}
		}
		for _, ghostdagDataStore := range s.ghostdagDataStores {
			ghostdagDataStore.Stage(stagingArea, model.VirtualGenesisBlockHash, externalapi.NewBlockGHOSTDAGData(
				0,
				big.NewInt(0),
				nil,
				nil,
				nil,
				nil,
			), false)
		}

		// Rupix: pre-stage reachability data for genesis hash (fair launch, no premine)
		s.reachabilityDataStore.StageReachabilityData(stagingArea, s.genesisHash, reachabilitydata.New(
			[]*externalapi.DomainHash{},
			model.VirtualGenesisBlockHash,
			&model.ReachabilityInterval{Start: 0, End: math.MaxUint64},
			model.FutureCoveringTreeNodeSet{},
		))

		// Rupix: initialize VirtualBlockHash GHOSTDAG data so BuildParents works on fresh node
		for _, ghostdagDataStore := range s.ghostdagDataStores {
			ghostdagDataStore.Stage(stagingArea, model.VirtualBlockHash, externalapi.NewBlockGHOSTDAGData(
				0,
				big.NewInt(0),
				model.VirtualGenesisBlockHash,
				[]*externalapi.DomainHash{model.VirtualGenesisBlockHash},
				nil,
				nil,
			), false)
		}

		// Rupix: initialize pruning point to genesis so BuildParents works on fresh node
		s.pruningStore.StagePruningPoint(s.databaseContext, stagingArea, s.genesisHash)

		err = staging.CommitAllChanges(s.databaseContext, stagingArea)
		if err != nil {
			return err
		}
	}

	if !skipAddingGenesis && s.blockStore.Count(stagingArea) == 0 {
		genesisWithTrustedData := &externalapi.BlockWithTrustedData{
			Block:     s.genesisBlock,
			DAAWindow: nil,
			GHOSTDAGData: []*externalapi.BlockGHOSTDAGDataHashPair{
                                {
                                        Hash:         s.genesisHash,
                                        GHOSTDAGData: externalapi.NewBlockGHOSTDAGData(0, big.NewInt(0), model.VirtualGenesisBlockHash, []*externalapi.DomainHash{s.genesisHash}, nil, make(map[externalapi.DomainHash]externalapi.KType)),
                                },
                        },
		}
		_, _, err = s.blockProcessor.ValidateAndInsertBlockWithTrustedData(genesisWithTrustedData, true)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *consensus) PruningPointAndItsAnticone() ([]*externalapi.DomainHash, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.pruningManager.PruningPointAndItsAnticone()
}

func (s *consensus) BuildBlock(coinbaseData *externalapi.DomainCoinbaseData,
	transactions []*externalapi.DomainTransaction) (*externalapi.DomainBlock, error) {

	s.lock.Lock()
	defer s.lock.Unlock()

	block, _, err := s.blockBuilder.BuildBlock(coinbaseData, transactions)
	return block, err
}

func (s *consensus) BuildBlockTemplate(coinbaseData *externalapi.DomainCoinbaseData,
	transactions []*externalapi.DomainTransaction) (*externalapi.DomainBlockTemplate, error) {

	s.lock.Lock()
	defer s.lock.Unlock()

	block, hasRedReward, err := s.blockBuilder.BuildBlock(coinbaseData, transactions)
	if err != nil {
		return nil, err
	}

	isNearlySynced, err := s.isNearlySyncedNoLock()
	if err != nil {
		return nil, err
	}

	return &externalapi.DomainBlockTemplate{
		Block:                block,
		CoinbaseData:         coinbaseData,
		CoinbaseHasRedReward: hasRedReward,
		IsNearlySynced:       isNearlySynced,
	}, nil
}

func (s *consensus) ValidateAndInsertBlock(block *externalapi.DomainBlock, updateVirtual bool) error {
	if updateVirtual {
		s.lock.Lock()
		if s.virtualNotUpdated {
			for {
				_, isCompletelyResolved, err := s.resolveVirtualChunkNoLock(virtualResolveChunk)
				if err != nil {
					s.lock.Unlock()
					return err
				}
				if isCompletelyResolved {
					_, err = s.validateAndInsertBlockNoLock(block, updateVirtual)
					s.lock.Unlock()
					if err != nil {
						return err
					}
					return nil
				}
				s.lock.Unlock()
				s.lock.Lock()
			}
		}
		_, err := s.validateAndInsertBlockNoLock(block, updateVirtual)
		s.lock.Unlock()
		if err != nil {
			return err
		}
		return nil
	}

	return s.validateAndInsertBlockWithLock(block, updateVirtual)
}

func (s *consensus) validateAndInsertBlockWithLock(block *externalapi.DomainBlock, updateVirtual bool) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	_, err := s.validateAndInsertBlockNoLock(block, updateVirtual)
	if err != nil {
		return err
	}
	return nil
}

func (s *consensus) validateAndInsertBlockNoLock(block *externalapi.DomainBlock, updateVirtual bool) (*externalapi.VirtualChangeSet, error) {
	virtualChangeSet, blockStatus, err := s.blockProcessor.ValidateAndInsertBlock(block, updateVirtual)
	if err != nil {
		return nil, err
	}

	if !updateVirtual && blockStatus != externalapi.StatusHeaderOnly {
		s.virtualNotUpdated = true
	}

	err = s.sendBlockAddedEvent(block, blockStatus)
	if err != nil {
		return nil, err
	}

	err = s.sendVirtualChangedEvent(virtualChangeSet, updateVirtual)
	if err != nil {
		return nil, err
	}

	return virtualChangeSet, nil
}

func (s *consensus) sendBlockAddedEvent(block *externalapi.DomainBlock, blockStatus externalapi.BlockStatus) error {
	if s.consensusEventsChan != nil {
		if blockStatus == externalapi.StatusHeaderOnly || blockStatus == externalapi.StatusInvalid {
			return nil
		}

		if len(s.consensusEventsChan) == cap(s.consensusEventsChan) {
			return errors.Errorf("consensusEventsChan is full")
		}
		s.consensusEventsChan <- &externalapi.BlockAdded{Block: block}
	}
	return nil
}

func (s *consensus) sendVirtualChangedEvent(virtualChangeSet *externalapi.VirtualChangeSet, wasVirtualUpdated bool) error {
	if !wasVirtualUpdated || s.consensusEventsChan == nil || virtualChangeSet == nil {
		return nil
	}

	if len(s.consensusEventsChan) == cap(s.consensusEventsChan) {
		return errors.Errorf("consensusEventsChan is full")
	}

	stagingArea := model.NewStagingArea()
	virtualGHOSTDAGData, err := s.ghostdagDataStores[0].Get(s.databaseContext, stagingArea, model.VirtualBlockHash, false)
	if err != nil {
		return err
	}

		if virtualGHOSTDAGData.SelectedParent() == nil {
			return nil
		}
	virtualSelectedParentGHOSTDAGData, err := s.ghostdagDataStores[0].Get(s.databaseContext, stagingArea, virtualGHOSTDAGData.SelectedParent(), false)
	if err != nil {
		return err
	}

	virtualDAAScore, err := s.daaBlocksStore.DAAScore(s.databaseContext, stagingArea, model.VirtualBlockHash)
	if err != nil {
		return err
	}

	virtualChangeSet.VirtualSelectedParentBlueScore = virtualSelectedParentGHOSTDAGData.BlueScore()
	virtualChangeSet.VirtualDAAScore = virtualDAAScore

	s.consensusEventsChan <- virtualChangeSet
	return nil
}

func (s *consensus) ValidateTransactionAndPopulateWithConsensusData(transaction *externalapi.DomainTransaction) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()

	daaScore, err := s.daaBlocksStore.DAAScore(s.databaseContext, stagingArea, model.VirtualBlockHash)
	if err != nil {
		return err
	}

	err = s.transactionValidator.ValidateTransactionInIsolation(transaction, daaScore)
	if err != nil {
		return err
	}

	err = s.consensusStateManager.PopulateTransactionWithUTXOEntries(stagingArea, transaction)
	if err != nil {
		return err
	}

	virtualPastMedianTime, err := s.pastMedianTimeManager.PastMedianTime(stagingArea, model.VirtualBlockHash)
	if err != nil {
		return err
	}

	err = s.transactionValidator.ValidateTransactionInContextIgnoringUTXO(stagingArea, transaction, model.VirtualBlockHash, virtualPastMedianTime)
	if err != nil {
		return err
	}
	return s.transactionValidator.ValidateTransactionInContextAndPopulateFee(
		stagingArea, transaction, model.VirtualBlockHash)
}

func (s *consensus) GetBlock(blockHash *externalapi.DomainHash) (*externalapi.DomainBlock, bool, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()

	block, err := s.blockStore.Block(s.databaseContext, stagingArea, blockHash)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return block, true, nil
}

func (s *consensus) GetBlockEvenIfHeaderOnly(blockHash *externalapi.DomainHash) (*externalapi.DomainBlock, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()

	block, err := s.blockStore.Block(s.databaseContext, stagingArea, blockHash)
	if err == nil {
		return block, nil
	}
	if !errors.Is(err, database.ErrNotFound) {
		return nil, err
	}

	header, err := s.blockHeaderStore.BlockHeader(s.databaseContext, stagingArea, blockHash)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, errors.Wrapf(err, "block %s does not exist", blockHash)
		}
		return nil, err
	}
	return &externalapi.DomainBlock{Header: header}, nil
}

func (s *consensus) GetBlockHeader(blockHash *externalapi.DomainHash) (externalapi.BlockHeader, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()

	blockHeader, err := s.blockHeaderStore.BlockHeader(s.databaseContext, stagingArea, blockHash)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, errors.Wrapf(err, "block header %s does not exist", blockHash)
		}
		return nil, err
	}
	return blockHeader, nil
}

func (s *consensus) GetBlockInfo(blockHash *externalapi.DomainHash) (*externalapi.BlockInfo, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()

	blockInfo := &externalapi.BlockInfo{}

	exists, err := s.blockStatusStore.Exists(s.databaseContext, stagingArea, blockHash)
	if err != nil {
		return nil, err
	}
	blockInfo.Exists = exists
	if !exists {
		return blockInfo, nil
	}

	blockStatus, err := s.blockStatusStore.Get(s.databaseContext, stagingArea, blockHash)
	if err != nil {
		return nil, err
	}
	blockInfo.BlockStatus = blockStatus

	if blockStatus == externalapi.StatusInvalid {
		return blockInfo, nil
	}

	ghostdagData, err := s.ghostdagDataStores[0].Get(s.databaseContext, stagingArea, blockHash, false)
	if err != nil {
		if database.IsNotFoundError(err) {
			return blockInfo, nil
		}
		return nil, err
	}

	blockInfo.BlueScore = ghostdagData.BlueScore()
	blockInfo.BlueWork = ghostdagData.BlueWork()
	blockInfo.SelectedParent = ghostdagData.SelectedParent()
	blockInfo.MergeSetBlues = ghostdagData.MergeSetBlues()
	blockInfo.MergeSetReds = ghostdagData.MergeSetReds()

	return blockInfo, nil
}

func (s *consensus) GetBlockRelations(blockHash *externalapi.DomainHash) (
	parents []*externalapi.DomainHash, children []*externalapi.DomainHash, err error) {

	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()

	blockRelation, err := s.blockRelationStores[0].BlockRelation(s.databaseContext, stagingArea, blockHash)
	if err != nil {
		return nil, nil, err
	}

	return blockRelation.Parents, blockRelation.Children, nil
}

func (s *consensus) GetBlockAcceptanceData(blockHash *externalapi.DomainHash) (externalapi.AcceptanceData, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()

	err := s.validateBlockHashExists(stagingArea, blockHash)
	if err != nil {
		return nil, err
	}

	return s.acceptanceDataStore.Get(s.databaseContext, stagingArea, blockHash)
}

func (s *consensus) GetBlocksAcceptanceData(blockHashes []*externalapi.DomainHash) ([]externalapi.AcceptanceData, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()
	blocksAcceptanceData := make([]externalapi.AcceptanceData, len(blockHashes))

	for i, blockHash := range blockHashes {
		err := s.validateBlockHashExists(stagingArea, blockHash)
		if err != nil {
			return nil, err
		}

		acceptanceData, err := s.acceptanceDataStore.Get(s.databaseContext, stagingArea, blockHash)
		if err != nil {
			return nil, err
		}

		blocksAcceptanceData[i] = acceptanceData
	}

	return blocksAcceptanceData, nil
}

func (s *consensus) GetHashesBetween(lowHash, highHash *externalapi.DomainHash, maxBlocks uint64) (
	hashes []*externalapi.DomainHash, actualHighHash *externalapi.DomainHash, err error) {

	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()

	err = s.validateBlockHashExists(stagingArea, lowHash)
	if err != nil {
		return nil, nil, err
	}
	err = s.validateBlockHashExists(stagingArea, highHash)
	if err != nil {
		return nil, nil, err
	}

	return s.syncManager.GetHashesBetween(stagingArea, lowHash, highHash, maxBlocks)
}

func (s *consensus) GetAnticone(blockHash, contextHash *externalapi.DomainHash,
	maxBlocks uint64) (hashes []*externalapi.DomainHash, err error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()

	err = s.validateBlockHashExists(stagingArea, blockHash)
	if err != nil {
		return nil, err
	}
	err = s.validateBlockHashExists(stagingArea, contextHash)
	if err != nil {
		return nil, err
	}

	return s.syncManager.GetAnticone(stagingArea, blockHash, contextHash, maxBlocks)
}

func (s *consensus) GetMissingBlockBodyHashes(highHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()

	err := s.validateBlockHashExists(stagingArea, highHash)
	if err != nil {
		return nil, err
	}

	return s.syncManager.GetMissingBlockBodyHashes(stagingArea, highHash)
}

func (s *consensus) GetPruningPointUTXOs(expectedPruningPointHash *externalapi.DomainHash,
	fromOutpoint *externalapi.DomainOutpoint, limit int) ([]*externalapi.OutpointAndUTXOEntryPair, error) {

	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()

	pruningPointHash, err := s.pruningStore.PruningPoint(s.databaseContext, stagingArea)
	if err != nil {
		return nil, err
	}

	if !expectedPruningPointHash.Equal(pruningPointHash) {
		return nil, errors.Wrapf(ruleerrors.ErrWrongPruningPointHash, "expected pruning point %s but got %s",
			expectedPruningPointHash,
			pruningPointHash)
	}

	pruningPointUTXOs, err := s.pruningStore.PruningPointUTXOs(s.databaseContext, fromOutpoint, limit)
	if err != nil {
		return nil, err
	}
	return pruningPointUTXOs, nil
}

func (s *consensus) GetVirtualUTXOs(expectedVirtualParents []*externalapi.DomainHash,
	fromOutpoint *externalapi.DomainOutpoint, limit int) ([]*externalapi.OutpointAndUTXOEntryPair, error) {

	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()

	virtualParents, err := s.dagTopologyManagers[0].Parents(stagingArea, model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}

	if !externalapi.HashesEqual(expectedVirtualParents, virtualParents) {
		return nil, errors.Wrapf(ruleerrors.ErrGetVirtualUTXOsWrongVirtualParents, "expected virtual parents %s but got %s",
			expectedVirtualParents,
			virtualParents)
	}

	virtualUTXOs, err := s.consensusStateStore.VirtualUTXOs(s.databaseContext, fromOutpoint, limit)
	if err != nil {
		return nil, err
	}
	return virtualUTXOs, nil
}

func (s *consensus) PruningPoint() (*externalapi.DomainHash, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()

	return s.pruningStore.PruningPoint(s.databaseContext, stagingArea)
}

func (s *consensus) PruningPointHeaders() ([]externalapi.BlockHeader, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()

	lastPruningPointIndex, err := s.pruningStore.CurrentPruningPointIndex(s.databaseContext, stagingArea)
	if err != nil {
		return nil, err
	}

	headers := make([]externalapi.BlockHeader, 0, lastPruningPointIndex)
	for i := uint64(0); i <= lastPruningPointIndex; i++ {
		pruningPoint, err := s.pruningStore.PruningPointByIndex(s.databaseContext, stagingArea, i)
		if err != nil {
			return nil, err
		}

		header, err := s.blockHeaderStore.BlockHeader(s.databaseContext, stagingArea, pruningPoint)
		if err != nil {
			return nil, err
		}

		headers = append(headers, header)
	}

	return headers, nil
}

func (s *consensus) ClearImportedPruningPointData() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.pruningManager.ClearImportedPruningPointData()
}

func (s *consensus) AppendImportedPruningPointUTXOs(outpointAndUTXOEntryPairs []*externalapi.OutpointAndUTXOEntryPair) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.pruningManager.AppendImportedPruningPointUTXOs(outpointAndUTXOEntryPairs)
}

func (s *consensus) ValidateAndInsertImportedPruningPoint(newPruningPoint *externalapi.DomainHash) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.blockProcessor.ValidateAndInsertImportedPruningPoint(newPruningPoint)
}

func (s *consensus) GetVirtualSelectedParent() (*externalapi.DomainHash, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()

	virtualGHOSTDAGData, err := s.ghostdagDataStores[0].Get(s.databaseContext, stagingArea, model.VirtualBlockHash, false)
	if err != nil {
		return nil, err
	}
	return virtualGHOSTDAGData.SelectedParent(), nil
}

func (s *consensus) Tips() ([]*externalapi.DomainHash, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()

	return s.consensusStateStore.Tips(stagingArea, s.databaseContext)
}

func (s *consensus) GetVirtualInfo() (*externalapi.VirtualInfo, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()

	blockRelations, err := s.blockRelationStores[0].BlockRelation(s.databaseContext, stagingArea, model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}
	bits, err := s.difficultyManager.RequiredDifficulty(stagingArea, model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}
	pastMedianTime, err := s.pastMedianTimeManager.PastMedianTime(stagingArea, model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}
	virtualGHOSTDAGData, err := s.ghostdagDataStores[0].Get(s.databaseContext, stagingArea, model.VirtualBlockHash, false)
	if err != nil {
		return nil, err
	}

	daaScore, err := s.daaBlocksStore.DAAScore(s.databaseContext, stagingArea, model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}

	return &externalapi.VirtualInfo{
		ParentHashes:   blockRelations.Parents,
		Bits:           bits,
		PastMedianTime: pastMedianTime,
		BlueScore:      virtualGHOSTDAGData.BlueScore(),
		DAAScore:       daaScore,
	}, nil
}

func (s *consensus) GetVirtualDAAScore() (uint64, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()

	return s.daaBlocksStore.DAAScore(s.databaseContext, stagingArea, model.VirtualBlockHash)
}

func (s *consensus) CreateBlockLocatorFromPruningPoint(highHash *externalapi.DomainHash, limit uint32) (externalapi.BlockLocator, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()

	err := s.validateBlockHashExists(stagingArea, highHash)
	if err != nil {
		return nil, err
	}

	pruningPoint, err := s.pruningStore.PruningPoint(s.databaseContext, stagingArea)
	if err != nil {
		return nil, err
	}

	return s.syncManager.CreateBlockLocator(stagingArea, pruningPoint, highHash, limit)
}

func (s *consensus) CreateFullHeadersSelectedChainBlockLocator() (externalapi.BlockLocator, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()

	lowHash, err := s.pruningStore.PruningPoint(s.databaseContext, stagingArea)
	if err != nil {
		return nil, err
	}

	highHash, err := s.headersSelectedTipStore.HeadersSelectedTip(s.databaseContext, stagingArea)
	if err != nil {
		return nil, err
	}

	return s.syncManager.CreateHeadersSelectedChainBlockLocator(stagingArea, lowHash, highHash)
}

func (s *consensus) CreateHeadersSelectedChainBlockLocator(lowHash, highHash *externalapi.DomainHash) (externalapi.BlockLocator, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()

	return s.syncManager.CreateHeadersSelectedChainBlockLocator(stagingArea, lowHash, highHash)
}

func (s *consensus) GetSyncInfo() (*externalapi.SyncInfo, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()

	return s.syncManager.GetSyncInfo(stagingArea)
}

func (s *consensus) IsValidPruningPoint(blockHash *externalapi.DomainHash) (bool, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()

	err := s.validateBlockHashExists(stagingArea, blockHash)
	if err != nil {
		return false, err
	}

	return s.pruningManager.IsValidPruningPoint(stagingArea, blockHash)
}

func (s *consensus) ArePruningPointsViolatingFinality(pruningPoints []externalapi.BlockHeader) (bool, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()

	return s.pruningManager.ArePruningPointsViolatingFinality(stagingArea, pruningPoints)
}

func (s *consensus) ImportPruningPoints(pruningPoints []externalapi.BlockHeader) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()
	err := s.consensusStateManager.ImportPruningPoints(stagingArea, pruningPoints)
	if err != nil {
		return err
	}

	err = staging.CommitAllChanges(s.databaseContext, stagingArea)
	if err != nil {
		return err
	}

	return nil
}

func (s *consensus) GetVirtualSelectedParentChainFromBlock(blockHash *externalapi.DomainHash) (*externalapi.SelectedChainPath, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()

	err := s.validateBlockHashExists(stagingArea, blockHash)
	if err != nil {
		return nil, err
	}

	return s.consensusStateManager.GetVirtualSelectedParentChainFromBlock(stagingArea, blockHash)
}

func (s *consensus) validateBlockHashExists(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) error {
	status, err := s.blockStatusStore.Get(s.databaseContext, stagingArea, blockHash)
	if database.IsNotFoundError(err) {
		return errors.Errorf("block %s does not exist", blockHash)
	}
	if err != nil {
		return err
	}

	if status == externalapi.StatusInvalid {
		return errors.Errorf("block %s is invalid", blockHash)
	}
	return nil
}

func (s *consensus) IsInSelectedParentChainOf(blockHashA *externalapi.DomainHash, blockHashB *externalapi.DomainHash) (bool, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()

	err := s.validateBlockHashExists(stagingArea, blockHashA)
	if err != nil {
		return false, err
	}
	err = s.validateBlockHashExists(stagingArea, blockHashB)
	if err != nil {
		return false, err
	}

	return s.dagTopologyManagers[0].IsInSelectedParentChainOf(stagingArea, blockHashA, blockHashB)
}

func (s *consensus) GetHeadersSelectedTip() (*externalapi.DomainHash, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()

	return s.headersSelectedTipStore.HeadersSelectedTip(s.databaseContext, stagingArea)
}

func (s *consensus) Anticone(blockHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()

	err := s.validateBlockHashExists(stagingArea, blockHash)
	if err != nil {
		return nil, err
	}

	tips, err := s.consensusStateStore.Tips(stagingArea, s.databaseContext)
	if err != nil {
		return nil, err
	}

	return s.dagTraversalManager.AnticoneFromBlocks(stagingArea, tips, blockHash, 0)
}

func (s *consensus) EstimateNetworkHashesPerSecond(startHash *externalapi.DomainHash, windowSize int) (uint64, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.difficultyManager.EstimateNetworkHashesPerSecond(startHash, windowSize)
}

func (s *consensus) PopulateMass(transaction *externalapi.DomainTransaction) {
	s.transactionValidator.PopulateMass(transaction)
}

func (s *consensus) ResolveVirtual(progressReportCallback func(uint64, uint64)) error {
	virtualDAAScoreStart, err := s.GetVirtualDAAScore()
	if err != nil {
		return err
	}

	for i := 0; ; i++ {
		if i%10 == 0 && progressReportCallback != nil {
			virtualDAAScore, err := s.GetVirtualDAAScore()
			if err != nil {
				return err
			}
			progressReportCallback(virtualDAAScoreStart, virtualDAAScore)
		}

		_, isCompletelyResolved, err := s.resolveVirtualChunkWithLock(virtualResolveChunk)
		if err != nil {
			return err
		}
		if isCompletelyResolved {
			break
		}
	}

	return nil
}

func (s *consensus) resolveVirtualChunkWithLock(maxBlocksToResolve uint64) (*externalapi.VirtualChangeSet, bool, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.resolveVirtualChunkNoLock(maxBlocksToResolve)
}

func (s *consensus) resolveVirtualChunkNoLock(maxBlocksToResolve uint64) (*externalapi.VirtualChangeSet, bool, error) {
	virtualChangeSet, isCompletelyResolved, err := s.consensusStateManager.ResolveVirtual(maxBlocksToResolve)
	if err != nil {
		return nil, false, err
	}
	s.virtualNotUpdated = !isCompletelyResolved

	stagingArea := model.NewStagingArea()
	err = s.pruningManager.UpdatePruningPointByVirtual(stagingArea)
	if err != nil {
		return nil, false, err
	}

	err = staging.CommitAllChanges(s.databaseContext, stagingArea)
	if err != nil {
		return nil, false, err
	}

	err = s.pruningManager.UpdatePruningPointIfRequired()
	if err != nil {
		return nil, false, err
	}

	err = s.sendVirtualChangedEvent(virtualChangeSet, true)
	if err != nil {
		return nil, false, err
	}

	return virtualChangeSet, isCompletelyResolved, nil
}

func (s *consensus) BuildPruningPointProof() (*externalapi.PruningPointProof, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.pruningProofManager.BuildPruningPointProof(model.NewStagingArea())
}

func (s *consensus) ValidatePruningPointProof(pruningPointProof *externalapi.PruningPointProof) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	log.Infof("Validating the pruning point proof")
	err := s.pruningProofManager.ValidatePruningPointProof(pruningPointProof)
	if err != nil {
		return err
	}

	log.Infof("Done validating the pruning point proof")
	return nil
}

func (s *consensus) ApplyPruningPointProof(pruningPointProof *externalapi.PruningPointProof) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	log.Infof("Applying the pruning point proof")
	err := s.pruningProofManager.ApplyPruningPointProof(pruningPointProof)
	if err != nil {
		return err
	}

	log.Infof("Done applying the pruning point proof")
	return nil
}

func (s *consensus) BlockDAAWindowHashes(blockHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()
	return s.dagTraversalManager.DAABlockWindow(stagingArea, blockHash)
}

func (s *consensus) TrustedDataDataDAAHeader(trustedBlockHash, daaBlockHash *externalapi.DomainHash, daaBlockWindowIndex uint64) (*externalapi.TrustedDataDataDAAHeader, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()
	header, err := s.blockHeaderStore.BlockHeader(s.databaseContext, stagingArea, daaBlockHash)
	if err != nil {
		return nil, err
	}

	ghostdagData, err := s.ghostdagDataStores[0].Get(s.databaseContext, stagingArea, daaBlockHash, false)
	isNotFoundError := database.IsNotFoundError(err)
	if !isNotFoundError && err != nil {
		return nil, err
	}

	if !isNotFoundError {
		return &externalapi.TrustedDataDataDAAHeader{
			Header:       header,
			GHOSTDAGData: ghostdagData,
		}, nil
	}

	ghostdagDataHashPair, err := s.blocksWithTrustedDataDAAWindowStore.DAAWindowBlock(s.databaseContext, stagingArea, trustedBlockHash, daaBlockWindowIndex)
	if err != nil {
		return nil, err
	}

	return &externalapi.TrustedDataDataDAAHeader{
		Header:       header,
		GHOSTDAGData: ghostdagDataHashPair.GHOSTDAGData,
	}, nil
}

func (s *consensus) TrustedBlockAssociatedGHOSTDAGDataBlockHashes(blockHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.pruningManager.TrustedBlockAssociatedGHOSTDAGDataBlockHashes(model.NewStagingArea(), blockHash)
}

func (s *consensus) TrustedGHOSTDAGData(blockHash *externalapi.DomainHash) (*externalapi.BlockGHOSTDAGData, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()
	ghostdagData, err := s.ghostdagDataStores[0].Get(s.databaseContext, stagingArea, blockHash, false)
	isNotFoundError := database.IsNotFoundError(err)
	if isNotFoundError || ghostdagData.SelectedParent().Equal(model.VirtualGenesisBlockHash) {
		return s.ghostdagDataStores[0].Get(s.databaseContext, stagingArea, blockHash, true)
	}

	return ghostdagData, nil
}

func (s *consensus) IsChainBlock(blockHash *externalapi.DomainHash) (bool, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()
	virtualGHOSTDAGData, err := s.ghostdagDataStores[0].Get(s.databaseContext, stagingArea, model.VirtualBlockHash, false)
	if err != nil {
		return false, err
	}

	return s.dagTopologyManagers[0].IsInSelectedParentChainOf(stagingArea, blockHash, virtualGHOSTDAGData.SelectedParent())
}

func (s *consensus) VirtualMergeDepthRoot() (*externalapi.DomainHash, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	stagingArea := model.NewStagingArea()
	return s.mergeDepthManager.VirtualMergeDepthRoot(stagingArea)
}

func (s *consensus) IsNearlySynced() (bool, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.isNearlySyncedNoLock()
}

func (s *consensus) isNearlySyncedNoLock() (bool, error) {
	stagingArea := model.NewStagingArea()
	virtualGHOSTDAGData, err := s.ghostdagDataStores[0].Get(s.databaseContext, stagingArea, model.VirtualBlockHash, false)
	if err != nil {
		return false, err
	}

	selectedParent := virtualGHOSTDAGData.SelectedParent()
	if selectedParent == nil || selectedParent.Equal(s.genesisHash) {
		return false, nil
	}

	virtualSelectedParentHeader, err := s.blockHeaderStore.BlockHeader(s.databaseContext, stagingArea, virtualGHOSTDAGData.SelectedParent())
	if err != nil {
		return false, err
	}

	now := mstime.Now().UnixMilliseconds()
	if now-virtualSelectedParentHeader.TimeInMilliseconds() < s.expectedDAAWindowDurationInMilliseconds {
		log.Debugf("The selected tip timestamp is recent (%d), so IsNearlySynced returns true",
			virtualSelectedParentHeader.TimeInMilliseconds())
		return true, nil
	}

	log.Debugf("The selected tip timestamp is old (%d), so IsNearlySynced returns false",
		virtualSelectedParentHeader.TimeInMilliseconds())
	return false, nil
}

func (s *consensus) GetAddressLevel(address string) (uint8, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.addressLevelStore == nil {
		return 1, nil
	}
	return s.addressLevelStore.Get(s.databaseContext, address)
}
