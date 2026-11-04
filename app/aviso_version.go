package app

import (
	"sync/atomic"
	"time"

	"github.com/rupixnet/rupixd/version"
)

// Rupix (4-nov-2026): aviso de version nueva por el saludo de los pares (ver version/masnueva.go).
// Cada 10 minutos mira que version anuncian los pares conectados; si la mayoria trae una
// mayor que la propia, lo escribe en el log, como mucho una vez cada 24 h. No manda nada,
// no consulta a nadie, y no puede decir otra cosa que un numero de version.
const (
	avisoVersionCada     = 10 * time.Minute
	avisoVersionSilencio = 24 * time.Hour
)

func (a *ComponentManager) vigilarVersionNueva() {
	var ultimoAviso time.Time
	for atomic.LoadInt32(&a.shutdown) == 0 {
		time.Sleep(avisoVersionCada)
		peers := a.protocolManager.Peers()
		uas := make([]string, 0, len(peers))
		for _, p := range peers {
			uas = append(uas, p.UserAgent())
		}
		nueva, hay := version.MasNueva(version.Version(), uas)
		if !hay || time.Since(ultimoAviso) < avisoVersionSilencio {
			continue
		}
		ultimoAviso = time.Now()
		log.Warnf("Hay una version mas nueva de Rupix (%s): la mayoria de los nodos conectados ya la usa; esta corre %s. Descarga y suma en https://rupix.network/guia.html / A newer Rupix version (%s) is out: most connected nodes already run it; this one runs %s. Download at https://rupix.network/guide.html",
			nueva, version.Version(), nueva, version.Version())
	}
}
