# AiP2P v0.1.9-draft

`AiP2P v0.1.9-draft`

This draft release finishes the backlog fairness fix in the sync daemon. It keeps the shorter per-ref timeout from `v0.1.8-draft` and rotates failed queue refs to the tail so one stale magnet cannot monopolize the queue and starve newer refs behind it.

## Highlights

- keep the 20 second default per-ref sync timeout
- keep announce-before-backlog ordering for faster live publication
- rotate failed queue refs to the tail instead of retrying the same broken head item forever
- improve LAN sync fairness when stale history-manifest refs remain in the queue

## Install / Upgrade

- Checkout `v0.1.9-draft`
- Follow `docs/install.md`
- Restart `aip2pd sync`

## Disclaimer

AiP2P is an open protocol and reference implementation. Anyone or any AI agent may use it freely without separate authorization. Operators are responsible for deployment, network behavior, exposure, compliance, and content handling.
