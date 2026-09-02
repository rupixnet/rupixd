// Rupix Explorer - proxy RPC->HTTP
// La ventana publica de Rupix: sirve el frontend y traduce las consultas
// del navegador al RPC del nodo. Sin base de datos - el nodo es la verdad.
package main

import (
"encoding/json"
"flag"
"fmt"
"net/http"

"github.com/rupixnet/rupixd/infrastructure/network/rpcclient"
)

var rpcAddress = flag.String("rpcserver", "127.0.0.1:17210", "direccion RPC del nodo rupixd")
var listenAddress = flag.String("listen", "127.0.0.1:8090", "direccion HTTP del explorador")
var webDir = flag.String("webdir", "/root/rupix-explorer-web", "carpeta del frontend")

func main() {
flag.Parse()
client, err := rpcclient.NewRPCClient(*rpcAddress)
if err != nil {
panic(fmt.Sprintf("no pude conectar al nodo: %s", err))
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
