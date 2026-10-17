// Rupix Explorer - proxy RPC->HTTP
// La ventana publica de Rupix: sirve el frontend y traduce las consultas
// del navegador al RPC del nodo. Sin base de datos - el nodo es la verdad.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/rupixnet/rupixd/infrastructure/network/rpcclient"
)

var rpcAddress = flag.String("rpcserver", "127.0.0.1:17210", "direccion RPC del nodo rupixd")
var listenAddress = flag.String("listen", "127.0.0.1:8090", "direccion HTTP del explorador")
var webDir = flag.String("webdir", "/root/rupix-explorer-web", "carpeta del frontend")
var geoDB = flag.String("geodb", "", "CSV de DB-IP lite (ciudad) para ubicar nodos por pais/ciudad aproximada; vacio = sin mapa")
var nodoPropioFlag = flag.String("nodo-propio", "", "este nodo en el mapa: \"lat,lon,PAIS,Ciudad,Nombre\"")

func main() {
	flag.Parse()
	client, err := rpcclient.NewRPCClient(*rpcAddress)
	if err != nil {
		panic(fmt.Sprintf("no pude conectar al nodo: %s", err))
	}
	geo, err := cargarGeoBase(*geoDB)
	if err != nil {
		panic(fmt.Sprintf("no pude cargar la base de ubicacion %s: %s", *geoDB, err))
	}
	if geo != nil {
		fmt.Printf("mapa de nodos: %d rangos IPv4 cargados de %s\n", len(geo.inicio), *geoDB)
	}
	propio := nodoPropio(*nodoPropioFlag)
	if propio != nil {
		if info, err := client.GetInfo(); err == nil {
			propio.Version = strings.TrimPrefix(info.ServerVersion, "v")
		}
	}
	fmt.Printf("Rupix Explorer proxy: nodo=%s http=%s\n", *rpcAddress, *listenAddress)

	http.HandleFunc("/api/dag", func(w http.ResponseWriter, r *http.Request) {
		resp, err := client.GetBlockDAGInfo()
		writeJSON(w, resp, err)
	})
	http.HandleFunc("/api/supply", func(w http.ResponseWriter, r *http.Request) {
		resp, err := client.GetCoinSupply()
		writeJSON(w, resp, err)
	})
	// Rupix (7-oct-2026): la escalera en vivo. Nacidas = conteo sellado en cada
	// encabezado; vivas = indice de UTXOs del nodo. Dos fuentes, a proposito.
	http.HandleFunc("/api/gems", func(w http.ResponseWriter, r *http.Request) {
		resp, err := client.GetGemsInfo()
		writeJSON(w, resp, err)
	})
	// Rupix (14-oct-2026): el mapa de nodos. Nombre elegido (--uacomment), version,
	// antiguedad y ciudad aproximada; nunca la IP (ver nodos.go).
	http.HandleFunc("/api/peers", func(w http.ResponseWriter, r *http.Request) {
		resp, err := client.GetConnectedPeerInfo()
		if err != nil {
			writeJSON(w, nil, err)
			return
		}
		nodos := nodosDesdePeers(resp.Infos, geo, time.Now())
		writeJSON(w, &RespuestaNodos{Propio: propio, Nodos: nodos, Total: len(nodos), Paises: contarPaises(nodos, propio), ConGeo: geo != nil,
			Nota: "Solo lo que cada nodo anuncia (nombre con --uacomment, version) y una ubicacion aproximada por pais/ciudad. La IP no se publica ni se guarda."}, nil)
	})
	http.HandleFunc("/api/hashrate", func(w http.ResponseWriter, r *http.Request) {
		resp, err := client.EstimateNetworkHashesPerSecond("", 1000)
		writeJSON(w, resp, err)
	})
	http.HandleFunc("/api/block", func(w http.ResponseWriter, r *http.Request) {
		hash := r.URL.Query().Get("hash")
		resp, err := client.GetBlock(hash, true)
		writeJSON(w, resp, err)
	})
	http.HandleFunc("/api/blocks", func(w http.ResponseWriter, r *http.Request) {
		low := r.URL.Query().Get("low")
		resp, err := client.GetBlocks(low, true, true)
		writeJSON(w, resp, err)
	})

	// /mapa y /map llevan a la pagina del mapa de nodos.
	for _, ruta := range []string{"/mapa", "/map"} {
		http.HandleFunc(ruta, func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/mapa.html", http.StatusFound) })
	}
	// Frontend: servir la carpeta web
	http.Handle("/", http.FileServer(http.Dir(*webDir)))

	err = http.ListenAndServe(*listenAddress, nil)
	if err != nil {
		panic(err)
	}
}

func writeJSON(w http.ResponseWriter, data interface{}, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if err != nil {
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(data)
}
