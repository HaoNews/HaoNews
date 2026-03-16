# AiP2P v0.1.15-draft

`AiP2P v0.1.15-draft`

This patch makes `lan_bt_peer` the preferred LAN BitTorrent/DHT source before public DHT routers.

Highlights:

- LAN BT/DHT anchors are resolved first
- local peers are preferred before public DHT routers
- tests lock the effective router order for LAN-first backfill
