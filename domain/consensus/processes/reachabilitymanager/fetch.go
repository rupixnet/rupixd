package reachabilitymanager

import (
	"github.com/rupixnet/rupixd/domain/consensus/database"
	"github.com/rupixnet/rupixd/domain/consensus/model"
	"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
	"github.com/rupixnet/rupixd/domain/consensus/utils/reachabilitydata"
	"github.com/pkg/errors"
)

func (rt *reachabilityManager) reachabilityDataForInsertion(stagingArea *model.StagingArea,
	blockHash *externalapi.DomainHash) (model.MutableReachabilityData, error) {
	data, err := rt.reachabilityDataStore.ReachabilityData(rt.databaseContext, stagingArea, blockHash)
	if err == nil {
		return data.CloneMutable(), nil
	}

	if errors.Is(err, database.ErrNotFound) {
		return reachabilitydata.EmptyReachabilityData(), nil
	}
	return nil, err
}

// FIX-011 (2026-05-31): revertido parche RUPIX-017 al comportamiento Kaspa upstream.
// El parche anterior retornaba FutureCoveringTreeNodeSet{} vacio cuando no encontraba
// reachability data, lo cual hacia que IsAncestorOf retornara FALSE incorrectamente.
// Eso causaba que BuildParents agregara ancestros como direct parents (bug del 67).
func (rt *reachabilityManager) futureCoveringSet(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (model.FutureCoveringTreeNodeSet, error) {
	data, err := rt.reachabilityDataStore.ReachabilityData(rt.databaseContext, stagingArea, blockHash)
	if err != nil {
		return nil, err
	}

	return data.FutureCoveringSet(), nil
}

// FIX-011: revertido parche RUPIX-017 al comportamiento Kaspa upstream.
// Retornar {Start:0, End:0} cuando NotFound hacia que las comparaciones de intervalos
// retornaran FALSE incorrectamente, rompiendo IsAncestorOf.
func (rt *reachabilityManager) interval(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (*model.ReachabilityInterval, error) {
	data, err := rt.reachabilityDataStore.ReachabilityData(rt.databaseContext, stagingArea, blockHash)
	if err != nil {
		return nil, err
	}

	return data.Interval(), nil
}

// FIX-011: revertido parche RUPIX-017 al comportamiento Kaspa upstream.
func (rt *reachabilityManager) children(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (
	[]*externalapi.DomainHash, error) {

	data, err := rt.reachabilityDataStore.ReachabilityData(rt.databaseContext, stagingArea, blockHash)
	if err != nil {
		return nil, err
	}

	return data.Children(), nil
}

// FIX-011: revertido parche RUPIX-017 al comportamiento Kaspa upstream.
func (rt *reachabilityManager) parent(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (
	*externalapi.DomainHash, error) {

	data, err := rt.reachabilityDataStore.ReachabilityData(rt.databaseContext, stagingArea, blockHash)
	if err != nil {
		return nil, err
	}

	return data.Parent(), nil
}

// FIX-011: revertido parche RUPIX-017 al comportamiento Kaspa upstream.
func (rt *reachabilityManager) reindexRoot(stagingArea *model.StagingArea) (*externalapi.DomainHash, error) {
	return rt.reachabilityDataStore.ReachabilityReindexRoot(rt.databaseContext, stagingArea)
}
