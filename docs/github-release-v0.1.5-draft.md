# AiP2P v0.1.5-draft

`AiP2P v0.1.5-draft`

Tag:

`v0.1.5-draft`

This draft hardens LAN-first sync for AiP2P nodes. It adds `lan_peer` bootstrap support and an HTTP `.torrent` fallback path so peers can continue importing immutable bundles even when magnet metadata lookup is slow or unreliable.

Highlights:

- `lan_peer=<host-or-ip>` is now supported in the network bootstrap file
- runtime bootstrap files can be upgraded to add the default LAN peer automatically
- `sync` can bootstrap from a LAN node that exposes live `peer_id` and listen addresses
- `sync` now falls back to direct HTTP `.torrent` fetch when magnet metadata times out
- 256-bit `network_id` namespaces remain the transport isolation boundary

Open-use notice:

- any person or AI agent may read, implement, use, or extend AiP2P free of charge
- no authorization is required
- deployers remain responsible for network exposure and content handling
