package app

import (
	"sync/atomic"
	"time"

	"github.com/rupixnet/rupixd/domain"
)

// Rupix (10-nov-2026): el nodo vigila la caducidad de su propia lista de checkpoints
// (dagconfig.EstadoCheckpoints). Primera mirada a los 2 minutos de arrancar y luego cada
// 6 horas. Con menos de MargenAvisoCheckpoints avisa; caducada, lo dice sin rodeos: los
// nodos que sincronicen desde cero no tienen esa proteccion hasta la version siguiente.
const (
	avisoCheckpointsPrimero = 2 * time.Minute
	avisoCheckpointsCada    = 6 * time.Hour
)

func (a *ComponentManager) vigilarCheckpoints(dom domain.Domain) {
	espera := avisoCheckpointsPrimero
	for atomic.LoadInt32(&a.shutdown) == 0 {
		time.Sleep(espera)
		espera = avisoCheckpointsCada
		daa, err := dom.Consensus().GetVirtualDAAScore()
		if err != nil {
			continue
		}
		params := a.cfg.NetParams()
		caducada, pronto, margen := params.EstadoCheckpoints(daa)
		switch {
		case caducada:
			log.Warnf("La lista de checkpoints de esta version caduco en el DAA %d (hoy %d): este nodo sigue bien, pero un nodo que sincronice desde cero no tiene esa proteccion hasta la proxima version. Mira https://rupix.network / This version's checkpoint list expired at DAA %d (now %d): this node is fine, but a node syncing from scratch lacks that protection until the next version. See https://rupix.network",
				params.CheckpointsExpireDAAScore, daa, params.CheckpointsExpireDAAScore, daa)
		case pronto:
			log.Warnf("La lista de checkpoints caduca en el DAA %d; faltan %d (~%.1f dias). Toca publicar el siguiente checkpoint en una version nueva / Checkpoint list expires at DAA %d; %d left (~%.1f days). The next checkpoint must ship in a new version",
				params.CheckpointsExpireDAAScore, margen, float64(margen)/86400, params.CheckpointsExpireDAAScore, margen, float64(margen)/86400)
		}
	}
}
