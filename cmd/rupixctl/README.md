# rupixctl

rupixctl is an RPC client for rupixd

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
$ cd rupixd/cmd/rupixctl
$ go install .
```

- rupixctl should now be installed in `$(go env GOPATH)/bin`. If you did not already add the bin directory to your system path during Go installation, you are encouraged to do so now.

## Usage

The full rupixctl configuration options can be seen with:
```bash
$ rupixctl --help
```

But the minimum configuration needed to run it is:
```bash
$ rupixctl <REQUEST_JSON>
```

For example:
```
$ rupixctl '{"getBlockDagInfoRequest":{}}'
```

For a list of all available requests check out the [RPC documentation](infrastructure/network/netadapter/server/grpcserver/protowire/rpc.md)
