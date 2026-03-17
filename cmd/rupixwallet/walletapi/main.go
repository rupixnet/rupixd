package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "time"

    "github.com/rupixnet/rupixd/cmd/rupixwallet/daemon/client"
    "github.com/rupixnet/rupixd/cmd/rupixwallet/daemon/pb"
)

const daemonAddr = "localhost:8082"
const timeout = 30 * time.Second

func cors(h http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
        if r.Method == "OPTIONS" { w.WriteHeader(200); return }
        h(w, r)
    }
}

func json200(w http.ResponseWriter, v interface{}) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(v)
}

func jsonErr(w http.ResponseWriter, err error) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(500)
    json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}

func connect() (pb.RupixwalletdClient, func(), error) {
    return client.Connect(daemonAddr)
}

func handleBalance(w http.ResponseWriter, r *http.Request) {
    c, td, err := connect()
    if err != nil { jsonErr(w, err); return }
    defer td()
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()
    resp, err := c.GetBalance(ctx, &pb.GetBalanceRequest{})
    if err != nil { jsonErr(w, err); return }
    json200(w, map[string]interface{}{
        "available": resp.Available,
        "pending":   resp.Pending,
        "addresses": resp.AddressBalances,
    })
}

func handleAddresses(w http.ResponseWriter, r *http.Request) {
    c, td, err := connect()
    if err != nil { jsonErr(w, err); return }
    defer td()
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()
    resp, err := c.ShowAddresses(ctx, &pb.ShowAddressesRequest{})
    if err != nil { jsonErr(w, err); return }
    json200(w, map[string]interface{}{"addresses": resp.Address})
}

func handleNewAddress(w http.ResponseWriter, r *http.Request) {
    c, td, err := connect()
    if err != nil { jsonErr(w, err); return }
    defer td()
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()
    resp, err := c.NewAddress(ctx, &pb.NewAddressRequest{})
    if err != nil { jsonErr(w, err); return }
    json200(w, map[string]string{"address": resp.Address})
}

func handleSend(w http.ResponseWriter, r *http.Request) {
    var req struct {
        ToAddress string  `json:"toAddress"`
        Amount    uint64  `json:"amount"`
        Password  string  `json:"password"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil { jsonErr(w, err); return }
    c, td, err := connect()
    if err != nil { jsonErr(w, err); return }
    defer td()
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()
    resp, err := c.Send(ctx, &pb.SendRequest{
        ToAddress: req.ToAddress,
        Amount:    req.Amount,
        Password:  req.Password,
    })
    if err != nil { jsonErr(w, err); return }
    json200(w, map[string]interface{}{"txIDs": resp.TxIDs})
}

func handleVersion(w http.ResponseWriter, r *http.Request) {
    c, td, err := connect()
    if err != nil { jsonErr(w, err); return }
    defer td()
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()
    resp, err := c.GetVersion(ctx, &pb.GetVersionRequest{})
    if err != nil { jsonErr(w, err); return }
    json200(w, map[string]string{"version": resp.Version})
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
    c, td, err := connect()
    if err != nil {
        json200(w, map[string]interface{}{"connected": false, "error": err.Error()})
        return
    }
    defer td()
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    resp, err := c.GetVersion(ctx, &pb.GetVersionRequest{})
    if err != nil {
        json200(w, map[string]interface{}{"connected": false})
        return
    }
    json200(w, map[string]interface{}{"connected": true, "version": resp.Version})
}

func main() {
    http.HandleFunc("/api/status",      cors(handleStatus))
    http.HandleFunc("/api/balance",     cors(handleBalance))
    http.HandleFunc("/api/addresses",   cors(handleAddresses))
    http.HandleFunc("/api/new-address", cors(handleNewAddress))
    http.HandleFunc("/api/send",        cors(handleSend))
    http.HandleFunc("/api/version",     cors(handleVersion))

    fmt.Println("Rupix Wallet API corriendo en http://localhost:8083")
    log.Fatal(http.ListenAndServe(":8083", nil))
}