package templatemanager

import (
	"github.com/rupixnet/rupixd/app/appmessage"
	"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
	"github.com/rupixnet/rupixd/domain/consensus/utils/pow"
	"sync"
)

var currentTemplate *externalapi.DomainBlock
var currentState *pow.State
var isSynced bool
var lock = &sync.Mutex{}

// Get returns the template to work on
func Get() (*externalapi.DomainBlock, *pow.State, bool) {
	lock.Lock()
	defer lock.Unlock()
	// Shallow copy the block so when the user replaces the header it won't affect the template here.
	if currentTemplate == nil {
		return nil, nil, false
	}
	block := *currentTemplate
	state := *currentState
	return &block, &state, isSynced
}

// Invalidate clears the current template forcing the miner to wait for a new one
func Invalidate() {
	lock.Lock()
	defer lock.Unlock()
	currentTemplate = nil
	currentState = nil
}

// Set sets the current template to work on
func Set(template *appmessage.GetBlockTemplateResponseMessage) error {
	block, err := appmessage.RPCBlockToDomainBlock(template.Block)
	if err != nil {
		return err
	}
	lock.Lock()
	defer lock.Unlock()
	currentTemplate = block
	currentState = pow.NewState(block.Header.ToMutable())
	isSynced = template.IsSynced
	return nil
}

