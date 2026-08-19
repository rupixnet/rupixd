package model

import "github.com/rupixnet/rupixd/domain/consensus/model/externalapi"

// KingsCountStore guarda, por bloque, cuantos Kings (nivel 4) existen en el
// UTXO set de ese bloque. Se calcula junto al multiset a partir del mismo
// acceptanceData, y hereda su ciclo de vida: reorgs y pruning incluidos.
// Es la muralla del tope MaxKings = 2,100.
type KingsCountStore interface {
Store
Stage(stagingArea *StagingArea, blockHash *externalapi.DomainHash, count uint64)
IsStaged(stagingArea *StagingArea) bool
Get(dbContext DBReader, stagingArea *StagingArea, blockHash *externalapi.DomainHash) (uint64, error)
Delete(stagingArea *StagingArea, blockHash *externalapi.DomainHash)
}
