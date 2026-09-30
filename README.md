# Rupix

[![Tests](https://github.com/rupixnet/rupixd/actions/workflows/tests.yaml/badge.svg)](https://github.com/rupixnet/rupixd/actions/workflows/tests.yaml)

🇲🇽 [Versión en español](./README.es.md)

**Rupix is digital money that no one controls, for everyone: no owner, no premine (no coins set aside for its creator), no permission needed to join. With a 42-million cap that no one can change, and a supply that only goes down. While ordinary money gets printed, Rupix gets scarcer. And you don't have to trust anyone: verify it.**

[rupix.network](https://rupix.network) | [@RupixNetwork](https://x.com/RupixNetwork) | [Changelog](./CHANGELOG.md) | [Thanks](./THANKS.md)

---

## Current state (Rupix v0.6.1)

- ✅ **Own mining algorithm — RupixHeavyHash**: a variant of kHeavyHash (Kaspa's algorithm; Rupix is a fork of kaspad under the ISC license, with gratitude). Rupix keeps Kaspa's proven engine (64x64 matrix, HeavyHash) and replaces the generator that fills the matrix (xoshiro256++) with its own, using a structurally different formula (a non-linear multiplication that xoshiro does not have) and a "RUPIX" seal in the seed. Effect: ASICs built for Kaspa cannot mine Rupix — their hardware produces the wrong matrix and the network rejects it. A fair start: minable with GPU/CPU, no inherited hardware advantage. Tested on devnet (22,000+ blocks, 0 rejections, commitment and economy intact) and on the public testnet. The miner (rupixminer, included in every release) uses the same internal function as the node, so it mines with RupixHeavyHash: there are no two algorithms — miner and validator share a single source. Kaspa's engine, Rupix's seed.

**Honest scope:** the custom generator was analyzed independently — it is linear over GF(2), bijective (transition matrix of rank 256, no transient states) and showed no short cycles in testing. Catastrophic collapse was ruled out; the maximum theoretical period (2^256−1) was **not** certified, which would require verifying the primitivity of the characteristic polynomial. And on ASICs: RupixHeavyHash locks out the fixed hardware built for Kaspa, which gives **months of head start, not permanent independence** — an FPGA can be reprogrammed in weeks. It is a fair start so everyone begins on equal footing, not an eternal barrier.

The **public testnet is live**: it accepts external nodes, mines on its own genesis with zero premine, and the full economy lives in consensus. Anyone can connect a node — see [TESTNET-GUIDE.md](./TESTNET-GUIDE.md).

- ✅ **Public testnet running 24/7** — own genesis, zero subsidy, self-adjusting difficulty, listening for connections (public seed)
- ✅ **Downloadable binaries** for Windows, macOS and Linux — see [Releases](https://github.com/rupixnet/rupixd/releases/latest) — always use the latest version
- ✅ **Live public explorer** — [explorer.rupix.network](https://explorer.rupix.network)
- ✅ **Full economy in consensus**: 5-level ladder, 10:1 burn, per-transaction burn, historical caps (2.1M/210k/21k/2,100) — 25+ attack scenarios covered by tests
- ✅ **Full identity**: `rupix:`/`rupixtest:` addresses, `rpub`/`rtub` extended keys, RPC in rupias
- ✅ **Total verification of the gem count (commitment in header)**: the count (Diamond/Platinum/Rhodium/Kings) is sealed into the hash of every block, protected by mining (PoW). The network recomputes and validates the seal on receiving each block, and persists it to disk. A false count does NOT pass: the seal doesn't match and it is rejected. Verifiable from genesis, trusting no one.
- ✅ **Forging working end to end**: mine Gold → burn it → forge a gem (Diamond) → the commitment reflects the real count. Tested live: the forge is mined, the block seal matches validation, the state persists. The ladder lives.
- ✅ **Real Diamond sealed on-chain** (14-Sep-2026): with the testnet past 100,000 blocks (halving 1, Diamond unlocked), a real Diamond was forged. The chain's commitment reflects the count (seal `780e9027…`), with no discrepancy between miner and validator. Code reviewed in two rounds of external audit (verification gaps closed) and with a regression test in CI.
- ✅ **First external forger — real community** (14-Sep-2026): a second user, from their own computer and their own node, forged 3 real Diamonds. They mined Gold, burned it to forge, and the network sealed the count in the commitment. Total verification works between several people, not only the creator.
- ✅ **First Platinum of the network** (14-Sep-2026): 10 Diamonds burned forever, 1 Platinum born. The second step of the ladder, tested on the real chain.
- ✅ **Third external node** (16-Sep-2026): a third participant synced their node from scratch with the RupixHeavyHash testnet (v0.5.0) and received RUPIX. The public network runs on 3 nodes: the seed server and two external ones.
- 🔄 **Testnet #4** (26-Sep-2026): relaunched with v0.6.0, a consensus change. The 14–16 Sep milestones above happened on previous testnets; on #4 the Diamond reopens at DAA 100,000.
- ✅ **First community miner on v0.6.0** (26-Sep-2026): from a Windows PC (~50 KH/s), a community member mined ~40% of the day's blocks — 14,000+ accepted by the network. Checked from the seed, not just from their wallet.
- 💎 **First v0.6.0 Diamond on the public network** (27-Sep-2026): forged by that same community miner, burning 10 Gold he mined himself. Tx `8653650f…3a4b`, DAA 180,710. Checked from the seed.
- ✅ **Live tests with a community node** (28-Sep-2026): same pruning point on both nodes; a Diamond moved between two different wallets and back; a Platinum forged before its unlock was rejected by the community member's own node (`nivel 2 bloqueado`). Checked from the seed.
- ✅ **Halving 2 and first v0.6.0 Platinum** (28-Sep-2026): at DAA 256,208 the issued Gold matched the calendar rule to within 0.5 RUPIX (82,025.5 minus what was burned; the extra 0.5 was explained on 29-Sep: blocks mined before the halving and paid after keep their old reward); Platinum forged after its unlock (tx `c382e0a7…`, DAA 256,577) and accepted, the night after an early one was rejected. Anyone can check issuance against their own node with `tools/verificar-emision.py`.
- ✅ **Halving 3, 10,000 RUPIX in 46 s, and the first community Platinum** (29-Sep-2026): reward dropped to 0.0625 at DAA 300,000 with the same pruning point on both nodes. The fixed wallet sent 1,000 RUPIX (24 txs, 8.6 s) and 10,000 RUPIX (232 txs, 46.1 s) from a ~78k-piece miner wallet — 1,000 used to time out. A community miner forged a Platinum from his Windows PC (tx `bc1bb97590ef7d81…`, DAA 348,162); seed and his node, queried independently, returned the same three gems with the same IDs and birth blocks.
- ✅ **v0.6.1 and the first published checkpoint** (29-Sep-2026): the external auditor approved the four branches (DAG checkpoints, mempool level test, linear UTXO selection, wallet keys UX); merged and released. Testnet checkpoint #1 on the pruning point both nodes already shared: blue score 86,400, `7e2ece393c7d991c…`. The 0.5 RUPIX over the calendar at halving 2 was explained and documented (see MEMORIA).
- ✅ **First green CI and reproducible builds verified** (30-Sep-2026): the inherited workflow had been red for weeks; replaced with what is ours (gofmt, vet, staticcheck, build, full test suite) — green on Linux and macOS. The v0.6.1 release binaries were rebuilt from the tag on the seed with the CI's flags: 4 of 4 SHA256 identical. `tools/verificar-binarios.sh` lets anyone repeat it.
- ✅ **Halving 4: the whole ladder is open** (30-Sep-2026): testnet #4 crossed DAA 400,000 at noon; reward 0.03125, era 5, all five levels open (Kings included). Issuance checked at DAA 407,293: 219.99 below the calendar = 220 Gold burned in 22 Diamonds. The pruning point moved past checkpoint #1 and the seed still holds H as a chain block: checkpoints on pruning points don't vanish.
- ✅ **Temporary checkpoints** (18-Sep-2026): defense against the 51% attack while hashrate is low. A block at a checkpoint's DAA score must have the canonical hash or it is rejected (`ErrCheckpointMismatch`). With expiry inside consensus (`CheckpointsExpireDAAScore`). Tested on devnet: a correct checkpoint accepts and mines on; a false checkpoint rejects and the chain stops. Empty list today: the first one will be published with an announcement. Full policy in [CHECKPOINTS.en.md](./CHECKPOINTS.en.md). Temporary, declared, not hidden centralization.
- ✅ **Verifiable binaries with SHA256**: every release publishes each binary's fingerprint, generated by CI — download, compare, and confirm no one altered it
- ✅ **Consensus bug found by the new end-to-end King test — fixed** (21-Sep-2026): the miner double-counted forges in the block it was building (a 13-Sep patch), while the validator counts what the block *accepts*. Any production-mined block containing a forge would have been rejected — the first real King on mainnet included. Fixed in **v0.6.0** (in production since 26-Sep), together with the keccak domain and integer rank; `TestKingsEndToEnd` fails if the fix is reverted. Found on testnet, not mainnet.
- ✅ **`go test ./...` green and `go vet` clean** (20-Sep-2026): the full suite passes. The last 18 red packages were not consensus bugs: they were tests inherited from Kaspa with Kaspa expectations (`kaspa:` prefixes, 500 reward, hash tie-break order) and outputs created with `Version = MaxScriptPublicKeyVersion`, which in Rupix is 4 = Kings — the ladder rejected them as fake Kings. Regenerated from the code, verifiable.
- ✅ **0 vulnerabilities** (govulncheck), built with Go 1.25+
- ✅ **A network of more than one node**: first external node connected and synced, first transaction between two people recorded on-chain

## What Rupix is

Rupix is a Layer 1 blockchain with Proof of Work consensus over a BlockDAG (not a linear chain). **Rupix is a fork of kaspad**: an independent chain, with its own coin, built from Kaspa's open-source code (GHOSTDAG, kHeavyHash) under the ISC license. **Rupix is not part of the Kaspa network and does not use KAS.** We acknowledge and thank that work: without the code the Kaspa team published, Rupix would not exist.

What Rupix adds on top:

- **Its own 5-level economic model** with permanent burning to forge each higher level
- **A supply of 42,000,000 RUPIX by schedule**, sealed in the protocol: 41,999,994.96 by calendar (integer truncation), plus a bounded excess at each halving boundary (blocks mined just before a halving and paid after keep the old reward — measured: 0.5 RUPIX at testnet halving 2). A few RUPIX over the life of the chain; `MaxRupia` is a per-transaction cap, not an emission cap. Said here before you compute it.
- **Per-transaction burn**: every transfer destroys rupias forever
- **Genesis with no premine**: the first RUPIX was mined after block 0, like Bitcoin
- **Unlocking by halvings**: each level of the ladder opens with a halving — scarcity has a calendar

## The ladder — the 5 levels

| Level | Name | Max supply | How it's forged | Unlocks |
|-------|------|------------|-----------------|---------|
| L1 | Gold | 42,000,000 | Mining | From genesis |
| L2 | Diamond | 2,100,000 | Burn 10 Gold | Halving 1 |
| L3 | Platinum | 210,000 | Burn 10 Diamond | Halving 2 |
| L4 | Rhodium | 21,000 | Burn 10 Platinum | Halving 3 |
| L5 | Kings | 2,100 | Burn 10 Rhodium | Halving 4 |

Each level is forged by burning 10 units of the previous one. Creating 1 Kings means 10,000 Gold have been destroyed along the chain; filling all 2,100 Kings would destroy 21 million Gold — half of all that will ever exist. The burn is irreversible and stays on-chain forever. No one can reverse it: not the creator, not the miners, not any future agreement.

## The fire — deflation with every use

Every Rupix transaction destroys a small amount, required by consensus:

```
burn = 1,000 rupias + (tx_bytes × 10 rupias)
```

Where 1 RUPIX = 100,000,000 rupias. These rupias don't go to a fund or to the miner: **they disappear**, in an OpReturn output visible forever and impossible to spend. This already works: the network's first transaction paid it.

## Why we believe in the trilemma

The blockchain trilemma states that any distributed network must choose between decentralization, security and scalability, and can only have two at once.

Rupix is built on the premise that a BlockDAG with Proof of Work allows pushing all three further at the same time than traditional architectures do. We do not claim the trilemma is solved: we say we are pushing it in a direction that respects all three principles.

- **Decentralization**: PoW with no premine, open source, no centralized governance, anyone-can-mine
- **Security**: full cryptographic validation, no shortcuts, no trusted parties
- **Scalability**: BlockDAG allows multiple parallel blocks without losing consistency

## Verify it yourself

- **Rule by rule**: [ESPECIFICACION.md](./ESPECIFICACION.md) lists every consensus rule next to the attack it stops and the test that violates it and confirms the rejection. Where the third column is empty, that is the open work (listed at the end). Spanish for now; English version pending.
- **That the published binaries are exactly what the code produces**: `sh tools/verificar-binarios.sh v0.6.1` builds the tag with the CI's flags and compares SHA256 binary by binary. Done on 30-Sep-2026 for v0.6.1: 4 of 4 identical (`rupixd` `ea973ce0d720678a…`). Reproducible builds: you don't have to trust GitHub, or us.
- **That the CI is green for real**: the badge above runs gofmt, go vet, staticcheck, build and `go test ./...` on Linux and macOS on every push. Green since 30-Sep-2026; before that, the inherited workflow was red for weeks over pieces that weren't ours (see MEMORIA).

Don't trust us. Check it:

- **That there is no premine**: `go run ./cmd/genesisgen` regenerates the genesis and shows the subsidy at zero, byte by byte
- **That the cap is 42M**: check `domain/consensus/utils/constants/constants.go` (MaxRupia)
- **That the economy has tests**: `go test ./domain/...` — the ladder, the burn and the Kings with their attack scenarios
- **That PoW cannot be disabled**: `go test ./domain/dagconfig/...`
- **The live supply** (with a running node): `rupixctl GetCoinSupply` → `maxRupias: 4200000000000000` — the 42,000,000 RUPIX per-transaction cap; `circulatingRupias` is what exists today

## How to run a node

Requirements: Go 1.25+, 4 GB RAM, 50 GB disk.

```
git clone https://github.com/rupixnet/rupixd.git
cd rupixd
go build -o rupixd .
go build -o rupixminer ./cmd/rupixminer
go build -o rupixwallet ./cmd/rupixwallet
go build -o rupixctl ./cmd/rupixctl
```

Connect to the testnet:

```
./rupixd --testnet --utxoindex
```

*(The public testnet is live 24/7 with an open seed: `--addpeer=178.104.69.148:17211`. Anyone can connect — see [TESTNET-GUIDE.md](./TESTNET-GUIDE.md).)*

## How to mine Rupix

Rupix is mined with an ordinary computer (GPU or CPU) — ASICs built for Kaspa don't work. The `rupixminer` miner ships in every [release](https://github.com/rupixnet/rupixd/releases/latest) for Linux, Windows and macOS.

1. Create an address to receive your mined coins:
   `./rupixwallet --testnet create`
   `./rupixwallet --testnet start-daemon --keys-file=<your-keys.json>`
   `./rupixwallet --testnet new-address`

2. With your node running (see above), start the miner pointing at it:
   `./rupixminer --testnet --rpcserver=127.0.0.1:17210 --miningaddr=<your-address>`

Every block you find pays you Gold. With that Gold you can forge gems (see [FORGE-GUIDE.md](./FORGE-GUIDE.md)). Mining on testnet has no monetary value: it helps the network grow its hashrate on the way to mainnet. The more honest, distributed hashrate, the more secure the network — mining is the most direct way to contribute to Rupix.

## Road to mainnet

- ✅ **Mining accessible to everyone — DONE (v0.6.0)** — Rupix migrated from Kaspa's inherited algorithm to RupixHeavyHash, its own algorithm. Kaspa ASICs can no longer mine Rupix; it is mined from an ordinary computer (GPU/CPU). Rupix is for everyone.
- ✅ **Total verification with commitment in header** — the gem count is sealed into every block's hash (protected by PoW), validated on receipt, and persisted to disk. A false count is rejected: the seal doesn't match. This CLOSES total verifiability — tested live (first transfer between nodes, testnet v0.4.2).
- ✅ **Temporary checkpoints — DONE (v0.6.0)** — with expiry inside consensus, tested on devnet, public policy in [CHECKPOINTS.en.md](./CHECKPOINTS.en.md). The first real one is yet to be published.
- **Go vs Rust — declared risk.** Rupix runs on kaspad-go, the legacy implementation; Kaspa's active development is in rusty-kaspa. kaspad-go does everything Rupix needs today (GHOSTDAG, pruning, kHeavyHash) but receives no upstream improvements or fixes. A migration to rusty-kaspa is a second-year goal, conditional on having contributors who can sustain it. Not a promise; a stated direction.
- **External audit of the consensus code**
- **Redundant infrastructure** (multiple seed nodes) and **committed hashrate**

**Mainnet date: we will announce it when the code is ready, not before.** We would rather launch late and right than early and compromised.

## Philosophy

Rupix is not a fork for novelty or hype. It is a new economic architecture on a proven consensus engine. We took what was already well made (the GHOSTDAG engine) and built on top of it an original economic proposal that bets on verifiable scarcity and radical honesty.

Whoever created Rupix mines from block 1, like anyone else. There are no privileged addresses, no sales, no rounds. The only advantage of arriving early is having been awake when the network started.

## License

ISC — Rupix developers, 2026.

## Acknowledgments

To the researchers and developers who created and published GHOSTDAG under an open license. Their work makes projects like Rupix possible.

**Don't trust, verify.** ER
