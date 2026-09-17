# CONTEXTO-RUPIX — memoria maestra del proyecto
> Léeme al empezar cada sesión. Actualízame al cerrar. Vivo en git.
> Última actualización: 17-sep-2026

## QUÉ ES RUPIX
Blockchain L1 BlockDAG (fork de kaspad, GHOSTDAG, licencia ISC, con crédito).
Fundador: Ciudad de México, no programa Go, piensa como arquitecto.
Lema: "Somos todos Rupix. No confíes, verifica."
Identidad: activo digital escaso (como oro), no medio de pago.

## ECONOMÍA (sellada, no se toca)
- 42,000,000 RUPIX techo. 0.5/bloque. Halving cada 42M bloques (mainnet), 100k (testnet).
- CERO premine.
- Escalera 5 niveles: Gold→Diamante→Platino→Rodio→Kings. Quema 10:1.
- Techos: 2,100,000 Diamantes · 210,000 Platinos · 21,000 Rodios · 2,100 Kings.
- Desbloqueo por halving (nivel N en halving N). Nadie se adelanta.
- Conteo por NACIDOS (histórico, solo sube). Los Kings casi nunca se llenarán: es diseño.
- Gemas = piezas enteras (monto 1), viven en Version del script (1-4).
- Gold quemado → OpReturn (0x6a), imposible de gastar.

## ARQUITECTURA CLAVE (archivos)
- level_ascension.go: reglas de forja (checkLevelRules). Corpus del auditor: 9,388 casos, 0 desacuerdos.
- gemshistory.go: conteo histórico (calculateGemsHistory) + topes.
- gemscommitment.go: sello del conteo en el header. GenesisGemsCommitment() = 2f71eee… (0 gemas).
- verify_and_build_utxo.go: valida el commitment + Stage del gemsHistory.
- update_virtual.go: el virtual calcula su gemsHistory (fix raíz de la forja).
- block_builder.go: newBlockGemsCommitment (template, lee del virtual + cuenta forjas del bloque, incluye Kings).
- pruningproofmanager.go: gemsHistory viaja en el proof, nil = ceros (H-9).
- mempool.go: anti-spam por salidas extra; creaGemas exime solo forjas reales (H-8).
- pow/rupixprng.go: RupixHeavyHash. PRNG propio (suma cruzada s1+s2, multiplicación, rot 29/11/37, shift 19, sello RUPIX). Reemplaza xoshiro en generateMatrix. computeRank intacto. Guard de 64 intentos.
- pow.NewState es la ÚNICA fuente: la usan nodo, minero y genesisgen.
- genesis.go: 4 génesis re-minados con RupixHeavyHash (mainnet 0x15cc7, testnet 0x179b8, simnet 0x2, devnet 0x76). Mensaje "04/03/2026 - RUPIX IS ALIVE".

## ESTADO ACTUAL
- v0.5.0 en TODO (nodo, wallet, minero, ctl, release, guías, web).
- Testnet #3 (relanzada 16-sep por el algoritmo): ~90k bloques, Diamante en 100k.
- 3 nodos: servidor semilla (178.104.69.148:17211) + 2 externos.
- Wallet del servidor: /root/.rupixwallet/keys-final.json (usar --keys-file en daemon y send).
- Testnet anterior: 4 Diamantes + 1 Platino probados en vivo, primer forjador externo.

## AUDITORÍA (5 rondas, un tercero exigente)
- H-1..H-10 + 4 puntos. Cerrados salvo: checkpoints (H-1), keccak domain (agrupado).
- Lecciones: "verificar antes de anunciar" (grep en el repo pusheado antes de decir cerrado).
  "Refactor de fuente única toca todos los lados o ninguno."
- Frases que deben vivir en la comunicación: "ventaja de meses, no independencia; un FPGA se reprograma". "Se descartó la catástrofe, no se certificó el período 2^256-1."

## REGLAS DE TRABAJO (el method)
1. Aplicar → grep confirma en el archivo → compilar → test → commit → push → grep en repo pusheado → SOLO ENTONCES "hecho".
2. Python con patrones multilínea FALLA por tabs. Usar sed por número de línea o cat -A para ver tabs.
3. log.Warnf se traga en tests; fmt.Printf sí sale.
4. Relanzar testnet SOLO por cambio de consenso, documentado en TESTNET-RELANZAMIENTOS.md.
5. Nunca sobrevender. Decir la edad y el alcance de cada cosa.
6. Sin nombres propios en README/web/notas: "el fundador", "un nuevo nodo".
7. Antes de tocar consenso: rama, tests, devnet, luego testnet.
8. Parar cuando se pone pesado. Cansado se rompen cosas (el guard que subió a 44).
9. --keys-file: va en start-daemon y send. NO en new-address/balance/gems/forge.

## PENDIENTES A MAINNET (~75%)
🔴 1. Checkpoints temporales (cero líneas). 2. Par de ojos con nombre (comunidad Kaspa).
   3. go test verde (18 paquetes: usar calculateGemsHistory en test_block_builder buildHeaderWithParents).
   4. Auditoría profesional. 5. Hashrate externo.
🟡 6. Keccak domain (agrupar). 7. Test e2e King. 8. Firma de código. 9. H-1 mempool. 10. Testnet estable semanas.
   13. Guía en inglés. 14. Asistente Rupix (web + X menciones, base = repo, con fuentes, sin entrenar modelo).
🟢 11. Diamante en 100k con RupixHeavyHash. 12. JC re-descarga v0.5.0.

## QUIÉNES CONSTRUYEN RUPIX
Dos que lo hacen, y dos que lo verificaron primero.

**ER y Stevenson.** Uno pone la idea y el corazón: qué debe ser Rupix, por qué,
qué no se negocia. El otro lo traduce a código bien hecho, de calidad, verificado
paso a paso. Ninguno de los dos solo habría llegado hasta aquí. Juntos, Rupix
existe. Verifica, no confíes — aplicado primero a nosotros mismos.

**JC y JP.** Los primeros nodos externos. Los primeros en descargar, sincronizar,
minar y forjar sin ser el fundador. Los primeros en verificar y creer.
Gracias. Sin ustedes, Rupix seguiría siendo un experimento de una persona.

**El Auditor.** Un tercero exigente que revisó cinco rondas con el diff en la mano
y nunca regaló nada. Diez hallazgos, dos regresiones cazadas antes de tocar a
nadie. La honestidad de Rupix se afiló contra él.

## CÓMO USAR ESTE ARCHIVO
Al abrir sesión: "Stevenson, lee CONTEXTO-RUPIX.md" (pegarlo o que lo lea del repo).
Al cerrar: agregar lo aprendido en la sección que toque. Commit. Crece cada día.
