package blockprocessor

import (
	"github.com/pkg/errors"
	"github.com/rupixnet/rupixd/domain/consensus/model"
	"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
	"github.com/rupixnet/rupixd/domain/consensus/ruleerrors"
)

// checkCheckpointsEnPruningPoints (Rupix, hueco #2 de ESPECIFICACION.md) protege
// al nodo que sincroniza desde cero.
//
// La regla por encabezado (checkCheckpoint en blockvalidator) exige que todo
// bloque con blue score >= X + MergeDepth tenga al checkpoint H en su pasado.
// Pero un nodo nuevo no descarga la historia completa: recibe de un peer una
// lista de pruning points (MsgPruningPoints) con su prueba y arranca desde el
// ultimo. Si ese peer es hostil y sirve una historia alterna sin H, el nodo
// nuevo nunca tendria a H y la regla por encabezado no se podria evaluar
// (checkpointBelowPruningPoint la salta a proposito). Esta funcion cierra ese
// hueco: antes de importar el pruning point, cada checkpoint activo que quede
// por debajo de el tiene que estar en su pasado.
//
// Como se decide, por orden:
//  1. H es uno de los pruning points de la lista recibida (indice 0..actual).
//     Es el caso normal: por politica (CHECKPOINTS.md) cada checkpoint publicado
//     es un punto de poda de la red, asi que esta en la lista de todo peer honesto.
//  2. Si no esta en la lista pero el nodo conoce a H (tiene sus datos de
//     alcanzabilidad, por ejemplo porque H llego en la prueba de poda), basta
//     con que H sea ancestro del pruning point.
//  3. Si no, ErrCheckpointMismatch: la lista no es la de la red publicada y el
//     pruning point no se importa.
//
// Un checkpoint por encima del pruning point (H todavia no llego) no se revisa
// aqui: lo revisa la regla por encabezado cuando lleguen los bloques. Y pasada
// la caducidad (CheckpointsExpireDAAScore) no se revisa nada, igual que alla.
func (bp *blockProcessor) checkCheckpointsEnPruningPoints(stagingArea *model.StagingArea,
	newPruningPointHash *externalapi.DomainHash) error {

	if len(bp.checkpoints) == 0 {
		return nil
	}

	ppHeader, err := bp.blockHeaderStore.BlockHeader(bp.databaseContext, stagingArea, newPruningPointHash)
	if err != nil {
		return err
	}
	if bp.checkpointsExpireDAAScore > 0 && ppHeader.DAAScore() > bp.checkpointsExpireDAAScore {
		return nil
	}

	currentIndex, err := bp.pruningStore.CurrentPruningPointIndex(bp.databaseContext, stagingArea)
	if err != nil {
		return err
	}

	for _, cp := range bp.checkpoints {
		if ppHeader.BlueScore() < cp.BlueScore {
			continue
		}

		// 1) H esta en la lista de pruning points.
		enLaLista := false
		for i := currentIndex; ; i-- {
			pruningPoint, err := bp.pruningStore.PruningPointByIndex(bp.databaseContext, stagingArea, i)
			if err != nil {
				return err
			}
			if pruningPoint.Equal(cp.Hash) {
				enLaLista = true
				break
			}
			if i == 0 {
				break
			}
		}
		if enLaLista {
			continue
		}

		// 2) El nodo conoce a H: tiene que ser ancestro del pruning point.
		conoceH, err := bp.reachabilityDataStore.HasReachabilityData(bp.databaseContext, stagingArea, cp.Hash)
		if err != nil {
			return err
		}
		if conoceH {
			esAncestro, err := bp.dagTopologyManager.IsAncestorOf(stagingArea, cp.Hash, newPruningPointHash)
			if err != nil {
				return err
			}
			if esAncestro {
				continue
			}
		}

		// 3) Historia alterna.
		return errors.Wrapf(ruleerrors.ErrCheckpointMismatch,
			"el pruning point %s (blue score %d) no tiene en su pasado al checkpoint %s (blue score %d): "+
				"la lista de pruning points de este peer no es la de la red publicada",
			newPruningPointHash, ppHeader.BlueScore(), cp.Hash, cp.BlueScore)
	}
	return nil
}
