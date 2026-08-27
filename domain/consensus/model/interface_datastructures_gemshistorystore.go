package model

import "github.com/rupixnet/rupixd/domain/consensus/model/externalapi"

// GemsHistoryStore guarda, por bloque, cuantas gemas de nivel Diamante/Platino/
// Rodio han NACIDO historicamente hasta ese bloque. Se calcula junto al multiset
// y al kingscount a partir del mismo acceptanceData, y hereda su ciclo de vida:
// reorgs y pruning incluidos. Es la muralla de los topes historicos.
type GemsHistoryStore interface {
Store
Stage(stagingArea *StagingArea, blockHash *externalapi.DomainHash, history *externalapi.GemsHistory)
IsStaged(stagingArea *StagingArea) bool
Get(dbContext DBReader, stagingArea *StagingArea, blockHash *externalapi.DomainHash) (*externalapi.GemsHistory, error)
Delete(stagingArea *StagingArea, blockHash *externalapi.DomainHash)
}
