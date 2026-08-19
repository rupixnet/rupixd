package protowire

import (
	"github.com/pkg/errors"
	"github.com/rupixnet/rupixd/app/appmessage"
)

func (x *RPCError) toAppMessage() (*appmessage.RPCError, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "RPCError is nil")
	}
	return &appmessage.RPCError{Message: x.Message}, nil
}
