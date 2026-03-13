package blockbuilder

import (
	"encoding/binary"
	"math/big"
	"github.com/rupixnet/rupixd/util/difficulty"
	"sort"
	"fmt"

	"github.com/rupixnet/rupixd/domain/consensus/ruleerrors"
	"github.com/rupixnet/rupixd/domain/consensus/utils/blockheader"
	"github.com/pkg/errors"

	"github.com/rupixnet/rupixd/domain/consensus/model"
	"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
	"github.com/rupixnet/rupixd/domain/consensus/utils/consensushashing"
	"github.com/rupixnet/rupixd/domain/consensus/utils/constants"
	"github.com/rupixnet/rupixd/domain/consensus/utils/merkle"
	"github.com/rupixnet/rupixd/infrastructure/logger"
	"github.com/rupixnet/rupixd/util/mstime"
	"github.com/rupixnet/rupixd/infrastructure/db/database"
)

type blockBuilder struct {
	databaseContext model.DBManager
	genesisHash     *externalapi.DomainHash

	difficultyManager     model.DifficultyManager
	pastMedianTimeManager model.PastMedianTimeManager
	coinbaseManager       model.CoinbaseManager
	consensusStateManager model.ConsensusStateManager
	ghostdagManager       model.GHOSTDAGManager
	transactionValidator  model.TransactionValidator
	finalityManager       model.FinalityManager
	pruningManager        model.PruningManager
	blockParentBuilder    model.BlockParentBuilder

	acceptanceDataStore model.AcceptanceDataStore
	blockRelationStore  model.BlockRelationStore
	multisetStore       model.MultisetStore
	ghostdagDataStore   model.GHOSTDAGDataStore
	daaBlocksStore      model.DAABlocksStore
	blockHeaderStore   model.BlockHeaderStore
}

// New creates a new instance of a BlockBuilder
func New(
	databaseContext model.DBManager,
	genesisHash *externalapi.DomainHash,

	difficultyManager model.DifficultyManager,
	pastMedianTimeManager model.PastMedianTimeManager,
	coinbaseManager model.CoinbaseManager,
	consensusStateManager model.ConsensusStateManager,
	ghostdagManager model.GHOSTDAGManager,
	transactionValidator model.TransactionValidator,
	finalityManager model.FinalityManager,
	blockParentBuilder model.BlockParentBuilder,
	pruningManager model.PruningManager,

	acceptanceDataStore model.AcceptanceDataStore,
	blockRelationStore model.BlockRelationStore,
	multisetStore model.MultisetStore,
	ghostdagDataStore model.GHOSTDAGDataStore,
	daaBlocksStore model.DAABlocksStore,
	blockHeaderStore model.BlockHeaderStore,
) model.BlockBuilder {

	return &blockBuilder{
		databaseContext: databaseContext,
		genesisHash:     genesisHash,

		difficultyManager:     difficultyManager,
		pastMedianTimeManager: pastMedianTimeManager,
		coinbaseManager:       coinbaseManager,
		consensusStateManager: consensusStateManager,
		ghostdagManager:       ghostdagManager,
		transactionValidator:  transactionValidator,
		finalityManager:       finalityManager,
		blockParentBuilder:    blockParentBuilder,
		pruningManager:        pruningManager,

		acceptanceDataStore: acceptanceDataStore,
		blockRelationStore:  blockRelationStore,
		multisetStore:       multisetStore,
		ghostdagDataStore:   ghostdagDataStore,
		daaBlocksStore:      daaBlocksStore,
            blockHeaderStore:   blockHeaderStore,
	}
}

// BuildBlock builds a block over the current state, with the given
// coinbaseData and the given transactions
func (bb *blockBuilder) BuildBlock(coinbaseData *externalapi.DomainCoinbaseData,
	transactions []*externalapi.DomainTransaction) (block *externalapi.DomainBlock, coinbaseHasRedReward bool, err error) {

	onEnd := logger.LogAndMeasureExecutionTime(log, "BuildBlock")
	defer onEnd()

	stagingArea := model.NewStagingArea()

	return bb.buildBlock(stagingArea, coinbaseData, transactions)
}

func (bb *blockBuilder) buildBlock(stagingArea *model.StagingArea, coinbaseData *externalapi.DomainCoinbaseData,
	transactions []*externalapi.DomainTransaction) (block *externalapi.DomainBlock, coinbaseHasRedReward bool, err error) {

	fmt.Println("buildBlock: validateTransactions")
	err = bb.validateTransactions(stagingArea, transactions)
	if err != nil {
		return nil, false, err
	}

	fmt.Println("buildBlock: newBlockPruningPoint")
	newBlockPruningPoint, err := bb.newBlockPruningPoint(stagingArea, model.VirtualBlockHash)
	if err != nil {
		return nil, false, err
	}

	fmt.Println("buildBlock: newBlockCoinbaseTransaction")
	coinbase, coinbaseHasRedReward, err := bb.newBlockCoinbaseTransaction(stagingArea, coinbaseData)
	if err != nil {
		return nil, false, err
	}
	transactionsWithCoinbase := append([]*externalapi.DomainTransaction{coinbase}, transactions...)

	fmt.Println("buildBlock: buildHeader")
	header, err := bb.buildHeader(stagingArea, transactionsWithCoinbase, newBlockPruningPoint)
	if err != nil {
		return nil, false, err
	}

	return &externalapi.DomainBlock{
		Header:       header,
		Transactions: transactionsWithCoinbase,
	}, coinbaseHasRedReward, nil
}

func (bb *blockBuilder) validateTransactions(stagingArea *model.StagingArea,
	transactions []*externalapi.DomainTransaction) error {

	invalidTransactions := make([]ruleerrors.InvalidTransaction, 0)
	for _, transaction := range transactions {
		err := bb.validateTransaction(stagingArea, transaction)
		if err != nil {
			ruleError := ruleerrors.RuleError{}
			if !errors.As(err, &ruleError) {
				return err
			}
			invalidTransactions = append(invalidTransactions,
				ruleerrors.InvalidTransaction{Transaction: transaction, Error: &ruleError})
		}
	}

	if len(invalidTransactions) > 0 {
		return ruleerrors.NewErrInvalidTransactionsInNewBlock(invalidTransactions)
	}

	return nil
}

func (bb *blockBuilder) validateTransaction(
	stagingArea *model.StagingArea, transaction *externalapi.DomainTransaction) error {

	originalEntries := make([]externalapi.UTXOEntry, len(transaction.Inputs))
	for i, input := range transaction.Inputs {
		originalEntries[i] = input.UTXOEntry
		input.UTXOEntry = nil
	}

	defer func() {
		for i, input := range transaction.Inputs {
			input.UTXOEntry = originalEntries[i]
		}
	}()

	err := bb.consensusStateManager.PopulateTransactionWithUTXOEntries(stagingArea, transaction)
	if err != nil {
		return err
	}

	virtualPastMedianTime, err := bb.pastMedianTimeManager.PastMedianTime(stagingArea, model.VirtualBlockHash)
	if err != nil {
		return err
	}

	err = bb.transactionValidator.ValidateTransactionInContextIgnoringUTXO(stagingArea, transaction, model.VirtualBlockHash, virtualPastMedianTime)
	if err != nil {
		return err
	}

	return bb.transactionValidator.ValidateTransactionInContextAndPopulateFee(stagingArea, transaction, model.VirtualBlockHash)
}

func (bb *blockBuilder) newBlockCoinbaseTransaction(stagingArea *model.StagingArea,
    coinbaseData *externalapi.DomainCoinbaseData) (expectedTransaction *externalapi.DomainTransaction, hasRedReward bool, err error) {

    coinbaseTx, hasRedReward, err := bb.coinbaseManager.ExpectedCoinbaseTransaction(stagingArea, model.VirtualBlockHash, coinbaseData)
    if err != nil {
        return nil, false, err
    }

    // Si el Virtual no tiene selectedParent, estamos en el primer bloque hijo del genesis.
    // El payload tiene BlueScore=0 (del Virtual) pero debe ser 1.
    virtualGHOSTDAGData, err := bb.ghostdagDataStore.Get(bb.databaseContext, stagingArea, model.VirtualBlockHash, false)
    if err != nil {
        return nil, false, err
    }
    if virtualGHOSTDAGData.SelectedParent() == nil {
        // Sobreescribir los primeros 8 bytes del payload con BlueScore=1
        binary.LittleEndian.PutUint64(coinbaseTx.Payload[:8], 1)
    }

    return coinbaseTx, hasRedReward, nil
}

func (bb *blockBuilder) buildHeader(stagingArea *model.StagingArea, transactions []*externalapi.DomainTransaction,
	newBlockPruningPoint *externalapi.DomainHash) (externalapi.BlockHeader, error) {

	fmt.Println("buildHeader: newBlockDAAScore")
	daaScore, err := bb.newBlockDAAScore(stagingArea)
	if err != nil {
		return nil, err
	}

	fmt.Println("buildHeader: newBlockParents")
	parents, err := bb.newBlockParents(stagingArea, daaScore)
	if err != nil {
		return nil, err
	}

fmt.Println("buildHeader: newBlockTime START")
timeInMilliseconds, err := bb.newBlockTime(stagingArea)
fmt.Printf("buildHeader: newBlockTime END err=%v\n", err)
if err != nil {
    return nil, err
}
	fmt.Println("buildHeader: newBlockDifficulty")
	bits, err := bb.newBlockDifficulty(stagingArea)
	if err != nil {
		return nil, err
	}

	hashMerkleRoot := bb.newBlockHashMerkleRoot(transactions)

	fmt.Println("buildHeader: newBlockAcceptedIDMerkleRoot")
	acceptedIDMerkleRoot, err := bb.newBlockAcceptedIDMerkleRoot(stagingArea)
	if err != nil {
		return nil, err
	}

	fmt.Println("buildHeader: newBlockUTXOCommitment")
	utxoCommitment, err := bb.newBlockUTXOCommitment(stagingArea)
	if err != nil {
		return nil, err
	}

	fmt.Println("buildHeader: newBlockBlueWork")
	blueWork, err := bb.newBlockBlueWork(stagingArea)
	if err != nil {
		return nil, err
	}

	fmt.Println("buildHeader: newBlockBlueScore")
	blueScore, err := bb.newBlockBlueScore(stagingArea)
	if err != nil {
		return nil, err
	}

	fmt.Println("buildHeader: done")
	return blockheader.NewImmutableBlockHeader(
		constants.BlockVersion,
		parents,
		hashMerkleRoot,
		acceptedIDMerkleRoot,
		utxoCommitment,
		timeInMilliseconds,
		bits,
		0,
		daaScore,
		blueScore,
		blueWork,
		newBlockPruningPoint,
	), nil
}

func (bb *blockBuilder) newBlockParents(stagingArea *model.StagingArea, daaScore uint64) ([]externalapi.BlockLevelParents, error) {
        virtualBlockRelations, err := bb.blockRelationStore.BlockRelation(bb.databaseContext, stagingArea, model.VirtualBlockHash)
        if err != nil {
                return nil, err
        }
        fmt.Printf("newBlockParents: virtualBlockRelations.Parents = %v\n", virtualBlockRelations.Parents)
        parents, err := bb.blockParentBuilder.BuildParents(stagingArea, daaScore, virtualBlockRelations.Parents)
        fmt.Printf("newBlockParents: BuildParents returned err=%v parents=%v\n", err, parents)
        return parents, err
}

func (bb *blockBuilder) newBlockTime(stagingArea *model.StagingArea) (int64, error) {
	// The timestamp for the block must not be before the median timestamp
	// of the last several blocks. Thus, choose the maximum between the
	// current time and one second after the past median time. The current
	// timestamp is truncated to a millisecond boundary before comparison since a
	// block timestamp does not supported a precision greater than one
	// millisecond.
	newTimestamp := mstime.Now().UnixMilliseconds()
	minTimestamp, err := bb.minBlockTime(stagingArea, model.VirtualBlockHash)
	if err != nil {
		return 0, err
	}
	if newTimestamp < minTimestamp {
		newTimestamp = minTimestamp
	}
	return newTimestamp, nil
}

func (bb *blockBuilder) minBlockTime(stagingArea *model.StagingArea, hash *externalapi.DomainHash) (int64, error) {
	pastMedianTime, err := bb.pastMedianTimeManager.PastMedianTime(stagingArea, hash)
	if err != nil {
		return 0, err
	}

	return pastMedianTime + 1, nil
}

func (bb *blockBuilder) newBlockDifficulty(stagingArea *model.StagingArea) (uint32, error) {
	return bb.difficultyManager.RequiredDifficulty(stagingArea, model.VirtualBlockHash)
}

func (bb *blockBuilder) newBlockHashMerkleRoot(transactions []*externalapi.DomainTransaction) *externalapi.DomainHash {
	return merkle.CalculateHashMerkleRoot(transactions)
}

func (bb *blockBuilder) newBlockAcceptedIDMerkleRoot(stagingArea *model.StagingArea) (*externalapi.DomainHash, error) {
        newBlockAcceptanceData, err := bb.acceptanceDataStore.Get(bb.databaseContext, stagingArea, model.VirtualBlockHash)
        if err != nil {
                if database.IsNotFoundError(err) {
                        return merkle.CalculateIDMerkleRoot(nil), nil
                }
                return nil, err
        }

        return bb.calculateAcceptedIDMerkleRoot(newBlockAcceptanceData)
}

func (bb *blockBuilder) calculateAcceptedIDMerkleRoot(acceptanceData externalapi.AcceptanceData) (*externalapi.DomainHash, error) {
	var acceptedTransactions []*externalapi.DomainTransaction
	for _, blockAcceptanceData := range acceptanceData {
		for _, transactionAcceptance := range blockAcceptanceData.TransactionAcceptanceData {
			if !transactionAcceptance.IsAccepted {
				continue
			}
			acceptedTransactions = append(acceptedTransactions, transactionAcceptance.Transaction)
		}
	}
	sort.Slice(acceptedTransactions, func(i, j int) bool {
		acceptedTransactionIID := consensushashing.TransactionID(acceptedTransactions[i])
		acceptedTransactionJID := consensushashing.TransactionID(acceptedTransactions[j])
		return acceptedTransactionIID.Less(acceptedTransactionJID)
	})

	return merkle.CalculateIDMerkleRoot(acceptedTransactions), nil
}

func (bb *blockBuilder) newBlockUTXOCommitment(stagingArea *model.StagingArea) (*externalapi.DomainHash, error) {
	newBlockMultiset, err := bb.multisetStore.Get(bb.databaseContext, stagingArea, model.VirtualBlockHash)
	if err != nil {
		return nil, err
	}
	newBlockUTXOCommitment := newBlockMultiset.Hash()
	return newBlockUTXOCommitment, nil
}

func (bb *blockBuilder) newBlockDAAScore(stagingArea *model.StagingArea) (uint64, error) {
	return bb.daaBlocksStore.DAAScore(bb.databaseContext, stagingArea, model.VirtualBlockHash)
}

func (bb *blockBuilder) newBlockBlueWork(stagingArea *model.StagingArea) (*big.Int, error) {
    virtualGHOSTDAGData, err := bb.ghostdagDataStore.Get(bb.databaseContext, stagingArea, model.VirtualBlockHash, false)
    if err != nil {
        return nil, err
    }
    selectedParent := virtualGHOSTDAGData.SelectedParent()

 if selectedParent == nil {
    genesisHeader, err := bb.blockHeaderStore.BlockHeader(bb.databaseContext, stagingArea, bb.genesisHash)
    if err != nil {
        return nil, err
    }
    result := difficulty.CalcWork(genesisHeader.Bits())
    fmt.Printf("DEBUG newBlockBlueWork: genesis path result=%v\n", result)
    return result, nil
}
    selectedParentGHOSTDAGData, err := bb.ghostdagDataStore.Get(bb.databaseContext, stagingArea, selectedParent, false)
    if err != nil {
        return nil, err
    }

    // Heredar blueWork del selectedParent (igual que GHOSTDAG)
    blueWork := new(big.Int).Set(selectedParentGHOSTDAGData.BlueWork())

    // Sumar CalcWork de cada blue en mergeSetBlues del Virtual (igual que GHOSTDAG)
    for _, blue := range virtualGHOSTDAGData.MergeSetBlues() {
        if blue.Equal(model.VirtualGenesisBlockHash) {
            continue
        }
        header, err := bb.blockHeaderStore.BlockHeader(bb.databaseContext, stagingArea, blue)
        if err != nil {
            return nil, err
        }
        blueWork.Add(blueWork, difficulty.CalcWork(header.Bits()))
    }

    fmt.Printf("DEBUG newBlockBlueWork: result=%v\n", blueWork)
    return blueWork, nil
}

func (bb *blockBuilder) newBlockBlueScore(stagingArea *model.StagingArea) (uint64, error) {
    virtualGHOSTDAGData, err := bb.ghostdagDataStore.Get(bb.databaseContext, stagingArea, model.VirtualBlockHash, false)
    if err != nil {
        return 0, err
    }

    // Si el Virtual no tiene selectedParent, estamos construyendo el primer
    // hijo del genesis. GHOSTDAG calculará BlueScore = genesis.BlueScore(0) + len(mergeSetBlues=1) = 1
    if virtualGHOSTDAGData.SelectedParent() == nil {
        return 1, nil
    }

    return virtualGHOSTDAGData.BlueScore(), nil
}

func (bb *blockBuilder) newBlockPruningPoint(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (*externalapi.DomainHash, error) {
	return bb.pruningManager.ExpectedHeaderPruningPoint(stagingArea, blockHash)
}

