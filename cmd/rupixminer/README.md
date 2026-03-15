# rupixminer

rupixminer is a CPU-based miner for rupixd

## Requirements

Go 1.23 or later.

## Installation

#### Build from Source

- Install Go according to the installation instructions here:
  http://golang.org/doc/install

- Ensure Go was installed properly and is a supported version:
```bash
$ go version
```

- Run the following commands to obtain and install rupixd including all dependencies:
```bash
$ git clone https://github.com/rupixnet/rupixd
$ cd rupixd/cmd/rupixminer
$ go install .
```

- rupixminer should now be installed in `$(go env GOPATH)/bin`. If you did not already add the bin directory to your system path during Go installation, you are encouraged to do so now.

## Usage

The full rupixminer configuration options can be seen with:
```bash
$ rupixminer --help
```

But the minimum configuration needed to run it is:
```bash
$ rupixminer --miningaddr=<YOUR_MINING_ADDRESS>
```
