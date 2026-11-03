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
	"forjar.resumen":         {"Vas a forjar %d %s en %s", "You are about to forge %d %s at %s"},
	"forjar.quema_gold":      {"Se queman %d Gold PARA SIEMPRE, mas una comision pequena por gema.", "%d Gold are burned FOREVER, plus a small fee per gem."},
	"forjar.count_invalido":  {"--count debe estar entre 1 y 1000", "--count must be between 1 and 1000"},
	"forjar.hecho_n":         {"Forja %d de %d: gema %s creada.", "Forge %d of %d: %s gem created."},
	"forjar.parcial":         {"Se forjaron %d de %d %s antes de parar.", "%d of %d %s were forged before stopping."},
	"forjar.no_aparecio":     {"la gema %s no aparecio en la wallet en %d s; la forja en serie se detiene (revisa con gems)", "the %s gem did not show up in the wallet within %d s; serial forging stops (check with gems)"},
	"forjar.quema_gemas":     {"Se queman %d %s PARA SIEMPRE (y una comision pequena en Gold). Quedaran %d %s.", "%d %s are burned FOREVER (plus a small fee in Gold). %d %s will remain."},
	"red.bloque":             {"La red va en el bloque %s; el %s esta abierto desde el %s.", "The network is at block %s; %s has been open since block %s."},
	"forjar.hecho":           {"Ascenso forjado: gema %s creada.", "Forged: %s gem created."},
	"forjar.verifica":        {"Verificalo desde cualquier nodo: rupixctl %sGetUtxosByAddresses %s (busca version = %d).", "Verify it from any node: rupixctl %sGetUtxosByAddresses %s (look for version = %d)."},

	"enviar.destino":           {"Destino: %s[%s]", "Destination: %s[%s]"},
	"enviar.escribe_final":     {"Escribe los ULTIMOS 4 caracteres de la direccion, leyendolos de tu papel o del mensaje de quien te la dio (no de esta pantalla): ", "Type the LAST 4 characters of the address, reading them from your paper or from the message of whoever gave it to you (not from this screen): "},
	"enviar.dir_mal":           {"No coinciden. No se envio nada. Si la direccion que te dieron termina distinto a la que ves aqui, algo en tu computadora la cambio: no la uses.", "They do not match. Nothing was sent. If the address you were given ends differently from the one shown here, something on your computer changed it: do not use it."},
	"crear.respaldo":           {"RESPALDO: tu dinero es este archivo + tu contrasena + la frase en papel.\n  Archivo: %s  (copialo a un USB; sin la frase, un disco muerto es dinero perdido)", "BACKUP: your money is this file + your password + the phrase on paper.\n  File: %s  (copy it to a USB stick; without the phrase, a dead disk is lost money)"},
	"verificar.respaldo":       {"== Tu respaldo", "== Your backup"},
	"verificar.respaldo.hay":   {"archivo de llaves en %s: respaldalo en un USB y ten la frase en papel; los dos juntos son tu dinero", "keys file at %s: back it up on a USB stick and keep the phrase on paper; together they are your money"},
	"verificar.respaldo.nohay": {"no hay archivo de llaves en %s (sin wallet en esta red, o esta en otra ruta: --keys-file)", "no keys file at %s (no wallet on this network, or it lives elsewhere: --keys-file)"},
	"enviar.todo":              {"Vas a enviar TODO el Gold de la wallet a %s en %d transaccion(es).", "You are about to send ALL the Gold in the wallet to %s in %d transaction(s)."},
	"enviar.resumen":           {"Vas a enviar %s RUPIX a %s en %d transaccion(es).", "You are about to send %s RUPIX to %s in %d transaction(s)."},
	"enviar.lotes":             {"(Se juntan pedazos chicos en lotes y la ultima paga; tarda ~%d s.)", "(Small pieces get batched and the last one pays; takes ~%d s.)"},

	"gema.enviar":  {"Vas a enviar 1 %s a %s. La gema deja de ser tuya.", "You are about to send 1 %s to %s. The gem stops being yours."},
	"gema.enviada": {"Gema %s transferida a %s", "%s gem transferred to %s"},

	"gemas.titulo": {"La escalera — tus gemas:", "The ladder — your gems:"},
	"gemas.total":  {"  Total: %d gema(s)", "  Total: %d gem(s)"},

	"tarjeta.llegando":   {"llegando (sin confirmar)", "arriving (unconfirmed)"},
	"tarjeta.nodo.aldia": {"Nodo al dia", "Node up to date"},
	"tarjeta.nodo.atras": {"Nodo poniendose al dia: el saldo puede estar atrasado", "Node catching up: the balance may be stale"},
	"tarjeta.nodo.sin":   {"Nodo: no se pudo comprobar (el daemon si contesto)", "Node: could not be checked (the daemon did answer)"},
	"tarjeta.gemas.sin":  {"AVISO: no se pudieron leer las gemas: %v", "WARNING: gems could not be read: %v"},
	"saldo.total":        {"Tienes %s RUPIX", "You have %s RUPIX"},
	"saldo.pendiente":    {" (+ %s llegando)", " (+ %s arriving)"},
	"saldo.cabecera":     {"Direccion                                                                     Disponible            Llegando", "Address                                                                       Available             Arriving"},

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

	"err.nivel_cerrado":  {"La red rechazo la forja porque ese nivel todavia no se abre (cada nivel se abre en su halving). Espera a que la red llegue al bloque que indica el mensaje.", "The network rejected the forge because that level isn't open yet (each level opens at its halving). Wait until the network reaches the block in the message."},
	"err.no_sync":        {"Tu wallet todavia no esta al dia con la red. Espera unos minutos con el daemon corriendo y vuelve a intentar.", "Your wallet isn't caught up with the network yet. Wait a few minutes with the daemon running and try again."},
	"err.gold":           {"No alcanza el Gold: la forja quema Gold y ademas paga una comision pequena.", "Not enough Gold: forging burns Gold and also pays a small fee."},
	"err.clave":          {"La contrasena no es la de esta wallet (o el daemon corre con otro archivo de llaves).", "That isn't this wallet's password (or the daemon runs with a different keys file)."},
	"err.daemon_sync":    {"Tu wallet todavia esta poniendose al dia con la red (el daemon sincroniza sus direcciones). Espera uno o dos minutos y repite el comando; mientras, `rupixwallet %s verificar` te dice si el nodo ya esta sincronizado.", "Your wallet is still catching up with the network (the daemon is syncing its addresses). Wait a minute or two and retry; meanwhile `rupixwallet %s verificar` tells you whether the node is synced."},
	"err.saldo":          {"No alcanza el saldo para ese envio (monto + comision). Mira cuanto tienes con `rupixwallet %s balance`; si acabas de recibir, lo que llega como 'pendiente' aun no se puede gastar.", "Not enough balance for that send (amount + fee). Check what you have with `rupixwallet %s balance`; if you just received funds, what shows as 'pending' cannot be spent yet."},
	"err.direccion":      {"Esa direccion no es valida para esta red. Una direccion de testnet empieza con rupixtest: y la de mainnet con rupix:; copiala completa, sin espacios. (La bandera %s decide la red.)", "That address is not valid for this network. A testnet address starts with rupixtest: and a mainnet one with rupix:; copy it whole, with no spaces. (The %s flag decides the network.)"},
	"err.sin_wallet":     {"No hay wallet creada todavia en esta red. Crea una con `rupixwallet %s create` y guarda la frase en papel.", "No wallet has been created yet on this network. Create one with `rupixwallet %s create` and write the phrase down on paper."},
	"err.wallet_ocupada": {"Otro comando de la wallet esta usando el archivo ahora mismo (por ejemplo el daemon en otra ventana). Espera a que termine o cierra el otro. (red: %s)", "Another wallet command is using the file right now (for instance the daemon in another window). Wait for it to finish or close the other one. (network: %s)"},
	"err.nodo_dijo":      {"%s\n  (el nodo dijo: %s)", "%s\n  (the node said: %s)"},

	"idioma.guardado": {"Idioma guardado: %s. Se usa en todos los comandos desde ahora.", "Language saved: %s. It applies to every command from now on."},
	"idioma.invalido": {"idioma no reconocido: %q (usa es o en)", "unknown language: %q (use es or en)"},
	"idioma.actual":   {"Idioma actual: %s (origen: %s)", "Current language: %s (source: %s)"},

	"verificar.binario":          {"== Este binario", "== This binary"},
	"verificar.version":          {"rupixwallet %s (%s/%s)", "rupixwallet %s (%s/%s)"},
	"verificar.sinhash":          {"no se pudo calcular el sha256 de este binario: %v", "could not compute this binary's sha256: %v"},
	"verificar.sininternet":      {"sin internet (--sin-internet): no se comparo con la release", "no internet (--sin-internet): not compared with the release"},
	"verificar.release.noleida":  {"no se pudo leer la lista de hashes de la release v%s (%v); compara a mano con tools/verificar-binarios.sh", "could not read the release v%s hash list (%v); compare by hand with tools/verificar-binarios.sh"},
	"verificar.release.sinlista": {"la release v%s no publica lista de hashes por binario (desde v0.6.5 si); compara a mano con tools/verificar-binarios.sh", "release v%s publishes no per-binary hash list (since v0.6.5 it does); compare by hand with tools/verificar-binarios.sh"},
	"verificar.release.igual":    {"este binario es exactamente el de la release v%s", "this binary is exactly the one in release v%s"},
	"verificar.release.distinto": {"este binario NO es el de la release v%s (publicado: %s). No lo uses con dinero real.", "this binary is NOT the one in release v%s (published: %s). Do not use it with real money."},
	"verificar.nodo":             {"== El nodo", "== The node"},
	"verificar.nodo.noresponde":  {"el nodo en %s no responde: %v", "the node at %s does not answer: %v"},
	"verificar.nodo.otraversion": {"el nodo es %s y esta wallet es %s: actualiza el que vaya atras", "the node is %s and this wallet is %s: update whichever lags"},
	"verificar.nodo.version":     {"nodo %s", "node %s"},
	"verificar.nodo.otrared":     {"el nodo esta en %s y esta wallet en %s", "the node is on %s and this wallet on %s"},
	"verificar.nodo.red":         {"red %s", "network %s"},
	"verificar.nodo.bloque":      {"DAA %d, %d bloques, sincronizado: %v", "DAA %d, %d blocks, synced: %v"},
	"verificar.nodo.peers":       {"%d peers conectados", "%d connected peers"},
	"verificar.nodo.poda":        {"punto de poda %s", "pruning point %s"},
	"verificar.nodo.sinindice":   {"el nodo corre sin --utxoindex: la wallet y GetGemsInfo no funcionan contra el", "the node runs without --utxoindex: the wallet and GetGemsInfo cannot work against it"},
	"verificar.checkpoints":      {"== Checkpoints compilados en este binario, contra el nodo", "== Checkpoints compiled into this binary, against the node"},
	"verificar.cp.ninguno":       {"esta red no tiene checkpoints", "this network has no checkpoints"},
	"verificar.cp.falta":         {"checkpoint #%d (blue %d) %s: el nodo NO lo tiene. O no ha sincronizado hasta ahi, o esta en otra historia.", "checkpoint #%d (blue %d) %s: the node does NOT have it. Either not synced that far, or on a different history."},
	"verificar.nodo.nombre":      {"Tu nodo aparece en el mapa (explorer.rupix.network/mapa.html) con el nombre que anuncia; sin --uacomment=TuNombre sale como \"nodo\".", "Your node shows on the map (explorer.rupix.network/mapa.html) under the name it announces; without --uacomment=YourName it shows as \"node\"."},
	"verificar.cp.podado":        {"checkpoint #%d (blue %d) %s: anterior a tu punto de poda (blue %d). Tu nodo lo exigio al sincronizar (la lista de puntos de poda tenia que traerlo); ya no se puede releer por RPC. No es un fallo.", "checkpoint #%d (blue %d) %s: below your pruning point (blue %d). Your node enforced it when syncing (the pruning-point list had to include it); it can no longer be re-read over RPC. Not a failure."},
	"verificar.cp.faltacaducado": {"checkpoint #%d (blue %d) %s: el nodo no lo tiene (ya caduco; no afecta)", "checkpoint #%d (blue %d) %s: the node lacks it (already expired; no effect)"},
	"verificar.cp.tiene":         {"checkpoint #%d (blue %d) %s: el nodo lo tiene (vigente hasta DAA %d)", "checkpoint #%d (blue %d) %s: the node has it (in force until DAA %d)"},
	"verificar.cp.caducado":      {"checkpoint #%d (blue %d) %s: el nodo lo tiene (caduco en DAA %d)", "checkpoint #%d (blue %d) %s: the node has it (expired at DAA %d)"},
	"verificar.gemas":            {"== La escalera, segun el nodo", "== The ladder, according to the node"},
	"verificar.gemas.noleidas":   {"no se pudo leer GetGemsInfo: %v (el nodo necesita v0.6.5+ y --utxoindex)", "could not read GetGemsInfo: %v (node needs v0.6.5+ and --utxoindex)"},
	"verificar.gemas.nacidas":    {"nacidas", "born"},
	"verificar.gemas.vivas":      {"vivas", "alive"},
	"verificar.gemas.tope":       {"tope", "cap"},
	"verificar.gemas.kings":      {"Kings nacidos (%d) != Kings vivos (%d): las dos fuentes divergen; reportalo", "Kings born (%d) != Kings alive (%d): the two sources diverge; report it"},
	"verificar.gemas.kingsok":    {"Kings: nacidos y vivos coinciden (%d)", "Kings: born and alive match (%d)"},
	"verificar.gemas.masvivas":   {"hay mas gemas vivas que nacidas en algun nivel: imposible; reportalo", "more gems alive than born at some level: impossible; report it"},
	"historial.titulo":           {"== Historial: lo que tu wallet tiene hoy (%d direcciones con saldo) · DAA actual %s", "== History: what your wallet holds today (%d addresses with balance) · current DAA %s"},
	"historial.vacio":            {"nada todavia: la wallet no tiene Gold ni gemas (o el daemon aun no termina de sincronizar)", "nothing yet: the wallet holds no Gold or gems (or the daemon has not finished syncing)"},
	"historial.minado":           {"Minado (agrupado por dia, fecha estimada UTC):", "Mined (grouped per day, estimated UTC date):"},
	"historial.bloques":          {"bloques", "blocks"},
	"historial.movimientos":      {"Recibido y gemas (fecha estimada, hora local):", "Received and gems (estimated date, local time):"},
	"historial.recibido":         {"recibido", "received"},
	"historial.minado.uno":       {"minado", "mined"},
	"historial.gema":             {"gema", "gem"},
	"historial.mas":              {"... y %d mas (usa --todo para verlas todas)", "... and %d more (use --todo to see them all)"},
	"historial.totales":          {"Totales de lo que tienes: Gold %s · Diamante %d · Platino %d · Rodio %d · Kings %d", "Totals of what you hold: Gold %s · Diamond %d · Platinum %d · Rhodium %d · Kings %d"},
	"historial.nota":             {"Nota: el nodo guarda solo lo que tienes ahora; lo que ya gastaste no aparece (no hay indice de transacciones y la cadena se poda). Las fechas son estimadas a un bloque por segundo desde el DAA; por eso llevan ≈. Para ver una transaccion exacta: explorer.rupix.network.", "Note: the node keeps only what you hold now; what you already spent does not appear (there is no transaction index and the chain is pruned). Dates are estimated at one block per second from the DAA; hence the ≈. To see an exact transaction: explorer.rupix.network."},
	"historial.esperando":        {"Pidiendo al nodo la lista de lo que tienes... con muchos bloques minados puede tardar un par de minutos.", "Asking the node for the list of what you hold... with many mined blocks this can take a couple of minutes."},
	"historial.nodo.noresponde":  {"el nodo en %s no responde (%s); historial necesita el nodo ademas del daemon", "the node at %s is not responding (%s); historial needs the node as well as the daemon"},
	"verificar.resumen":          {"Resultado: %d fallo(s). OK = comprobado; AVISO = no se pudo comprobar o conviene mirar; FALLO = algo esta mal.", "Result: %d failure(s). OK = verified; AVISO = could not verify or worth a look; FALLO = something is wrong."},
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
