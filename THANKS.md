# Gracias / Thanks

🇲🇽 Rupix se hace mejor cada vez que alguien lo cuestiona, lo prueba o lo corrige. Este archivo registra a quienes lo hicieron, con fecha y con lo que cambió. **Mencionar a alguien aquí no significa que apoye a Rupix**: significa que nos señaló algo y tenía razón, o que fue el primero en probar algo. Alias públicos solo cuando la aportación fue pública; nombres reales solo con permiso.

🇬🇧 Rupix gets better every time someone questions it, tests it, or corrects it. This file records those who did, with the date and what changed. **Being mentioned here does not mean endorsing Rupix**: it means they pointed something out and were right, or were the first to try something. Public aliases only when the contribution was public; real names only with permission.

---

## Revisión externa / External review

| Fecha / Date | Quién / Who | Qué señaló / What they pointed out | Qué cambió / What changed |
|---|---|---|---|
| 29-ago → 20-sep-2026 | **El Auditor** (anónimo por decisión propia / anonymous by choice) | 5 rondas, 10 hallazgos (H-1..H-10), 2 regresiones cazadas, un documento de traspaso para auditoría profesional. Predijo que H-5 (campo `Version` sobrecargado) volvería a aparecer / 5 rounds, 10 findings, 2 regressions caught, a handoff document for professional audit. Predicted H-5 would resurface | 9 de 10 cerrados. H-5 resultó ser la causa raíz de 18 tests rojos, meses después / 9 of 10 closed. H-5 turned out to be the root cause of 18 red tests, months later |
| 19-sep-2026 | **supertypo** y otro miembro / and another member, Discord de Kaspa | "Esto no está construido sobre Kaspa; es un fork, va en #off-topic" / "This isn't built on Kaspa; it's a fork, belongs in #off-topic" | "Fork de kaspad, cadena independiente, no usa KAS" en README (es/en), guías, web. Post movido / wording fixed everywhere, post moved |
| 20-sep-2026 | **FreshAir08**, Discord de Kaspa | Commitment de gemas sobrevendido; zero premine no es distinto de Kaspa; README solo en español; Go en vez de Rust / gems commitment overstated; zero premine not distinct from Kaspa; README only in Spanish; Go instead of Rust | Wording corregido ("defensa en profundidad, no único"); README.en.md; Go/Rust al roadmap como riesgo declarado / wording fixed; English README; Go/Rust in the roadmap as a declared risk |

## Primeros nodos / First nodes

| Fecha / Date | Quién / Who | Qué hizo / What they did |
|---|---|---|
| 4-sep-2026 | **JC** | Primer nodo externo: descargó v0.4.0, sincronizó desde cero, minó. Primera transacción entre dos personas en la cadena / First external node: downloaded v0.4.0, synced from scratch, mined. First transaction between two people on-chain |
| 14-sep-2026 | **JC** | **Primer forjador externo**: forjó 3 Diamantes reales desde su propio nodo (tx `adbfe36b…`, commitment `6c8f913d…`). Encontró que la wallet pedía flags distintos a la guía → guía corregida / **First external forger**: 3 real Diamonds from his own node. Found the wallet flags differed from the guide → guide fixed |
| 16-sep-2026 | **JP** | Tercer nodo: sincronizó desde cero con RupixHeavyHash (v0.5.0) el mismo día del relanzamiento. Su binario reportaba 0.4.0 → se descubrió que el release se compiló antes del fix de versión → release regenerado / Third node: synced from scratch with RupixHeavyHash the day of the relaunch. His binary reported 0.4.0 → revealed the release was built before the version fix → release regenerated |
| 17-sep-2026 | **JP** | Su wallet en Windows reveló el orden correcto de flags (`--keys-file` después del subcomando) y el caso del daemon cerrado → guía corregida / His Windows wallet revealed the correct flag order and the closed-daemon case → guide fixed |

*Sin ustedes, Rupix seguiría siendo un experimento de una persona. / Without you, Rupix would still be a one-person experiment.*

*No confíes, verifica. / Don't trust, verify.*
