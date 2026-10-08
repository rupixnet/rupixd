package grpcserver

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/rupixnet/rupixd/app/appmessage"
	"github.com/rupixnet/rupixd/infrastructure/logger"
	"io"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/pkg/errors"
	routerpkg "github.com/rupixnet/rupixd/infrastructure/network/netadapter/router"

	"github.com/rupixnet/rupixd/infrastructure/network/netadapter/server/grpcserver/protowire"
)

func (c *gRPCConnection) connectionLoops() error {
	errChan := make(chan error, 1) // buffered channel because one of the loops might try write after disconnect

	spawn("gRPCConnection.receiveLoop", func() { errChan <- c.receiveLoop() })
	spawn("gRPCConnection.sendLoop", func() { errChan <- c.sendLoop() })

	err := <-errChan

	c.Disconnect()

	return err
}

var blockDelayOnce sync.Once
var blockDelay = 0

func (c *gRPCConnection) sendLoop() error {
	outgoingRoute := c.router.OutgoingRoute()
	for c.IsConnected() {
		message, err := outgoingRoute.Dequeue()
		if err != nil {
			if errors.Is(err, routerpkg.ErrRouteClosed) {
				return nil
			}
			return err
		}

		blockDelayOnce.Do(func() {
			experimentalDelayEnv := os.Getenv("KASPA_EXPERIMENTAL_DELAY")
			if experimentalDelayEnv != "" {
				blockDelay, err = strconv.Atoi(experimentalDelayEnv)
				if err != nil {
					panic(err)
				}
			}
		})

		if blockDelay != 0 && message.Command() == appmessage.CmdBlock {
			time.Sleep(time.Duration(blockDelay) * time.Second)
		}

		log.Debugf("outgoing '%s' message to %s", message.Command(), c)
		log.Tracef("outgoing '%s' message to %s: %s", message.Command(), c, logger.NewLogClosure(func() string {
			return spew.Sdump(message)
		}))

		messageProto, err := protowire.FromAppMessage(message)
		if err != nil {
			return err
		}

		err = c.send(messageProto)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *gRPCConnection) receiveLoop() error {
	messageNumber := uint64(0)
	for c.IsConnected() {
		protoMessage, err := c.receive()
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return err
		}
		// Rupix: lista blanca por tipo de conexion, antes de convertir nada.
		if err := cargaPermitida(c.server.name, protoMessage); err != nil {
			log.Warnf("%s desde %s: %v; se cierra la conexion", c.server.name, c, err)
			if c.onInvalidMessageHandler != nil {
				c.onInvalidMessageHandler(err)
			}
			return err
		}
		message, err := toAppMessageSinCaer(protoMessage)
		if err != nil {
			if c.onInvalidMessageHandler != nil {
				c.onInvalidMessageHandler(err)
			}
			return err
		}

		messageNumber++
		message.SetMessageNumber(messageNumber)
		message.SetReceivedAt(time.Now())

		log.Debugf("incoming '%s' message from %s (message number %d)", message.Command(), c,
			message.MessageNumber())

		log.Tracef("incoming '%s' message from %s  (message number %d): %s", message.Command(),
			c, message.MessageNumber(), logger.NewLogClosure(func() string {
				return spew.Sdump(message)
			}))

		err = c.router.EnqueueIncomingMessage(message)
		if err != nil {
			if errors.Is(err, routerpkg.ErrRouteClosed) {
				return nil
			}

			// ErrRouteCapacityReached isn't an invalid message error, so
			// we return it in order to log it later on.
			if errors.Is(err, routerpkg.ErrRouteCapacityReached) {
				return err
			}
			if c.onInvalidMessageHandler != nil {
				c.onInvalidMessageHandler(err)
			}
			return err
		}
	}
	return nil
}

// toAppMessageSinCaer (Rupix, hueco #6, 3-oct-2026) convierte un mensaje recibido de
// la red y, si el conversor entra en panico por un mensaje malformado, lo vuelve un
// error en vez de dejar que mate al proceso. El fuzzing encontro un mensaje de 9 bytes
// que hacia exactamente eso (RpcFeeEstimate sin PriorityBucket; arreglado en la raiz).
// Esta red de seguridad es para los que aun no se han encontrado: un mensaje malo
// cierra ESA conexion y queda en el log; el nodo sigue. Vive aqui, en el punto de
// recepcion, y no dentro de ToAppMessage, para que el fuzzing siga viendo los panicos
// de los conversores y se arreglen en la raiz.
// convertidorDeRed es lo unico que toAppMessageSinCaer necesita de un mensaje; asi el
// test puede meterle un conversor que entra en panico a proposito.
type convertidorDeRed interface {
	ToAppMessage() (appmessage.Message, error)
}

func toAppMessageSinCaer(protoMessage convertidorDeRed) (message appmessage.Message, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Warnf("mensaje de red malformado: el conversor entro en panico (%v); se cierra la conexion", r)
			message, err = nil, errors.Errorf("malformed network message: %v", r)
		}
	}()
	return protoMessage.ToAppMessage()
}

// cargaPermitida (Rupix) rechaza un payload que no pertenece al protocolo de la conexion:
// en una conexion P2P solo entran payloads P2P; en una RPC, solo RPC. Reduce la superficie
// que un peer puede tocar de todos los conversores a los de su protocolo, y vuelve
// irrelevante para un peer cualquier bug que quede en los conversores RPC. Un nombre de
// servidor desconocido no restringe nada (no hay un tercer tipo hoy; si lo hubiera, se
// agrega aqui a proposito).
func cargaPermitida(nombreServidor string, protoMessage *protowire.KaspadMessage) error {
	switch nombreServidor {
	case "P2P":
		if !protoMessage.EsCargaP2P() {
			return errors.Errorf("payload %d no pertenece al protocolo P2P", protoMessage.NumeroDeCarga())
		}
	case "RPC":
		if !protoMessage.EsCargaRPC() {
			return errors.Errorf("payload %d no pertenece al RPC", protoMessage.NumeroDeCarga())
		}
	}
	return nil
}
