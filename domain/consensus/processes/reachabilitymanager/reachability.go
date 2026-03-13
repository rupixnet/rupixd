package reachabilitymanager

import (
	"github.com/rupixnet/rupixd/domain/consensus/model"
	"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
	"github.com/rupixnet/rupixd/domain/consensus/database"
)

// IsDAGAncestorOf returns true if blockHashA is an ancestor of
// blockHashB in the DAG.
//
// Note: this method will return true if blockHashA == blockHashB
// The complexity of this method is O(log(|this.futureCoveringTreeNodeSet|))
func (rt *reachabilityManager) IsDAGAncestorOf(stagingArea *model.StagingArea, blockHashA *externalapi.DomainHash, blockHashB *externalapi.DomainHash) (bool, error) {
	// Check if this node is a reachability tree ancestor of the
	// other node
	isReachabilityTreeAncestor, err := rt.IsReachabilityTreeAncestorOf(stagingArea, blockHashA, blockHashB)
	if err != nil {
		if database.IsNotFoundError(err) {
			return false, nil
		}
		return false, err
	}
	if isReachabilityTreeAncestor {
		return true, nil
	}

	// Otherwise, use previously registered future blocks to complete the
	// reachability test
	return rt.futureCoveringSetHasAncestorOf(stagingArea, blockHashA, blockHashB)
}

func (rt *reachabilityManager) UpdateReindexRoot(stagingArea *model.StagingArea, selectedTip *externalapi.DomainHash) error {
    // Si el selectedTip es el VirtualGenesisBlockHash, no hay nada que hacer
    if selectedTip.Equal(model.VirtualGenesisBlockHash) {
        return nil
    }
    return rt.updateReindexRoot(stagingArea, selectedTip)
}

