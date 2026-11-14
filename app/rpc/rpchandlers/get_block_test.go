package rpchandlers_test

import (
	"strings"
	"testing"

	"github.com/rupixnet/rupixd/app/appmessage"
	"github.com/rupixnet/rupixd/app/rpc/rpccontext"
	"github.com/rupixnet/rupixd/app/rpc/rpchandlers"
	"github.com/rupixnet/rupixd/domain/consensus"
	"github.com/rupixnet/rupixd/domain/consensus/model"
	"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
	"github.com/rupixnet/rupixd/domain/consensus/utils/consensushashing"
	"github.com/rupixnet/rupixd/domain/consensus/utils/testutils"
	"github.com/rupixnet/rupixd/infrastructure/config"
)

// TestGetBlockEncabezadoSinEstado (Rupix, 14-nov-2026, heredado de kaspad): un encabezado
// conocido pero sin estado guardado (como los puntos de poda historicos durante la IBD) NO
// es un bloque invalido. Antes GetBlock contestaba "is invalid" porque el valor por defecto
// de BlockStatus es StatusInvalid; se vio en vivo confirmando el checkpoint #4 desde una
// laptop a medio sincronizar. Ahora dice que el nodo aun no lo valido.
func TestGetBlockEncabezadoSinEstado(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		factory := consensus.NewFactory()
		tc, teardown, err := factory.NewTestConsensus(consensusConfig, "TestGetBlockEncabezadoSinEstado")
		if err != nil {
			t.Fatalf("consenso de prueba: %+v", err)
		}
		defer teardown(false)

		// Un encabezado valido hijo del genesis, guardado SOLO en la tienda de encabezados.
		header, err := tc.BuildHeaderWithParents([]*externalapi.DomainHash{consensusConfig.GenesisHash})
		if err != nil {
			t.Fatalf("BuildHeaderWithParents: %+v", err)
		}
		hash := consensushashing.HeaderHash(header)
		stagingArea := model.NewStagingArea()
		tc.BlockHeaderStore().Stage(stagingArea, hash, header)
		tx, err := tc.DatabaseContext().Begin()
		if err != nil {
			t.Fatalf("Begin: %+v", err)
		}
		if err := stagingArea.Commit(tx); err != nil {
			t.Fatalf("Commit: %+v", err)
		}
		if err := tx.Commit(); err != nil {
			t.Fatalf("tx.Commit: %+v", err)
		}

		ctx := rpccontext.Context{
			Config: &config.Config{Flags: &config.Flags{NetworkFlags: config.NetworkFlags{ActiveNetParams: &consensusConfig.Params}}},
			Domain: fakeDomain{tc},
		}
		resp, err := rpchandlers.HandleGetBlock(&ctx, nil, &appmessage.GetBlockRequestMessage{Hash: hash.String()})
		if err != nil {
			t.Fatalf("HandleGetBlock: %+v", err)
		}
		msg := resp.(*appmessage.GetBlockResponseMessage)
		if msg.Error == nil {
			t.Fatalf("sin estado debe contestar con un error explicativo, no con el bloque")
		}
		if strings.Contains(msg.Error.Message, "is invalid") || !strings.Contains(msg.Error.Message, "not validated yet") {
			t.Fatalf("mensaje engañoso: %q", msg.Error.Message)
		}
	})
}
