# 🔷 How to join the Rupix testnet

🇪🇸 [Versión en español](./GUIA-TESTNET.md)

A simple guide to connect your node to the Rupix test network.
You don't need to be an expert — just follow the steps.

> **Rupix testnet** is a TEST network. Testnet coins (RUPIX) have NO
> real value. It's for testing, learning and helping strengthen the
> network before mainnet. Thanks for joining!

> **⚠️ Windows vs Linux/Mac:** the commands below start with `./` (Linux/Mac).
> In **Windows PowerShell**, use `.\` instead of `./`, for example: `.\rupixd.exe`.
> Without `./` or `.\`, the system can't find the program. (Thanks to JC for the tip!)
---

## Step 1 — Download Rupix

Go to the official downloads (always the latest version):
**https://github.com/rupixnet/rupixd/releases/latest**

Download the file for your system:
- **Windows:** `rupix-v0.6.2-win64.zip`
- **Mac:** `rupix-v0.6.2-osx.zip`
- **Linux:** `rupix-v0.6.2-linux.zip`

Unzip it. Inside you'll find 4 programs:
`rupixd` (the node), `rupixctl` (control), `rupixwallet` (wallet), `rupixminer` (miner).

---

## Step 2 — Start your node and connect to the network

Open a terminal in the folder where you unzipped Rupix and run:

**Linux / Mac:**
```
./rupixd --testnet --utxoindex --addpeer=178.104.69.148:17211
```

**Windows:**
```
.\rupixd.exe --testnet --utxoindex --addpeer=178.104.69.148:17211
```

Your node will connect to the Rupix seed node and start
**syncing** (downloading the chain). You'll see lots of messages —
that's normal. When it catches up with the latest block,
you're on the network! 🎉

---

## Step 3 — Check that you're connected

In ANOTHER terminal (leave the node running in the first one), run:

**Linux / Mac:**
```
./rupixctl --testnet GetBlockDagInfo
```

**Windows:**
```
.\rupixctl.exe --testnet GetBlockDagInfo
```

You should see:
- `networkName: "rupix-testnet"` (you're on the right network)
- `blockCount` going up (your node is syncing)

Compare your `blockCount` with the official explorer
(**https://explorer.rupix.network**). When they're close, you're up to date.

---

## Step 4 (optional) — Mine Rupix

> **About mining:** Rupix uses **RupixHeavyHash**, its own algorithm
> derived from Kaspa's engine (kHeavyHash; Rupix is a fork of kaspad) with its own seed.
> You can mine with your CPU or GPU, and on testnet your computer can win
> blocks. What matters most right now is running your **node** (steps 2-3); mining is
> an extra that also helps the network.
>
> RupixHeavyHash stops the industrial machines built for Kaspa from
> mining Rupix — a fair start so everyone begins on equal footing. Honestly:
> it's an advantage of months, not permanent independence; an FPGA can be
> reprogrammed. It's a fair start, not an eternal barrier.
>
> **About the keys file:** if you start the daemon with
> `--keys-file=PATH`, the `send` command must also use the same
> `--keys-file=PATH` (it signs locally with those keys). The other commands
> (`new-address`, `balance`, `gems`, `forge`) do NOT take it: they ask the daemon.
> If you don't use `--keys-file` anywhere, everything uses the default file.

Want to help mine the testnet? First create your wallet (only once):

```
./rupixwallet --testnet create
```
(It will ask for a password. Keep it safe.) **When it finishes it shows your seed phrase once:** write it on paper, in order, and clear the screen (`clear` / `cls`). That phrase recovers the wallet if you lose the file or the password; without it, there is no way. Don't keep it on your phone or send it over chat. Start the daemon and leave it running in its own window:
```
./rupixwallet --testnet start-daemon
```
In ANOTHER window, create your address:
```
./rupixwallet --testnet new-address
```

Copy your address (it starts with `rupixtest:...`) and start the miner:

```
./rupixminer --testnet --miningaddr=YOUR_ADDRESS_HERE
```

You'll be mining testnet Gold and helping strengthen the network!

---

## Problems?

- **Not syncing / can't connect:** check your internet and that you typed
  the peer address correctly (`178.104.69.148:17211`).
- **"address already in use":** you already have a node running; close it first.
- Write to me and we'll sort it out together.

---

*Rupix testnet — the public lab. We can all mine, we can all verify.*
**Don't trust. Verify — from genesis.**
🔗 rupix.network · github.com/rupixnet/rupixd · explorer.rupix.network
