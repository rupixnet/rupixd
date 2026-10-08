package rpcclient

import "github.com/rupixnet/rupixd/app/appmessage"

// GetGemsInfo (Rupix) sends an RPC request respective to the function's name and returns the RPC server's response
func (c *RPCClient) GetGemsInfo() (*appmessage.GetGemsInfoResponseMessage, error) {
	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewGetGemsInfoRequestMessage())
	if err != nil {
		return nil, err
	}
	response, err := c.route(appmessage.CmdGetGemsInfoResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return nil, err
	}
	resp := response.(*appmessage.GetGemsInfoResponseMessage)
	if resp.Error != nil {
		return nil, c.convertRPCError(resp.Error)
	}
	return resp, nil
}
