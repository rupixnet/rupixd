package protowire

import (
	"github.com/rupixnet/rupixd/app/appmessage"
	"github.com/pkg/errors"
)

func (x *RupixdMessage_IbdBlock) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "RupixdMessage_IbdBlock is nil")
	}
	msgBlock, err := x.IbdBlock.toAppMessage()
	if err != nil {
		return nil, err
	}
	return &appmessage.MsgIBDBlock{MsgBlock: msgBlock}, nil
}

func (x *RupixdMessage_IbdBlock) fromAppMessage(msgIBDBlock *appmessage.MsgIBDBlock) error {
	x.IbdBlock = new(BlockMessage)
	return x.IbdBlock.fromAppMessage(msgIBDBlock.MsgBlock)
}

