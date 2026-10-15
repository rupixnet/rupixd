package main

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/rupixnet/rupixd/cmd/rupixwallet/daemon/pb"
	"google.golang.org/grpc"
)

// clienteGemsFalso responde a Gems con una secuencia de conteos de Diamante.
type clienteGemsFalso struct {
	pb.KaspawalletdClient
	diamantes []uint64
	llamadas  int
}

func (c *clienteGemsFalso) Gems(_ context.Context, _ *pb.GemsRequest, _ ...grpc.CallOption) (*pb.GemsResponse, error) {
	i := c.llamadas
	if i >= len(c.diamantes) {
		i = len(c.diamantes) - 1
	}
	c.llamadas++
	return &pb.GemsResponse{Diamante: c.diamantes[i]}, nil
}

// TestEsperarGema (Rupix, 14-oct-2026, forge --count): entre forja y forja la wallet
// espera a VER la gema nueva; si aparece, sigue con el conteo nuevo; si no aparece en
// el plazo, se detiene con un error que lo dice.
func TestEsperarGema(t *testing.T) {
	viejo := esperaGemaMax
	defer func() { esperaGemaMax = viejo }()
	esperaGemaMax = 10 * time.Second

	c := &clienteGemsFalso{diamantes: []uint64{5, 5, 6}}
	got, err := esperarGema(c, 1, 5)
	if err != nil {
		t.Fatalf("esperarGema: %v", err)
	}
	if got != 6 {
		t.Fatalf("conteo %d, esperaba 6", got)
	}
	if c.llamadas != 3 {
		t.Fatalf("debio preguntar 3 veces, pregunto %d", c.llamadas)
	}

	esperaGemaMax = 0
	c = &clienteGemsFalso{diamantes: []uint64{5}}
	if _, err := esperarGema(c, 1, 5); err == nil || !(strings.Contains(err.Error(), "Diamante") || strings.Contains(err.Error(), "Diamond")) {
		t.Fatalf("sin gema nueva debe fallar nombrando el nivel; err=%v", err)
	}
}
