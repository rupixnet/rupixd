package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/rupixnet/rupixd/cmd/rupixwallet/keys"
)

// Idioma (Rupix, v0.6.2): la wallet habla espanol o ingles, y se puede cambiar en
// cualquier momento. Orden de eleccion:
//   1. --lang=es|en en el comando
//   2. la variable de entorno RUPIX_LANG
//   3. el archivo "idioma" junto a las llaves (lo escribe `rupixwallet language en`)
//   4. el idioma del sistema (LANG / LC_ALL / LC_MESSAGES)
//   5. espanol
// Solo cambia lo que ve la persona. El consenso y los mensajes del nodo no se tocan.

var idiomaActivo = "es"

var idiomasValidos = map[string]bool{"es": true, "en": true}

// textos: clave -> {espanol, ingles}. Los %s/%d se rellenan con fmt como siempre.
var textos = map[string][2]string{
	"nivel.0": {"Gold", "Gold"},
	"nivel.1": {"Diamante", "Diamond"},
	"nivel.2": {"Platino", "Platinum"},
	"nivel.3": {"Rodio", "Rhodium"},
	"nivel.4": {"Kings", "Kings"},

	"clave.prompt":   {"Contrasena:", "Password:"},
	"confirmar":      {"Confirmas? Escribe 'si' para continuar: ", "Confirm? Type 'yes' to continue: "},
	"cancelado":      {"Cancelado. No se movio nada.", "Cancelled. Nothing moved."},
	"dump.aviso":     {"ATENCION: este comando muestra tu FRASE SEMILLA en pantalla. Quien la vea puede vaciar tu wallet. Hazlo sin conexion, sin nadie mirando, y copiala a papel.", "WARNING: this command shows your SEED PHRASE on screen. Anyone who sees it can empty your wallet. Do it offline, with nobody watching, and copy it to paper."},
	"dump.confirmar": {"Entiendes el riesgo? Escribe 'si' para mostrarla: ", "Do you understand the risk? Type 'yes' to show it: "},
	"dump.frase":     {"Frase semilla #%d:\n%s\n\n", "Seed phrase #%d:\n%s\n\n"},
	"dump.cancelado": {"Cancelado. No se mostro nada.", "Cancelled. Nothing was shown."},
	"dump.limpia":    {"Ya la anotaste? Limpia la pantalla: clear (Linux/Mac) o cls (Windows).", "Written down? Clear the screen: clear (Linux/Mac) or cls (Windows)."},

	"forjar.nivel_invalido":  {"--level debe ser 1 (Diamante), 2 (Platino), 3 (Rodio) o 4 (Kings)", "--level must be 1 (Diamond), 2 (Platinum), 3 (Rhodium) or 4 (Kings)"},
	"wallet.sin_direcciones": {"la wallet no tiene direcciones todavia; crea una con new-address", "the wallet has no addresses yet; create one with new-address"},
	"falta.gold_diamante":    {"Gold: tienes %s y necesitas al menos %s (10 que se queman + comision)", "Gold: you have %s and need at least %s (10 burned + fee)"},
	"falta.gemas":            {"%s: tienes %d y necesitas %d", "%s: you have %d and need %d"},
	"falta.gold_comision":    {"Gold para la comision: tienes %s y necesitas al menos 2", "Gold for the fee: you have %s and need at least 2"},
	"aviso.nodo":             {"Aviso: no pude preguntarle al nodo en que bloque va (%s). La red revisara el nivel de todos modos.", "Note: I couldn't ask the node which block it is at (%s). The network will check the level anyway."},
	"falta.nivel_cerrado":    {"el %s se abre en el bloque %s y la red va en el %s: faltan %s bloques (~%s)", "%s opens at block %s and the network is at %s: %s blocks to go (~%s)"},
	"forjar.todavia_no":      {"Todavia no se puede forjar:", "Can't forge yet:"},
	"forjar.cancelada":       {"forja cancelada: no se mando nada a la red", "forge cancelled: nothing was sent to the network"},
	"forjar.resumen":         {"Vas a forjar 1 %s en %s", "You are about to forge 1 %s at %s"},
	"forjar.quema_gold":      {"Se queman 10 Gold PARA SIEMPRE, mas una comision pequena.", "10 Gold are burned FOREVER, plus a small fee."},
	"forjar.quema_gemas":     {"Se queman %d %s PARA SIEMPRE (y una comision pequena en Gold). Quedaran %d %s.", "%d %s are burned FOREVER (plus a small fee in Gold). %d %s will remain."},
	"red.bloque":             {"La red va en el bloque %s; el %s esta abierto desde el %s.", "The network is at block %s; %s has been open since block %s."},
	"forjar.hecho":           {"Ascenso forjado: gema %s creada.", "Forged: %s gem created."},
	"forjar.verifica":        {"Verificalo desde cualquier nodo: rupixctl %sGetUtxosByAddresses %s (busca version = %d).", "Verify it from any node: rupixctl %sGetUtxosByAddresses %s (look for version = %d)."},

	"enviar.todo":    {"Vas a enviar TODO el Gold de la wallet a %s en %d transaccion(es).", "You are about to send ALL the Gold in the wallet to %s in %d transaction(s)."},
	"enviar.resumen": {"Vas a enviar %s RUPIX a %s en %d transaccion(es).", "You are about to send %s RUPIX to %s in %d transaction(s)."},
	"enviar.lotes":   {"(Se juntan pedazos chicos en lotes y la ultima paga; tarda ~%d s.)", "(Small pieces get batched and the last one pays; takes ~%d s.)"},

	"gema.enviar":  {"Vas a enviar 1 %s a %s. La gema deja de ser tuya.", "You are about to send 1 %s to %s. The gem stops being yours."},
	"gema.enviada": {"Gema %s transferida a %s", "%s gem transferred to %s"},

	"gemas.titulo": {"La escalera — tus gemas:", "The ladder — your gems:"},
	"gemas.total":  {"  Total: %d gema(s)", "  Total: %d gem(s)"},

	"saldo.total":     {"Tienes %s RUPIX", "You have %s RUPIX"},
	"saldo.pendiente": {" (+ %s llegando)", " (+ %s arriving)"},
	"saldo.cabecera":  {"Direccion                                                                     Disponible            Llegando", "Address                                                                       Available             Arriving"},

	"direccion.nueva": {"Direccion nueva:", "New address:"},

	"crear.guardado": {"Llaves guardadas en %s", "Keys saved to %s"},
	"crear.frase.1":  {"FRASE SEMILLA — copiala en PAPEL, ahora.", "SEED PHRASE — copy it on PAPER, now."},
	"crear.frase.2":  {"Es la unica forma de recuperar esta wallet si pierdes el", "It is the only way to recover this wallet if you lose the"},
	"crear.frase.3":  {"archivo o la contrasena. No la guardes en el telefono ni", "file or the password. Don't keep it on your phone or"},
	"crear.frase.4":  {"en el chat. Se muestra UNA sola vez.", "in a chat. It is shown ONCE."},
	"crear.frase.n":  {"Frase #%d:", "Phrase #%d:"},
	"crear.limpia":   {"Cuando la tengas en papel, limpia la pantalla (clear / cls).", "Once it is on paper, clear the screen (clear / cls)."},
	"crear.verifica": {"Para asegurarnos de que la anotaste bien, escribe la palabra numero %d de la frase #%d: ", "To make sure you wrote it down, type word number %d of phrase #%d: "},
	"crear.mal":      {"No coincide. Revisa tu papel: la palabra %d de la frase #%d es otra. La wallet ya esta creada; vuelve a mirar la frase arriba.", "It doesn't match. Check your paper: word %d of phrase #%d is different. The wallet is already created; look at the phrase above again."},
	"crear.bien":     {"Bien: la frase esta en tu papel. Ahora limpia la pantalla (clear / cls).", "Good: the phrase is on your paper. Now clear the screen (clear / cls)."},
	"daemon.apagado": {"El daemon de la wallet no esta corriendo (o no en %s). Abrelo en OTRA ventana y dejalo ahi:\n  rupixwallet %s start-daemon\nLuego repite este comando aqui.", "The wallet daemon isn't running (or not at %s). Start it in ANOTHER window and leave it there:\n  rupixwallet %s start-daemon\nThen run this command again here."},

	"tiempo.seg":  {"%d segundos", "%d seconds"},
	"tiempo.min":  {"%d minutos", "%d minutes"},
	"tiempo.hora": {"%.1f horas", "%.1f hours"},
	"tiempo.dia":  {"%.1f dias", "%.1f days"},

	"err.nivel_cerrado": {"La red rechazo la forja porque ese nivel todavia no se abre (cada nivel se abre en su halving). Espera a que la red llegue al bloque que indica el mensaje.", "The network rejected the forge because that level isn't open yet (each level opens at its halving). Wait until the network reaches the block in the message."},
	"err.no_sync":       {"Tu wallet todavia no esta al dia con la red. Espera unos minutos con el daemon corriendo y vuelve a intentar.", "Your wallet isn't caught up with the network yet. Wait a few minutes with the daemon running and try again."},
	"err.gold":          {"No alcanza el Gold: la forja quema Gold y ademas paga una comision pequena.", "Not enough Gold: forging burns Gold and also pays a small fee."},
	"err.clave":         {"La contrasena no es la de esta wallet (o el daemon corre con otro archivo de llaves).", "That isn't this wallet's password (or the daemon runs with a different keys file)."},
	"err.nodo_dijo":     {"%s\n  (el nodo dijo: %s)", "%s\n  (the node said: %s)"},

	"idioma.guardado": {"Idioma guardado: %s. Se usa en todos los comandos desde ahora.", "Language saved: %s. It applies to every command from now on."},
	"idioma.invalido": {"idioma no reconocido: %q (usa es o en)", "unknown language: %q (use es or en)"},
	"idioma.actual":   {"Idioma actual: %s (origen: %s)", "Current language: %s (source: %s)"},
}

// T devuelve el texto de una clave en el idioma activo. Una clave desconocida se
// devuelve tal cual, para que un olvido se vea en pantalla y no se esconda.
func T(clave string) string {
	t, ok := textos[clave]
	if !ok {
		return clave
	}
	if idiomaActivo == "en" {
		return t[1]
	}
	return t[0]
}

// nombreNivel devuelve el nombre del nivel (0..4) en el idioma activo.
func nombreNivel(nivel uint32) string {
	return T(fmt.Sprintf("nivel.%d", nivel))
}

func archivoIdioma() string {
	return filepath.Join(keys.DefaultAppDir(), "idioma")
}

// elegirIdioma decide el idioma activo (ver el orden arriba) y dice de donde salio.
func elegirIdioma(flagLang string) (idioma, origen string) {
	defer func() { idiomaActivo = idioma }()
	if v := strings.ToLower(strings.TrimSpace(flagLang)); idiomasValidos[v] {
		return v, "--lang"
	}
	if v := strings.ToLower(strings.TrimSpace(os.Getenv("RUPIX_LANG"))); idiomasValidos[v] {
		return v, "RUPIX_LANG"
	}
	if b, err := os.ReadFile(archivoIdioma()); err == nil {
		if v := strings.ToLower(strings.TrimSpace(string(b))); idiomasValidos[v] {
			return v, archivoIdioma()
		}
	}
	for _, k := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		v := strings.ToLower(os.Getenv(k))
		if strings.HasPrefix(v, "en") {
			return "en", k
		}
		if strings.HasPrefix(v, "es") {
			return "es", k
		}
	}
	return "es", "default"
}

// language (comando): `rupixwallet language en` guarda la preferencia; sin argumento la muestra.
func language(conf *languageConfig) error {
	nuevo := strings.ToLower(strings.TrimSpace(conf.Args.Lang))
	if nuevo == "" {
		idioma, origen := elegirIdioma("")
		fmt.Printf(T("idioma.actual")+"\n", idioma, origen)
		return nil
	}
	if !idiomasValidos[nuevo] {
		return errors.Errorf(T("idioma.invalido"), nuevo)
	}
	if err := os.MkdirAll(filepath.Dir(archivoIdioma()), 0700); err != nil {
		return err
	}
	if err := os.WriteFile(archivoIdioma(), []byte(nuevo+"\n"), 0600); err != nil {
		return err
	}
	idiomaActivo = nuevo
	fmt.Printf(T("idioma.guardado")+"\n", nuevo)
	return nil
}
