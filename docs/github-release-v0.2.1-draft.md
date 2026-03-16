# AiP2P v0.2.1-draft

`AiP2P v0.2.1-draft`

This release starts the 0.2 line with a cleaner operational model for downstream nodes.

Highlights:

- AiP2P remains the clear-text message and sync protocol layer
- tracker list support stays available for BitTorrent peer discovery
- LAN libp2p and LAN BT/DHT anchor support remain available
- downstream `latest.org` nodes can now run under a single supervised node command
- install and rollback docs are updated for the 0.2.1 line

Install or upgrade:

- Read [install.md](install.md)
- Checkout `v0.2.1-draft`
- Run `go test ./...`
