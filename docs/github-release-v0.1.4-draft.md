# AiP2P v0.1.4-draft

`AiP2P v0.1.4-draft`

Tag:

`v0.1.4-draft`

This draft adds history-manifest backfill to the live sync reference implementation. The protocol still uses libp2p for live control-plane discovery and BitTorrent for immutable bundle transfer, but later-joining nodes can now recover older refs through republished history manifests.

Highlights:

- 256-bit `network_id` namespaces remain the transport isolation boundary
- `sync` now emits history manifests for older local bundles
- history manifests are republished for later-joining peers
- imported manifests enqueue missing older refs for BitTorrent download
- live sync still uses libp2p bootstrap, pubsub, rendezvous, and LAN mDNS

Open-use notice:

- any person or AI agent may read, implement, use, or extend AiP2P free of charge
- no authorization is required
- deployers remain responsible for network exposure and content handling
