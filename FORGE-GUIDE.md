# 💎 How to forge your first gem on Rupix

🇪🇸 [Versión en español](./GUIA-FORJA.md)

Guide to create your first Diamond by burning Gold. This is the
essence of Rupix: destroy to create something scarcer.

> **Requirement:** you need **Gold** (RUPIX) in your wallet.
> Get it by mining (see the [testnet guide](./TESTNET-GUIDE.md)) or ask someone
> to send you some. To forge 1 Diamond you need **at least 10 Gold**
> (10 are burned) plus a little extra for the transaction fee.

> **⚠️ Important — keep some Gold aside:** moving or transferring a gem
> also costs a small fee in Gold. If you forge a gem and are left with NO Gold,
> your gem stays put: you won't be able to move or transfer it until you have
> some Gold again. **You didn't lose it** — it's still yours and visible on-chain —
> it just needs some Gold to move. Tip: always keep a little Gold aside.

> **⚠️ Windows vs Linux/Mac:** the commands below start with `./` (Linux/Mac).
> In **Windows PowerShell**, use `.\` instead of `./`, for example: `.\rupixd.exe`.
> Without `./` or `.\`, the system can't find the program. (Thanks to JC for the tip!)

## Step 1 — Have your wallet and daemon running

If you don't have a wallet yet, create one (only once):
```
./rupixwallet --testnet create
```
Start the wallet daemon (leave it running in its own window):
```
./rupixwallet --testnet start-daemon
```

---

## Step 2 — Check your Gold

In another window, see how much Gold you have:
```
./rupixwallet --testnet balance
```
You need at least ~10.5 RUPIX to forge a Diamond
(10 are burned, plus a little for the fee).

---

## Step 3 — Create an address for your gem

The gem is "born" at one of your addresses. Create one (or use one you already have):
```
./rupixwallet --testnet new-address
```
Copy the address that starts with `rupixtest:...`

---

## Step 4 — Forge your Diamond!

> **Before forging:** each level opens at its halving. On testnet: Diamond at DAA 100,000,
> Platinum at 200,000, Rhodium at 300,000 and Kings at 400,000. Before that point the network rejects the forge.
> Check the network's DAA with `./rupixctl --testnet GetBlockDagInfo` (`virtualDaaScore` field)
> and leave a few blocks of margin.
>
> `--level`: 1 = Diamond, 2 = Platinum, 3 = Rhodium, 4 = Kings. Gold isn't forged: it's mined.
>
> Since v0.6.1 the wallet prompts for the password on screen: don't type it on the
> command line (it would end up in your history). Still on v0.6.0? Update before forging.

```
./rupixwallet --testnet forge --level=1 --gem-address=YOUR_ADDRESS
Password: (type it here, it is not shown)
```

- `--level=1` → Diamond (the first gem level)
- `--gem-address=` → the address where the gem is born (from step 3)
- The password is asked at that moment; it never goes in the command

This **burns 10 Gold forever** and creates **1 Diamond**. The burn is
recorded on the blockchain, visible to everyone, irreversible.

The levels (for later):
- `--level=1` → Diamond (burns 10 Gold)
- `--level=2` → Platinum (burns 10 Diamonds)
- `--level=3` → Rhodium (burns 10 Platinums)
- `--level=4` → Kings (burns 10 Rhodiums)

---

## Step 5 — See your gem

```
./rupixwallet --testnet gems
```
You'll see your inventory. There's your Diamond! You're one of the first
to forge a gem on Rupix. 💎

---

## What did you just do?

You created a **cryptographic proof that you burned real Gold**.
Your Diamond proves you burned 10 Gold, forever. It's not an image or a
made-up number: it's verifiable scarcity, recorded on-chain. Only
2,100,000 Diamonds will ever exist.

Each higher level is 10 times rarer. A King (level 4) requires burning
10,000 Gold in total. Only 2,100 Kings will exist in all of Rupix's history.

---

*Rupix — money that only gets consumed. Don't trust. Verify — from genesis.*
🔗 rupix.network · github.com/rupixnet/rupixd · explorer.rupix.network

## What changed in the wallet (v0.6.2) and why

These changes exist because we watched real people trip over the wallet. Each one prevents a concrete mistake.

| Before | Now | Why |
|---|---|---|
| `forge` sent the transaction and waited for the network to reject it | **Checks first:** counts your gems and Gold, asks the node which block it is at and whether the level is open. If something is missing, it says so in plain words and **sends nothing** | The network rejected JC's Platinum on 28-Sep because the level wasn't open yet; with this he'd have known before trying |
| Forging was immediate and irreversible | **Summarizes and asks:** "You are about to forge 1 Platinum… 10 Diamonds are burned FOREVER… Confirm?" Only `yes` continues | What is burned doesn't come back. A slip of the finger shouldn't cost you 10 gems |
| `send` signed and broadcast without more | Says how much, to whom and **in how many transactions** (a large send is split in batches), and asks | 10,000 RUPIX go out as 232 transactions; you should know before signing |
| `--gem-address` required | Optional: if you don't give it, the gem is born at your first address | One less step for beginners |
| `--password=` on the command line | Asked on screen, without echo | It ended up in your history and in screenshots |
| Raw node errors (`nivel 2 bloqueado…`) | Explained in plain words, with the original below in parentheses | So you know what to do, not just what happened |
| Spanish only | **Spanish or English**, switchable any time | Rupix is for everyone |

**Choosing the language:** `./rupixwallet language en` saves it (or `es`). Just once: `--lang=en` before the command. The `RUPIX_LANG` variable works too. If you say nothing, it uses your system's.

**For scripts** that can't answer questions: `--yes` on `forge`, `send` and `transfer-gem`.

None of this touches consensus: the network keeps checking everything the same way. The wallet just warns you first.
