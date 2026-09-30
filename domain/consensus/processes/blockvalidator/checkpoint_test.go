package blockvalidator

import (
	"testing"

	"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
	"github.com/rupixnet/rupixd/domain/consensus/utils/blockheader"
	"github.com/rupixnet/rupixd/domain/dagconfig"
	"math/big"
)

// Pruebas unitarias de la regla de checkpoints para DAG (v0.6.1). Las pruebas con
// un DAG real (hermanos, atacante, bloque tardio) estan en
// domain/consensus/checkpoint_dag_test.go.

func hashDe(b byte) *externalapi.DomainHash {
	return externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{b})
}

// TestSinCheckpointsPasaTodo: lista vacia (como todas las redes hoy) -> no afecta nada
// y no toca ningun store.
func TestSinCheckpointsPasaTodo(t *testing.T) {
	v := &blockValidator{}
	header := blockheader.NewImmutableBlockHeader(0, nil, hashDe(0), hashDe(0), hashDe(0), hashDe(0),
		0, 0, 0, 100, 100, big.NewInt(0), hashDe(0))
	if err := v.checkCheckpoint(nil, hashDe(1), header); err != nil {
		t.Fatalf("sin checkpoints debe pasar todo, dio: %v", err)
	}
}

// TestCheckpointAplicaDesdeElMargen: la regla aplica desde X + MergeDepth, no antes.
// Un hermano de H (mismo blue score X) nunca queda sujeto a la regla.
func TestCheckpointAplicaDesdeElMargen(t *testing.T) {
	v := &blockValidator{mergeDepth: 10}
	cp := dagconfig.Checkpoint{BlueScore: 100, Hash: hashDe(7)}
	if v.checkpointApplies(cp, 100, 100) {
		t.Fatalf("un bloque con el mismo blue score que H no debe quedar sujeto a la regla")
	}
	if v.checkpointApplies(cp, 109, 109) {
		t.Fatalf("antes de X + MergeDepth la regla no aplica")
	}
	if !v.checkpointApplies(cp, 110, 110) {
		t.Fatalf("en X + MergeDepth la regla aplica")
	}
}

// TestCheckpointCaducadoSeIgnora: pasado checkpointsExpireDAAScore la regla no aplica.
func TestCheckpointCaducadoSeIgnora(t *testing.T) {
	v := &blockValidator{mergeDepth: 10, checkpointsExpireDAAScore: 150}
	cp := dagconfig.Checkpoint{BlueScore: 100, Hash: hashDe(7)}
	if v.checkpointApplies(cp, 200, 200) {
		t.Fatalf("un checkpoint caducado debe ignorarse")
	}
	if !v.checkpointApplies(cp, 140, 140) {
		t.Fatalf("antes de caducar, la regla aplica")
	}
}
