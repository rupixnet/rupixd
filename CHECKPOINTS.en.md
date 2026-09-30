# Temporary checkpoints in Rupix

🇪🇸 [Versión en español](./CHECKPOINTS.md)

## What they are

A checkpoint is a known canonical block: at a certain height of the network (DAA score),
the valid block is one and only one, identified by its hash. Any node that receives an
alternative chain that doesn't pass through that block rejects it, no matter how much
mining work it has.

## Why Rupix uses them

A young network has little hashrate. With little hashrate, an attacker with enough
hardware could rewrite history (a 51% attack): undo transactions, double-spend, erase
forges. Checkpoints make that impossible for everything before the latest checkpoint:
the attacker could only affect the stretch after it, and that stretch is short if
checkpoints are published regularly.

It is a **temporary** defense and a form of **declared centralization**: whoever
publishes the checkpoint decides which history is canonical. Rupix does it in the open,
with the rule written down, and with a commitment to remove it.

## How it works (in the code)

- Each network has a `Checkpoints` list (`domain/dagconfig/checkpoints.go`):
  `{BlueScore, Hash}` pairs: the canonical block H and its blue score X.
- **Rule (v0.6.1, DAG formulation):** when validating a header (`checkCheckpoint`
  in `block_header_in_context.go`), every block with blue score ≥ X + MergeDepth (3,600)
  must have H in the past of one of its parents. Otherwise it is rejected with
  `ErrCheckpointMismatch`. A sibling of H is unaffected (its blue score is below the
  threshold); an honest later block always has H in its past; an alternative history
  that skips H is rejected whole. Tested in `TestCheckpointDAG`, and we verified the
  test fails if the rule is switched off.
- A node that already pruned below H (does not have it) does not apply the rule: it can't.
- **Scope today:** it protects nodes that are already synced. A node syncing from
  scratch receives H from an honest peer as its own pruning point, so the rule works by
  construction; against a hostile peer serving a different pruning point with its own
  proof, validating the pruning-point list against the checkpoints lands in v0.6.2
  (after `ArePruningPointsInValidChain`). That is why checkpoints are published **only
  on pruning points**.
- `CheckpointsExpireDAAScore`: past that DAA score, checkpoints are ignored.
  It is the expiry date, inside consensus, verifiable.
- Empty list = no effect. Mainnet, simnet and devnet have empty lists; testnet has
  checkpoint #1 (below).

Tested on devnet (September 18, 2026): with the correct checkpoint the node accepts
and mines on top; with a fake checkpoint, it rejects the block at that DAA score and
the chain doesn't advance.

## When one is published

A checkpoint is only published on a block that already has enough depth (several
thousand blocks on top) and that external nodes already have. Never on recent blocks.
Each published checkpoint is announced with: network, DAA score, hash, date, and the
node version that includes it.

**History (27-Sep-2026), resolved in v0.6.1:** the inherited rule was by exact DAA; on a DAG it would have split the network (see MEMORIA, "El checkpoint que habría partido la red"). What it said then: only one block at that DAA. Rupix is a DAG: two sibling blocks can share a DAA score. The code rejects *any* block at the checkpoint's DAA with a different hash, so a checkpoint placed where there are two blocks would invalidate the sibling and everything that includes it, and a new node could not sync. So before publishing we check with the node that there is exactly one block at that DAA. The real fix (require the hash only from chain blocks) ships in the next version, with a test.

## When they are removed

Checkpoints are removed when the network can stand on its own: sustained external
hashrate, several independent nodes, and weeks of stability. The removal is done by
setting `CheckpointsExpireDAAScore` in consensus and announcing it. There is no fixed
date: there are public conditions.

## Register of published checkpoints

| Network | DAA score | Hash | Date | Version |
|---|---|---|---|---|
| testnet #4 | blue score 86,400 (DAA 86,399) | `7e2ece393c7d991c86e7ba915276cd85b5fc19e8647d5d197fa26bf116604fad` | 29-Sep-2026 | v0.6.1 · expires at DAA 2,000,000 |

*Don't trust, verify.*
