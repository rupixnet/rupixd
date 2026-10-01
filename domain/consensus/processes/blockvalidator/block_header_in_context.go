package blockvalidator

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/rupixnet/rupixd/domain/consensus/model"
	"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
	"github.com/rupixnet/rupixd/domain/consensus/ruleerrors"
	"github.com/rupixnet/rupixd/domain/consensus/utils/consensushashing"
	"github.com/rupixnet/rupixd/domain/dagconfig"
	"github.com/rupixnet/rupixd/infrastructure/logger"
)

// ValidateHeaderInContext validates block headers in the context of the current
// consensus state
func (v *blockValidator) ValidateHeaderInContext(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash, isBlockWithTrustedData bool) error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "ValidateHeaderInContext")
	defer onEnd()

	header, err := v.blockHeaderStore.BlockHeader(v.databaseContext, stagingArea, blockHash)
	if err != nil {
		return err
	}

	hasValidatedHeader, err := v.hasValidatedHeader(stagingArea, blockHash)
	if err != nil {
		return err
	}

	if !hasValidatedHeader {
		var logErr error
		log.Debug(logger.NewLogClosure(func() string {
			var ghostdagData *externalapi.BlockGHOSTDAGData
			ghostdagData, logErr = v.ghostdagDataStores[0].Get(v.databaseContext, stagingArea, blockHash, false)
			if err != nil {
				return ""
			}

			return fmt.Sprintf("block %s blue score is %d", blockHash, ghostdagData.BlueScore())
		}))

		if logErr != nil {
			return logErr
		}
	}

	err = v.validateMedianTime(stagingArea, header)
	if err != nil {
		return err
	}

	err = v.checkMergeSizeLimit(stagingArea, blockHash)
	if err != nil {
		return err
	}

	// If needed - calculate reachability data right before calling CheckBoundedMergeDepth,
	// since it's used to find a block's finality point.
	// This might not be required if this block's header has previously been received during
	// headers-first synchronization.
	hasReachabilityData, err := v.reachabilityStore.HasReachabilityData(v.databaseContext, stagingArea, blockHash)
	if err != nil {
		return err
	}
	if !hasReachabilityData {
		err = v.reachabilityManager.AddBlock(stagingArea, blockHash)
		if err != nil {
			return err
		}
	}

	if !isBlockWithTrustedData {
		err = v.checkIndirectParents(stagingArea, header)
		if err != nil {
			return err
		}
	}

	err = v.mergeDepthManager.CheckBoundedMergeDepth(stagingArea, blockHash, isBlockWithTrustedData)
	if err != nil {
		return err
	}

	err = v.checkDAAScore(stagingArea, blockHash, header)
	if err != nil {
		return err
	}

	// Rupix: checkpoint temporal. Todo bloque con blue score >= X + MergeDepth debe
	// tener al bloque canonico H en su pasado. Defensa contra reorganizaciones
	// profundas mientras el hashrate es bajo. Caduca en checkpointsExpireDAAScore.
	err = v.checkCheckpoint(stagingArea, blockHash, header)
	if err != nil {
		return err
	}

	err = v.checkBlueWork(stagingArea, blockHash, header)
	if err != nil {
		return err
	}

	err = v.checkHeaderBlueScore(stagingArea, blockHash, header)
	if err != nil {
		return err
	}

	if !isBlockWithTrustedData {
		err = v.validateHeaderPruningPoint(stagingArea, blockHash)
		if err != nil {
			return err
		}
	}

	return nil
}

func (v *blockValidator) hasValidatedHeader(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (bool, error) {
	exists, err := v.blockStatusStore.Exists(v.databaseContext, stagingArea, blockHash)
	if err != nil {
		return false, err
	}

	if !exists {
		return false, nil
	}

	status, err := v.blockStatusStore.Get(v.databaseContext, stagingArea, blockHash)
	if err != nil {
		return false, err
	}

	return status == externalapi.StatusHeaderOnly, nil
}

// checkParentsIncest validates that no parent is an ancestor of another parent
func (v *blockValidator) checkParentsIncest(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) error {
	parents, err := v.dagTopologyManagers[0].Parents(stagingArea, blockHash)
	if err != nil {
		return err
	}

	for _, parentA := range parents {
		for _, parentB := range parents {
			if parentA.Equal(parentB) {
				continue
			}

			isAAncestorOfB, err := v.dagTopologyManagers[0].IsAncestorOf(stagingArea, parentA, parentB)
			if err != nil {
				return err
			}

			if isAAncestorOfB {
				return errors.Wrapf(ruleerrors.ErrInvalidParentsRelation, "parent %s is an "+
					"ancestor of another parent %s",
					parentA,
					parentB,
				)
			}
		}
	}
	return nil
}

func (v *blockValidator) validateMedianTime(stagingArea *model.StagingArea, header externalapi.BlockHeader) error {
	if len(header.DirectParents()) == 0 {
		return nil
	}

	// Ensure the timestamp for the block header is not before the
	// median time of the last several blocks (medianTimeBlocks).
	hash := consensushashing.HeaderHash(header)
	pastMedianTime, err := v.pastMedianTimeManager.PastMedianTime(stagingArea, hash)
	if err != nil {
		return err
	}

	if header.TimeInMilliseconds() <= pastMedianTime {
		return errors.Wrapf(ruleerrors.ErrTimeTooOld, "block timestamp of %d is not after expected %d",
			header.TimeInMilliseconds(), pastMedianTime)
	}

	return nil
}

func (v *blockValidator) checkMergeSizeLimit(stagingArea *model.StagingArea, hash *externalapi.DomainHash) error {
	ghostdagData, err := v.ghostdagDataStores[0].Get(v.databaseContext, stagingArea, hash, false)
	if err != nil {
		return err
	}

	mergeSetSize := len(ghostdagData.MergeSetBlues()) + len(ghostdagData.MergeSetReds())

	if uint64(mergeSetSize) > v.mergeSetSizeLimit {
		return errors.Wrapf(ruleerrors.ErrViolatingMergeLimit,
			"The block merges %d blocks > %d merge set size limit", mergeSetSize, v.mergeSetSizeLimit)
	}

	return nil
}

func (v *blockValidator) checkIndirectParents(stagingArea *model.StagingArea, header externalapi.BlockHeader) error {
	expectedParents, err := v.blockParentBuilder.BuildParents(stagingArea, header.DAAScore(), header.DirectParents())
	if err != nil {
		return err
	}

	areParentsEqual := externalapi.ParentsEqual(header.Parents(), expectedParents)
	if !areParentsEqual {
		return errors.Wrapf(ruleerrors.ErrUnexpectedParents, "unexpected indirect block parents")
	}
	return nil
}

func (v *blockValidator) checkDAAScore(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash,
	header externalapi.BlockHeader) error {

	expectedDAAScore, err := v.daaBlocksStore.DAAScore(v.databaseContext, stagingArea, blockHash)
	if err != nil {
		return err
	}
	if header.DAAScore() != expectedDAAScore {
		return errors.Wrapf(ruleerrors.ErrUnexpectedDAAScore, "block DAA score of %d is not the expected value of %d", header.DAAScore(), expectedDAAScore)
	}
	return nil
}

func (v *blockValidator) checkBlueWork(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash,
	header externalapi.BlockHeader) error {

	ghostdagData, err := v.ghostdagDataStores[0].Get(v.databaseContext, stagingArea, blockHash, false)
	if err != nil {
		return err
	}
	expectedBlueWork := ghostdagData.BlueWork()
	if header.BlueWork().Cmp(expectedBlueWork) != 0 {
		return errors.Wrapf(ruleerrors.ErrUnexpectedBlueWork, "block blue work of %d is not the expected value of %d", header.BlueWork(), expectedBlueWork)
	}
	return nil
}

func (v *blockValidator) checkHeaderBlueScore(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash,
	header externalapi.BlockHeader) error {

	ghostdagData, err := v.ghostdagDataStores[0].Get(v.databaseContext, stagingArea, blockHash, false)
	if err != nil {
		return err
	}
	if header.BlueScore() != ghostdagData.BlueScore() {
		return errors.Wrapf(ruleerrors.ErrUnexpectedBlueWork, "block blue work of %d is not the expected "+
			"value of %d", header.BlueWork(), ghostdagData.BlueScore())
	}
	return nil
}

// checkpointApplies (Rupix) dice si la regla del checkpoint cp aplica a un bloque
// con este blue score y DAA score: solo desde X + MergeDepth y antes de la caducidad.
func (v *blockValidator) checkpointApplies(cp dagconfig.Checkpoint, blueScore, daaScore uint64) bool {
	if v.checkpointsExpireDAAScore > 0 && daaScore > v.checkpointsExpireDAAScore {
		return false
	}
	return blueScore >= cp.BlueScore+v.mergeDepth
}

// checkCheckpoint (Rupix) exige que todo bloque con blue score >= X + MergeDepth
// tenga al bloque canonico H en su pasado (o sea H). Se revisa contra los padres,
// igual que checkPruningPointViolation: H debe ser ancestro de al menos uno.
// IsAncestorOf es inclusivo (H == padre cuenta).
// Si el nodo no tiene a H porque sincronizo desde un pruning point posterior a H,
// la regla no se puede evaluar aqui: esa defensa vive en la validacion de los
// pruning points recibidos (checkCheckpointsEnPruningPoints en blockprocessor,
// desde v0.6.2; ver CHECKPOINTS.md).
func (v *blockValidator) checkCheckpoint(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash, header externalapi.BlockHeader) error {
	if len(v.checkpoints) == 0 {
		return nil
	}
	// Se usa el blue score del header: checkHeaderBlueScore (mas abajo, siempre)
	// rechaza un header cuyo blue score no coincide con el calculado.
	blueScore := header.BlueScore()
	for _, cp := range v.checkpoints {
		if !v.checkpointApplies(cp, blueScore, header.DAAScore()) {
			continue
		}
		if blockHash.Equal(cp.Hash) {
			continue
		}
		hasH, err := v.reachabilityStore.HasReachabilityData(v.databaseContext, stagingArea, cp.Hash)
		if err != nil {
			return err
		}
		if !hasH {
			prunedBelow, err := v.checkpointBelowPruningPoint(stagingArea, cp)
			if err != nil {
				return err
			}
			if prunedBelow {
				continue
			}
			return errors.Wrapf(ruleerrors.ErrCheckpointMismatch,
				"bloque %s (blue score %d) no tiene en su pasado al checkpoint %s (blue score %d): el nodo no conoce ese bloque",
				blockHash, blueScore, cp.Hash, cp.BlueScore)
		}
		parents, err := v.dagTopologyManagers[0].Parents(stagingArea, blockHash)
		if err != nil {
			return err
		}
		isInPast, err := v.dagTopologyManagers[0].IsAncestorOfAny(stagingArea, cp.Hash, parents)
		if err != nil {
			return err
		}
		if !isInPast {
			return errors.Wrapf(ruleerrors.ErrCheckpointMismatch,
				"bloque %s (blue score %d) no tiene en su pasado al checkpoint %s (blue score %d)",
				blockHash, blueScore, cp.Hash, cp.BlueScore)
		}
	}
	return nil
}

// checkpointBelowPruningPoint (Rupix) dice si el pruning point actual del nodo ya
// esta por encima del checkpoint (el nodo sincronizo desde ahi y no guarda a H).
func (v *blockValidator) checkpointBelowPruningPoint(stagingArea *model.StagingArea, cp dagconfig.Checkpoint) (bool, error) {
	hasPruningPoint, err := v.pruningStore.HasPruningPoint(v.databaseContext, stagingArea)
	if err != nil || !hasPruningPoint {
		return false, err
	}
	pruningPoint, err := v.pruningStore.PruningPoint(v.databaseContext, stagingArea)
	if err != nil {
		return false, err
	}
	ppHeader, err := v.blockHeaderStore.BlockHeader(v.databaseContext, stagingArea, pruningPoint)
	if err != nil {
		return false, err
	}
	return ppHeader.BlueScore() > cp.BlueScore, nil
}
