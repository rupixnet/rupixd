package appmessage

// GetLevelByAddressRequestMessage consulta el nivel L1-L5 de una direccion
type GetLevelByAddressRequestMessage struct {
	baseMessage
	Address string `json:"address"`
}

// Command devuelve el comando RPC
func (msg *GetLevelByAddressRequestMessage) Command() MessageCommand {
	return CmdGetLevelByAddressRequestMessage
}

// NewGetLevelByAddressRequest crea un nuevo request
func NewGetLevelByAddressRequest(address string) *GetLevelByAddressRequestMessage {
	return &GetLevelByAddressRequestMessage{
		Address: address,
	}
}

// GetLevelByAddressResponseMessage respuesta con el nivel de la direccion
type GetLevelByAddressResponseMessage struct {
	baseMessage
	Address   string `json:"address"`
	Level     uint32 `json:"level"`
	LevelName string `json:"levelName"`
	Error     *RPCError `json:"error,omitempty"`
}

// Command devuelve el comando RPC
func (msg *GetLevelByAddressResponseMessage) Command() MessageCommand {
	return CmdGetLevelByAddressResponseMessage
}