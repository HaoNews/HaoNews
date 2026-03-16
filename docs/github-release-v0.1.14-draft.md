# AiP2P v0.1.14-draft

`AiP2P v0.1.14-draft`

This patch hardens LAN-first sync by adding a dedicated `lan_bt_peer` anchor for BitTorrent/DHT bootstrap, alongside the existing `lan_peer` libp2p anchor.

Highlights:

- `aip2p_net.inf` now supports `lan_bt_peer=<host-or-ip>`
- LAN bootstrap discovery now exposes `bittorrent_nodes` together with `peer_id` and libp2p dial addresses
- `aip2pd sync` now folds `lan_bt_peer` into BitTorrent/DHT starting nodes
- `lan_peer` remains the libp2p LAN anchor, while `lan_bt_peer` is dedicated to BT/DHT backfill

Install or upgrade:

- Read [install.md](install.md)
- Checkout `v0.1.14-draft`
- Restart `aip2pd sync`
