package blockvalidator

import (
"testing"

"github.com/pkg/errors"
"github.com/rupixnet/rupixd/domain/consensus/model/externalapi"
"github.com/rupixnet/rupixd/domain/consensus/ruleerrors"
"github.com/rupixnet/rupixd/domain/consensus/utils/blockheader"
"github.com/rupixnet/rupixd/domain/dagconfig"
"math/big"
)

func hashDe(b byte) *externalapi.DomainHash {
return externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{b})
}

// headerConDAA construye un header minimo con el DAA score dado (solo eso importa aqui).
func headerConDAA(daa uint64) externalapi.BlockHeader {
return blockheader.NewImmutableBlockHeader(0, nil, hashDe(0), hashDe(0), hashDe(0), hashDe(0),
0, 0, 0, daa, 0, big.NewInt(0), hashDe(0))
}

// TestCheckpointCorrectoPasa: bloque en el DAA score del checkpoint con el hash canonico -> pasa.
func TestCheckpointCorrectoPasa(t *testing.T) {
v := &blockValidator{checkpoints: []dagconfig.Checkpoint{{DAAScore: 100, Hash: hashDe(7)}}}
if err := v.checkCheckpoint(hashDe(7), headerConDAA(100)); err != nil {
t.Fatalf("checkpoint correcto debe pasar, dio: %v", err)
}
}

// TestCheckpointHashDistintoRechaza: mismo DAA score, otro hash -> ErrCheckpointMismatch.
// Este es el caso del atacante del 51%: reescribe la historia y su bloque no es el canonico.
func TestCheckpointHashDistintoRechaza(t *testing.T) {
v := &blockValidator{checkpoints: []dagconfig.Checkpoint{{DAAScore: 100, Hash: hashDe(7)}}}
err := v.checkCheckpoint(hashDe(8), headerConDAA(100))
if !errors.Is(err, ruleerrors.ErrCheckpointMismatch) {
t.Fatalf("hash distinto en un checkpoint debe dar ErrCheckpointMismatch, dio: %v", err)
}
}

// TestSinCheckpointEnEseScorePasa: DAA score sin checkpoint -> pasa cualquier hash.
func TestSinCheckpointEnEseScorePasa(t *testing.T) {
v := &blockValidator{checkpoints: []dagconfig.Checkpoint{{DAAScore: 100, Hash: hashDe(7)}}}
if err := v.checkCheckpoint(hashDe(9), headerConDAA(101)); err != nil {
t.Fatalf("sin checkpoint en ese score debe pasar, dio: %v", err)
}
}

// TestCheckpointCaducadoSeIgnora: pasado checkpointsExpireDAAScore, un hash distinto
// en el score del checkpoint YA NO se rechaza (la caducidad publicada se respeta).
func TestCheckpointCaducadoSeIgnora(t *testing.T) {
v := &blockValidator{
checkpoints:               []dagconfig.Checkpoint{{DAAScore: 100, Hash: hashDe(7)}},
checkpointsExpireDAAScore: 50, // caduco antes del score 100
}
if err := v.checkCheckpoint(hashDe(8), headerConDAA(100)); err != nil {
t.Fatalf("checkpoint caducado debe ignorarse, dio: %v", err)
}
}

// TestSinCheckpointsPasaTodo: lista vacia (como la testnet hoy) -> no afecta nada.
func TestSinCheckpointsPasaTodo(t *testing.T) {
v := &blockValidator{}
if err := v.checkCheckpoint(hashDe(1), headerConDAA(100)); err != nil {
t.Fatalf("sin checkpoints debe pasar todo, dio: %v", err)
}
}
