package protowire

import (
	"github.com/rupixnet/rupixd/app/appmessage"
	"github.com/pkg/errors"
)

func (x *RupixdMessage_Verack) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "RupixdMessage_Verack is nil")
	}
	return &appmessage.MsgVerAck{}, nil
}

func (x *RupixdMessage_Verack) fromAppMessage(_ *appmessage.MsgVerAck) error {
	return nil
}

