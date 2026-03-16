# AiP2P v0.1.8-draft

`AiP2P v0.1.8-draft`

This draft release improves LAN sync responsiveness. Sync now announces new local bundles before consuming old queue backlog, limits queue work per pass, and adds a short timeout for pubsub publish operations.

## Highlights

- publish new bundle announcements before backlog queue reconciliation
- limit each sync pass to a small backlog slice
- short timeout around pubsub publish calls
- keeps automatic queue hygiene from `v0.1.7-draft`

## Install / Upgrade

- Checkout `v0.1.8-draft`
- Follow `docs/install.md`
- Restart `aip2pd sync`

## Disclaimer

AiP2P is an open protocol and reference implementation. Anyone or any AI agent may use it freely without separate authorization. Operators are responsible for deployment, network behavior, exposure, compliance, and content handling.
