# Temporary checkpoints in Rupix

🇪🇸 [Versión en español](./CHECKPOINTS.md)

## What they are

A checkpoint is a known canonical block, H, identified by its hash and its blue score X.
The rule (v0.6.1, written for a DAG): every block with blue score ≥ X + MergeDepth must
have H in its past. Any node that receives an alternative history that doesn't pass
through H rejects it, no matter how much mining work it has. (Up to v0.6.0 the inherited
rule was "at a certain DAA score the valid block is one and only one"; on a DAG that
would have split the network. See the history below.)

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
- **Scope (v0.6.1):** it protected nodes that were already synced. A node syncing from
  scratch receives H from an honest peer as its own pruning point, so the rule works by
  construction; against a hostile peer serving a different pruning point with its own
  proof there was no defense.
- **Since v0.6.2, the fresh node too:** before importing the pruning point a peer serves
  (after `ArePruningPointsInValidChain`), every active checkpoint below that point must be
  in the received pruning-point list or be an ancestor the node knows; otherwise
  `ErrCheckpointMismatch` and the pruning point is not imported
  (`TestNodoNuevoContraPeerHostil`). That is why checkpoints are published **only on
  pruning points**: that way they are in every honest peer's list.
- `CheckpointsExpireDAAScore`: past that DAA score, checkpoints are ignored.
  It is the expiry date, inside consensus, verifiable.
- Empty list = no effect. Mainnet, simnet and devnet have empty lists; testnet has
  checkpoint #1 (below).

Tested on devnet (September 18, 2026, under the earlier exact-DAA rule): with the
correct checkpoint the node accepts and mines on top; with a fake one it rejects and the
chain doesn't advance. The current blue-score rule is proven by `TestCheckpointDAG`
(v0.6.1): a late sibling of H gets in, a history without H is rejected, and the test
fails if the rule is switched off.

## When one is published

A checkpoint is only published on a block that already has enough depth (several
thousand blocks on top) and that external nodes already have. Never on recent blocks.
Each published checkpoint is announced with: network, DAA score, hash, date, and the
node version that includes it. (The table below gives the blue score, which is what the rule uses; the DAA score is noted for reference.)

**History (27-Sep-2026), resolved in v0.6.1:** the inherited rule was by exact DAA; on a DAG it would have split the network (see MEMORIA, "El checkpoint que habría partido la red"). What it said then: only one block at that DAA. Rupix is a DAG: two sibling blocks can share a DAA score. The code rejects *any* block at the checkpoint's DAA with a different hash, so a checkpoint placed where there are two blocks would invalidate the sibling and everything that includes it, and a new node could not sync. So before publishing we check with the node that there is exactly one block at that DAA. The real fix (H in the past of every block with blue score ≥ X + MergeDepth) shipped in v0.6.1, with `TestCheckpointDAG`.

## What happens when the pruning point moves past the checkpoint

Verified on 30-Sep-2026: the testnet pruning point moved from H (blue score 86,400) to blue score 172,800, and the seed kept H as a chain block and kept applying the rule. Past pruning points are retained by the node even as pruning advances; that is why checkpoints are published only on pruning points: they never vanish from the node that validates them.

## Renewal: there is never a gap between checkpoints

Every checkpoint has an expiry (`CheckpointsExpireDAAScore`). The rule since 30-Sep-2026: **the next checkpoint is published before the previous one expires**, always on a more recent pruning point and with a new expiry, in a node version announced ahead of time. A gap between checkpoints would be an attack window announced in advance; that is why it is not allowed. If for some reason no new checkpoint is ready before expiry, a version that only extends the current one's expiry is published. The network is left without the lock only when it is removed on purpose (next section), never by neglect. For #1 (expires at DAA 2,000,000): #2 is published no later than DAA 1,700,000 (~3.5 testnet days before).

**How the next one is prepared (since 6-Oct-2026):** `tools/checkpoint-propuesto.sh [pruningPointHash of an external node]` reads the node's current pruning point, checks depth (≥ 100,000 blocks on top) and whether the external node has the same pruning point, and prints the exact lines for `checkpoints.go`, `checkpoints_test.go` and this table. It changes nothing by itself. The expiry is **a single value for the whole list** (`CheckpointsExpireDAAScore`): publishing a new checkpoint moves it for the whole list, so earlier checkpoints stay in force until the new expiry. Nothing is lost: every earlier checkpoint is in the past of the next one, and a node syncing from scratch checks all of them (v0.6.2). If no external node is online to compare against at publication time, the checkpoint is published anyway (a gap is worse) and the table says nobody external confirmed it.

## When they are removed

Checkpoints are removed when the network can stand on its own: sustained external
hashrate, several independent nodes, and weeks of stability. The removal is done by
setting `CheckpointsExpireDAAScore` in consensus and announcing it. There is no fixed
date: there are public conditions.

## Register of published checkpoints

| Network | Blue score (and DAA) | Hash | Date | Version |
|---|---|---|---|---|
| testnet #4 | blue score 86,400 (DAA 86,399) | `7e2ece393c7d991c86e7ba915276cd85b5fc19e8647d5d197fa26bf116604fad` | 29-Sep-2026 | v0.6.1 · expires at DAA 2,000,000 |

*Don't trust, verify.*
