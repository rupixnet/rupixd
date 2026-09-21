# Rupix

🇲🇽 [Versión en español](./README.es.md)

**Rupix is a scarce digital asset for everyone: no owner, no premine (no coins set aside for its creator), no permission needed to join. With a 42-million cap that no one can change, and a supply that only goes down. While ordinary money gets printed, Rupix gets scarcer. And you don't have to trust anyone: verify it.**

[rupix.network](https://rupix.network) | [@RupixNetwork](https://x.com/RupixNetwork) | [Changelog](./CHANGELOG.md) | [Thanks](./THANKS.md)

---

## Current state (Rupix v0.5.2)

- ✅ **Own mining algorithm — RupixHeavyHash**: a variant of kHeavyHash (Kaspa's algorithm; Rupix is a fork of kaspad under the ISC license, with gratitude). Rupix keeps Kaspa's proven engine (64x64 matrix, HeavyHash) and replaces the generator that fills the matrix (xoshiro256++) with its own, using a structurally different formula (a non-linear multiplication that xoshiro does not have) and a "RUPIX" seal in the seed. Effect: ASICs built for Kaspa cannot mine Rupix — their hardware produces the wrong matrix and the network rejects it. A fair start: minable with GPU/CPU, no inherited hardware advantage. Tested on devnet (22,000+ blocks, 0 rejections, commitment and economy intact) and on the public testnet. The miner (rupixminer, included in every release) uses the same internal function as the node, so it mines with RupixHeavyHash: there are no two algorithms — miner and validator share a single source. Kaspa's engine, Rupix's seed.

**Honest scope:** the custom generator was analyzed independently — it is linear over GF(2), bijective (transition matrix of rank 256, no transient states) and showed no short cycles in testing. Catastrophic collapse was ruled out; the maximum theoretical period (2^256−1) was **not** certified, which would require verifying the primitivity of the characteristic polynomial. And on ASICs: RupixHeavyHash locks out the fixed hardware built for Kaspa, which gives **months of head start, not permanent independence** — an FPGA can be reprogrammed in weeks. It is a fair start so everyone begins on equal footing, not an eternal barrier.

The **public testnet is live**: it accepts external nodes, mines on its own genesis with zero premine, and the full economy lives in consensus. Anyone can connect a node — see [GUIA-TESTNET.md](./GUIA-TESTNET.md).

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
- ✅ **Third external node** (16-Sep-2026): a third participant synced their node from scratch with the RupixHeavyHash testnet (v0.5.2) and received RUPIX. The public network runs on 3 nodes: the seed server and two external ones.
- ✅ **Temporary checkpoints** (18-Sep-2026): defense against the 51% attack while hashrate is low. A block at a checkpoint's DAA score must have the canonical hash or it is rejected (`ErrCheckpointMismatch`). With expiry inside consensus (`CheckpointsExpireDAAScore`). Tested on devnet: a correct checkpoint accepts and mines on; a false checkpoint rejects and the chain stops. Empty list today: the first one will be published with an announcement. Full policy in [CHECKPOINTS.md](./CHECKPOINTS.md). Temporary, declared, not hidden centralization.
- ✅ **Verifiable binaries with SHA256**: every release publishes each binary's fingerprint, generated by CI — download, compare, and confirm no one altered it
- ✅ **`go test ./...` green and `go vet` clean** (20-Sep-2026): the full suite passes. The last 18 red packages were not consensus bugs: they were tests inherited from Kaspa with Kaspa expectations (`kaspa:` prefixes, 500 reward, hash tie-break order) and outputs created with `Version = MaxScriptPublicKeyVersion`, which in Rupix is 4 = Kings — the ladder rejected them as fake Kings. Regenerated from the code, verifiable.
- ✅ **0 vulnerabilities** (govulncheck), built with Go 1.26.6
- ✅ **A network of more than one node**: first external node connected and synced, first transaction between two people recorded on-chain

## What Rupix is

Rupix is a Layer 1 blockchain with Proof of Work consensus over a BlockDAG (not a linear chain). **Rupix is a fork of kaspad**: an independent chain, with its own coin, built from Kaspa's open-source code (GHOSTDAG, kHeavyHash) under the ISC license. **Rupix is not part of the Kaspa network and does not use KAS.** We acknowledge and thank that work: without the code the Kaspa team published, Rupix would not exist.

What Rupix adds on top:

- **Its own 5-level economic model** with permanent burning to forge each higher level
- **An absolute supply of 42,000,000 RUPIX**, sealed in the protocol
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

Don't trust us. Check it:

- **That there is no premine**: `go run ./cmd/genesisgen` regenerates the genesis and shows the subsidy at zero, byte by byte
- **That the cap is 42M**: check `domain/consensus/utils/constants/constants.go` (MaxRupia)
- **That the economy has tests**: `go test ./domain/...` — the ladder, the burn and the Kings with their attack scenarios
- **That PoW cannot be disabled**: `go test ./domain/dagconfig/...`
- **The live supply** (with a running node): `rupixctl GetCoinSupply` → `maxRupias: 4200000000000000` — exactly 42,000,000 RUPIX

## How to run a node

Requirements: Go 1.21+, 4 GB RAM, 50 GB disk.

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

*(The public testnet is live 24/7 with an open seed: `--addpeer=178.104.69.148:17211`. Anyone can connect — see [GUIA-TESTNET.md](./GUIA-TESTNET.md).)*

## Road to mainnet

- ✅ **Mining accessible to everyone — DONE (v0.5.2)** — Rupix migrated from Kaspa's inherited algorithm to RupixHeavyHash, its own algorithm. Kaspa ASICs can no longer mine Rupix; it is mined from an ordinary computer (GPU/CPU). Rupix is for everyone.
- ✅ **Total verification with commitment in header** — the gem count is sealed into every block's hash (protected by PoW), validated on receipt, and persisted to disk. A false count is rejected: the seal doesn't match. This CLOSES total verifiability — tested live (first transfer between nodes, testnet v0.4.2).
- ✅ **Temporary checkpoints — DONE (v0.5.2)** — with expiry inside consensus, tested on devnet, public policy in [CHECKPOINTS.md](./CHECKPOINTS.md). The first real one is yet to be published.
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
