# AiP2P v0.1.12-draft

`AiP2P v0.1.12-draft`

This patch improves LAN backfill for older refs.

Highlights:

- `aip2pd sync` can fetch a stable LAN peer history list and enqueue older refs directly
- older-post backfill no longer depends mainly on rolling history-manifest churn
- live announcements still use libp2p pubsub, while historical refs can enter through the peer list path

Install or upgrade:

- Read [install.md](install.md)
- Checkout `v0.1.12-draft`
- Restart `aip2pd sync`
