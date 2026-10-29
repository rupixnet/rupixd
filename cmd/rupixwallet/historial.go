// historial (Rupix, 28-oct-2026): lo que tiene tu wallet y cuando llego, en lenguaje llano.
//
// Honestidad primero: el nodo no guarda un indice de transacciones y la cadena se poda, asi
// que lo unico que se puede listar con certeza es lo que la wallet TIENE hoy (sus UTXOs) y en
// que bloque (DAA) llego cada cosa. Lo que ya se gasto no aparece. La fecha se estima a un
// bloque por segundo desde el DAA actual; por eso se imprime con "≈".
//
// Fuente: el nodo (GetUTXOsByAddresses) con las direcciones que el daemon reporta con saldo
// (GetBalance), asi tambien entran las direcciones de cambio.
package main

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/rupixnet/rupixd/app/appmessage"
	"github.com/rupixnet/rupixd/cmd/rupixwallet/daemon/client"
	"github.com/rupixnet/rupixd/cmd/rupixwallet/daemon/pb"
	"github.com/rupixnet/rupixd/infrastructure/network/rpcclient"
)

// movimiento es una entrada del historial: un UTXO que la wallet tiene hoy.
type movimiento struct {
	DAA       uint64
	Cuando    time.Time // estimada
	Coinbase  bool
	Nivel     uint32 // 0 = Gold, 1..4 = gema
	Monto     uint64 // rupias
	Direccion string
	TxID      string
}

// fechaDesdeDAA estima cuando se mino el bloque con ese DAA: un bloque por segundo hacia
// atras desde el DAA virtual actual. Es una estimacion y se muestra como tal.
func fechaDesdeDAA(daa, daaVirtual uint64, ahora time.Time) time.Time {
	if daa >= daaVirtual {
		return ahora
	}
	return ahora.Add(-time.Duration(daaVirtual-daa) * time.Second)
}

// movimientosDesdeUTXOs convierte la respuesta del nodo en movimientos ordenados del mas
// reciente al mas viejo.
func movimientosDesdeUTXOs(entradas []*appmessage.UTXOsByAddressesEntry, daaVirtual uint64, ahora time.Time) []movimiento {
	ms := make([]movimiento, 0, len(entradas))
	for _, e := range entradas {
		if e == nil || e.UTXOEntry == nil {
			continue
		}
		m := movimiento{DAA: e.UTXOEntry.BlockDAAScore, Coinbase: e.UTXOEntry.IsCoinbase, Monto: e.UTXOEntry.Amount, Direccion: e.Address}
		if e.UTXOEntry.ScriptPublicKey != nil {
			m.Nivel = uint32(e.UTXOEntry.ScriptPublicKey.Version)
		}
		if e.Outpoint != nil {
			m.TxID = e.Outpoint.TransactionID
		}
		m.Cuando = fechaDesdeDAA(m.DAA, daaVirtual, ahora)
		ms = append(ms, m)
	}
	sort.SliceStable(ms, func(i, j int) bool { return ms[i].DAA > ms[j].DAA })
	return ms
}

// diaMinado agrupa el minado de un dia: muchos bloques chicos que, uno por uno, no dicen nada.
type diaMinado struct {
	Dia     string // YYYY-MM-DD
	Bloques int
	Monto   uint64
}

// agruparMinado separa el minado (coinbase, Gold) agrupado por dia del resto, que se lista
// uno por uno. Las gemas nunca se agrupan aunque vengan de coinbase (no vienen).
func agruparMinado(ms []movimiento) (dias []diaMinado, resto []movimiento) {
	porDia := map[string]*diaMinado{}
	for _, m := range ms {
		if m.Coinbase && m.Nivel == 0 {
			d := m.Cuando.UTC().Format("2006-01-02")
			if porDia[d] == nil {
				porDia[d] = &diaMinado{Dia: d}
			}
			porDia[d].Bloques++
			porDia[d].Monto += m.Monto
			continue
		}
		resto = append(resto, m)
	}
	for _, d := range porDia {
		dias = append(dias, *d)
	}
	sort.Slice(dias, func(i, j int) bool { return dias[i].Dia > dias[j].Dia })
	return dias, resto
}

// filtrar aplica --desde, --hasta (fechas YYYY-MM-DD, inclusivas, UTC) y --buscar (texto en
// direccion o id de transaccion). Fechas mal escritas devuelven error en vez de un historial
// vacio que parezca verdadero.
func filtrar(ms []movimiento, desde, hasta, buscar string) ([]movimiento, error) {
	var d0, d1 time.Time
	var err error
	if desde != "" {
		if d0, err = time.Parse("2006-01-02", desde); err != nil {
			return nil, fmt.Errorf("--desde: %q no es una fecha AAAA-MM-DD", desde)
		}
	}
	if hasta != "" {
		if d1, err = time.Parse("2006-01-02", hasta); err != nil {
			return nil, fmt.Errorf("--hasta: %q no es una fecha AAAA-MM-DD", hasta)
		}
		d1 = d1.Add(24*time.Hour - time.Nanosecond)
	}
	buscar = strings.ToLower(strings.TrimSpace(buscar))
	var out []movimiento
	for _, m := range ms {
		t := m.Cuando.UTC()
		if desde != "" && t.Before(d0) {
			continue
		}
		if hasta != "" && t.After(d1) {
			continue
		}
		if buscar != "" && !strings.Contains(strings.ToLower(m.Direccion), buscar) && !strings.Contains(strings.ToLower(m.TxID), buscar) {
			continue
		}
		out = append(out, m)
	}
	return out, nil
}

func historial(conf *historialConfig) error {
	params := conf.NetParams()
	// 1) Direcciones con saldo, del daemon (incluye las de cambio).
	daemonClient, tearDown, err := client.Connect(conf.DaemonAddress)
	if err != nil {
		return traducirErrorDaemon(err, conf.DaemonAddress, banderaRed())
	}
	defer tearDown()
	ctx, cancel := context.WithTimeout(context.Background(), daemonTimeout)
	defer cancel()
	bal, err := daemonClient.GetBalance(ctx, &pb.GetBalanceRequest{})
	if err != nil {
		return traducirErrorDaemon(err, conf.DaemonAddress, banderaRed())
	}
	var direcciones []string
	for _, a := range bal.AddressBalances {
		if a.Available > 0 || a.Pending > 0 {
			direcciones = append(direcciones, a.Address)
		}
	}
	// 2) Los UTXOs, del nodo, con el DAA de cada uno.
	rpcServer := conf.RPCServer
	if rpcServer == "" {
		rpcServer = "127.0.0.1:" + params.RPCPort
	}
	c, err := rpcclient.NewRPCClient(rpcServer)
	if err != nil {
		return fmt.Errorf(T("historial.nodo.noresponde"), rpcServer, err)
	}
	defer c.Close()
	dag, err := c.GetBlockDAGInfo()
	if err != nil {
		return fmt.Errorf(T("historial.nodo.noresponde"), rpcServer, err)
	}
	var entradas []*appmessage.UTXOsByAddressesEntry
	if len(direcciones) > 0 {
		resp, err := c.GetUTXOsByAddresses(direcciones)
		if err != nil {
			return traducirErrorNodo(err)
		}
		entradas = resp.Entries
	}
	ahora := time.Now()
	ms := movimientosDesdeUTXOs(entradas, dag.VirtualDAAScore, ahora)
	ms, err = filtrar(ms, conf.Desde, conf.Hasta, conf.Buscar)
	if err != nil {
		return err
	}
	imprimirHistorial(ms, conf, dag.VirtualDAAScore, len(direcciones))
	return nil
}

const historialMaxFilas = 40

func imprimirHistorial(ms []movimiento, conf *historialConfig, daaVirtual uint64, nDirecciones int) {
	fmt.Printf(T("historial.titulo")+"\n", nDirecciones, miles(daaVirtual))
	if len(ms) == 0 {
		fmt.Println("  " + T("historial.vacio"))
		fmt.Println()
		fmt.Println(T("historial.nota"))
		return
	}
	var dias []diaMinado
	resto := ms
	if !conf.SinAgrupar {
		dias, resto = agruparMinado(ms)
	}
	var gold, nGemas uint64
	var gemas [5]uint64
	for _, m := range ms {
		if m.Nivel == 0 {
			gold += m.Monto
		} else if m.Nivel <= 4 {
			gemas[m.Nivel]++
			nGemas++
		}
	}
	// Minado por dia.
	if len(dias) > 0 {
		fmt.Println()
		fmt.Println(T("historial.minado"))
		mostrados := 0
		for _, d := range dias {
			if !conf.Todo && mostrados >= historialMaxFilas {
				fmt.Printf("  "+T("historial.mas")+"\n", len(dias)-mostrados)
				break
			}
			fmt.Printf("  %s  ⛏ %6s %-8s  %s\n", d.Dia, miles(uint64(d.Bloques)), T("historial.bloques"), rupixTxt(d.Monto))
			mostrados++
		}
	}
	// El resto, uno por uno.
	if len(resto) > 0 {
		fmt.Println()
		fmt.Println(T("historial.movimientos"))
		mostrados := 0
		for _, m := range resto {
			if !conf.Todo && mostrados >= historialMaxFilas {
				fmt.Printf("  "+T("historial.mas")+"\n", len(resto)-mostrados)
				break
			}
			que := rupixTxt(m.Monto)
			origen := T("historial.recibido")
			switch {
			case m.Nivel > 0 && m.Nivel <= 4:
				que = nombreNivel(m.Nivel)
				origen = T("historial.gema")
			case m.Coinbase:
				origen = T("historial.minado.uno")
			}
			fmt.Printf("  ≈%s  DAA %-11s %-10s %-22s %s  tx %s…\n", m.Cuando.Local().Format("2006-01-02 15:04"), miles(m.DAA), origen, que, acortar(m.Direccion), acortar(m.TxID))
			mostrados++
		}
	}
	fmt.Println()
	fmt.Printf(T("historial.totales")+"\n", rupixTxt(gold), gemas[1], gemas[2], gemas[3], gemas[4])
	fmt.Println()
	fmt.Println(T("historial.nota"))
}

func acortar(s string) string {
	if len(s) <= 18 {
		return s
	}
	return s[:10] + "…" + s[len(s)-6:]
}
