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

- Each network has a `Checkpoints` list in `domain/dagconfig/params.go`:
  `{DAAScore, Hash}` pairs.
- When validating a header (`checkCheckpoint` in `block_header_in_context.go`),
  if the block's DAA score matches a checkpoint and its hash is not the canonical
  one, the block is rejected with `ErrCheckpointMismatch`.
- A block at a DAA score without a checkpoint passes as usual.
- `CheckpointsExpireDAAScore`: past that DAA score, checkpoints are ignored.
  It is the expiry date, inside consensus, verifiable.
- Empty list = no effect. Today every network has an empty list.

Tested on devnet (September 18, 2026): with the correct checkpoint the node accepts
and mines on top; with a fake checkpoint, it rejects the block at that DAA score and
the chain doesn't advance.

## When one is published

A checkpoint is only published on a block that already has enough depth (several
thousand blocks on top) and that external nodes already have. Never on recent blocks.
Each published checkpoint is announced with: network, DAA score, hash, date, and the
node version that includes it.

**Rule (27-Sep-2026): only one block at that DAA.** Rupix is a DAG: two sibling blocks can share a DAA score. The code rejects *any* block at the checkpoint's DAA with a different hash, so a checkpoint placed where there are two blocks would invalidate the sibling and everything that includes it, and a new node could not sync. So before publishing we check with the node that there is exactly one block at that DAA. The real fix (require the hash only from chain blocks) ships in the next version, with a test.

## When they are removed

Checkpoints are removed when the network can stand on its own: sustained external
hashrate, several independent nodes, and weeks of stability. The removal is done by
setting `CheckpointsExpireDAAScore` in consensus and announcing it. There is no fixed
date: there are public conditions.

## Register of published checkpoints

| Network | DAA score | Hash | Date | Version |
|---|---|---|---|---|
| — | — | — | — | (none yet) |

*Don't trust, verify.*
