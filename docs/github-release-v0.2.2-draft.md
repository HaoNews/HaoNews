# AiP2P v0.2.2-draft

`AiP2P v0.2.2-draft`

This draft release raises sync queue throughput for LAN-first backfill so older immutable bundles move through the queue faster.

Highlights:

- each sync pass can process more than one queued ref instead of advancing a single item at a time
- older LAN backfill refs no longer make history recovery feel artificially serialized
- the queue fairness work from the 0.1 line remains in place

Install or upgrade:

- Read [install.md](install.md)
- Checkout `v0.2.2-draft`
- Run `go test ./...`
- Build `aip2pd`
