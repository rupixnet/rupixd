package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/rupixnet/rupixd/infrastructure/network/rpcclient"
	"github.com/rupixnet/rupixd/version"
)

// verificar (Rupix, 7-oct-2026): una sola pantalla, sin contrasena, con lo que cualquier
// usuario puede comprobar por si mismo y nada que tenga que creer:
//  1. este binario: version y si su sha256 es el que publico la release;
//  2. el nodo: version, red, bloque, sincronizado, peers, punto de poda;
//  3. los checkpoints compilados: si el nodo tiene cada uno como bloque conocido;
//  4. la escalera: nacidas y vivas por nivel, topes, y que las dos fuentes de Kings coincidan.
//
// Cada linea empieza con OK, AVISO o FALLO. Si algo no se pudo comprobar, lo dice; no inventa.
func verificar(conf *verificarConfig) error {
	params := conf.NetParams()
	ok, aviso, fallo := "OK   ", "AVISO", "FALLO"
	linea := func(estado, formato string, a ...interface{}) {
		fmt.Printf("%s  %s\n", estado, fmt.Sprintf(formato, a...))
	}
	fallos := 0

	// 1) El binario.
	fmt.Println(T("verificar.binario"))
	ver := version.Version()
	linea(ok, T("verificar.version"), ver, runtime.GOOS, runtime.GOARCH)
	exe, err := os.Executable()
	if err == nil {
		exe, _ = filepath.EvalSymlinks(exe)
	}
	hashLocal, err := sha256DeArchivo(exe)
	if err != nil {
		linea(aviso, T("verificar.sinhash"), err)
	} else {
		linea(ok, "sha256 %s  %s", filepath.Base(exe), hashLocal)
		if conf.SinInternet {
			linea(aviso, T("verificar.sininternet"))
		} else {
			publicado, err := sha256Publicado(ver, filepath.Base(exe))
			switch {
			case err != nil:
				linea(aviso, T("verificar.release.noleida"), ver, err)
			case publicado == "":
				linea(aviso, T("verificar.release.sinlista"), ver)
			case publicado == hashLocal:
				linea(ok, T("verificar.release.igual"), ver)
			default:
				fallos++
				linea(fallo, T("verificar.release.distinto"), ver, publicado)
			}
		}
	}

	// 2) El nodo.
	fmt.Println()
	fmt.Println(T("verificar.nodo"))
	rpcServer := conf.RPCServer
	if rpcServer == "" {
		rpcServer = "127.0.0.1:" + params.RPCPort
	}
	c, err := rpcclient.NewRPCClient(rpcServer)
	if err != nil {
		linea(fallo, T("verificar.nodo.noresponde"), rpcServer, err)
		fmt.Println()
		fmt.Printf(T("verificar.resumen")+"\n", fallos+1)
		return nil
	}
	defer c.Close()
	info, err := c.GetInfo()
	if err != nil {
		linea(fallo, T("verificar.nodo.noresponde"), rpcServer, err)
		return nil
	}
	dag, err := c.GetBlockDAGInfo()
	if err != nil {
		linea(fallo, T("verificar.nodo.noresponde"), rpcServer, err)
		return nil
	}
	if strings.TrimPrefix(info.ServerVersion, "v") != ver {
		linea(aviso, T("verificar.nodo.otraversion"), info.ServerVersion, ver)
	} else {
		linea(ok, T("verificar.nodo.version"), info.ServerVersion)
	}
	if dag.NetworkName != params.Name {
		fallos++
		linea(fallo, T("verificar.nodo.otrared"), dag.NetworkName, params.Name)
	} else {
		linea(ok, T("verificar.nodo.red"), dag.NetworkName)
	}
	estadoSync := ok
	if !info.IsSynced {
		estadoSync = aviso
	}
	linea(estadoSync, T("verificar.nodo.bloque"), dag.VirtualDAAScore, dag.BlockCount, info.IsSynced)
	peers, err := c.GetConnectedPeerInfo()
	if err == nil {
		estadoPeers := ok
		if len(peers.Infos) == 0 {
			estadoPeers = aviso
		}
		linea(estadoPeers, T("verificar.nodo.peers"), len(peers.Infos))
	}
	linea(ok, T("verificar.nodo.poda"), dag.PruningPointHash)
	if !info.IsUtxoIndexed {
		linea(aviso, T("verificar.nodo.sinindice"))
	}

	// 3) Los checkpoints compilados en este binario, contra el nodo.
	fmt.Println()
	fmt.Println(T("verificar.checkpoints"))
	if len(params.Checkpoints) == 0 {
		linea(ok, T("verificar.cp.ninguno"))
	}
	for i, cp := range params.Checkpoints {
		vigente := dag.VirtualDAAScore < params.CheckpointsExpireDAAScore
		_, err := c.GetBlock(cp.Hash.String(), false)
		switch {
		case err != nil && vigente:
			fallos++
			linea(fallo, T("verificar.cp.falta"), i+1, cp.BlueScore, cp.Hash)
		case err != nil:
			linea(aviso, T("verificar.cp.faltacaducado"), i+1, cp.BlueScore, cp.Hash)
		case vigente:
			linea(ok, T("verificar.cp.tiene"), i+1, cp.BlueScore, cp.Hash, params.CheckpointsExpireDAAScore)
		default:
			linea(ok, T("verificar.cp.caducado"), i+1, cp.BlueScore, cp.Hash, params.CheckpointsExpireDAAScore)
		}
	}

	// 4) La escalera.
	fmt.Println()
	fmt.Println(T("verificar.gemas"))
	gi, err := c.GetGemsInfo()
	if err != nil {
		linea(aviso, T("verificar.gemas.noleidas"), err)
	} else {
		fmt.Printf("       %-9s %10s %10s %10s\n", "", T("verificar.gemas.nacidas"), T("verificar.gemas.vivas"), T("verificar.gemas.tope"))
		fmt.Printf("       %-9s %10d %10d %10d\n", nombreNivel(1), gi.DiamantesNacidos, gi.DiamantesVivos, gi.TopeDiamantes)
		fmt.Printf("       %-9s %10d %10d %10d\n", nombreNivel(2), gi.PlatinosNacidos, gi.PlatinosVivos, gi.TopePlatinos)
		fmt.Printf("       %-9s %10d %10d %10d\n", nombreNivel(3), gi.RodiosNacidos, gi.RodiosVivos, gi.TopeRodios)
		fmt.Printf("       %-9s %10d %10d %10d\n", nombreNivel(4), gi.KingsNacidos, gi.KingsVivos, gi.TopeKings)
		if gi.KingsNacidos != gi.KingsVivos {
			fallos++
			linea(fallo, T("verificar.gemas.kings"), gi.KingsNacidos, gi.KingsVivos)
		} else {
			linea(ok, T("verificar.gemas.kingsok"), gi.KingsNacidos)
		}
		if gi.DiamantesVivos > gi.DiamantesNacidos || gi.PlatinosVivos > gi.PlatinosNacidos || gi.RodiosVivos > gi.RodiosNacidos {
			fallos++
			linea(fallo, T("verificar.gemas.masvivas"))
		}
	}

	fmt.Println()
	fmt.Printf(T("verificar.resumen")+"\n", fallos)
	return nil
}

func sha256DeArchivo(ruta string) (string, error) {
	f, err := os.Open(ruta)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// sha256Publicado baja la lista SHA256SUMS-<os>.txt de la release de esta version y
// devuelve el hash publicado para el binario con ese nombre ("" si la lista no trae ese
// nombre o la release no tiene lista; error si no se pudo leer).
func sha256Publicado(ver, nombreBinario string) (string, error) {
	so := map[string]string{"linux": "linux", "windows": "win64", "darwin": "osx"}[runtime.GOOS]
	if so == "" {
		return "", fmt.Errorf("sistema sin release publicada: %s", runtime.GOOS)
	}
	url := fmt.Sprintf("https://github.com/rupixnet/rupixd/releases/download/v%s/SHA256SUMS-%s.txt", ver, so)
	cliente := &http.Client{Timeout: 10 * time.Second}
	resp, err := cliente.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return "", nil
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	cuerpo, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return "", err
	}
	return hashDeLista(string(cuerpo), nombreBinario), nil
}

// hashDeLista busca en una salida de `sha256sum` ("<hash>  <nombre>") el hash del nombre dado.
func hashDeLista(lista, nombre string) string {
	for _, l := range strings.Split(lista, "\n") {
		campos := strings.Fields(l)
		if len(campos) >= 2 && (campos[1] == nombre || campos[1] == "*"+nombre || filepath.Base(campos[1]) == nombre) {
			return strings.ToLower(campos[0])
		}
	}
	return ""
}
