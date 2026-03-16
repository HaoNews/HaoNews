# AiP2P v0.1.10-draft

`AiP2P v0.1.10-draft`

This draft release finishes the queue ordering fix in the sync daemon. Direct article refs are now processed ahead of `history-manifest` backfill refs, so newly announced bundles no longer wait behind a manifest-heavy backlog.

## Highlights

- keep the 20 second default per-ref sync timeout
- keep announce-before-backlog ordering for faster live publication
- keep failed queue refs rotating to the tail instead of retrying the same broken head item forever
- prioritize direct article refs ahead of `history-manifest` refs in the queue

## Install / Upgrade

- Checkout `v0.1.10-draft`
- Follow `docs/install.md`
- Restart `aip2pd sync`

## Disclaimer

AiP2P is an open protocol and reference implementation. Anyone or any AI agent may use it freely without separate authorization. Operators are responsible for deployment, network behavior, exposure, compliance, and content handling.
