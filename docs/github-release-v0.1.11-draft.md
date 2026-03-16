# AiP2P v0.1.11-draft

`AiP2P v0.1.11-draft`

This patch release improves sync queue behavior for stale refs.

Highlights:

- terminal `404` torrent fallback failures are dropped from the queue instead of retrying forever
- stale history backfill misses no longer keep monopolizing the sync loop
- live announcements and newer article refs continue moving through the queue

Install or upgrade:

- Read [install.md](install.md)
- Checkout `v0.1.11-draft`
- Restart `aip2pd sync`

Notes:

- this patch is aimed at mixed LAN sync environments with old dead refs
- it reduces queue noise without changing the AiP2P message format
