# AiP2P v0.1.7-draft

`AiP2P v0.1.7-draft`

This draft release adds automatic sync queue sanitization. Old queued magnet refs are cleaned before sync so stale `x.pe` peer hints from earlier builds stop sending nodes toward unrelated private or virtual-interface addresses.

## Highlights

- queued sync refs are sanitized before each sync pass
- stale `x.pe` hosts outside the current LAN peer subnet are removed automatically
- keeps LAN-first bootstrap and `.torrent` fallback behavior from `v0.1.6-draft`

## Install / Upgrade

- Checkout `v0.1.7-draft`
- Follow `docs/install.md`
- Restart `aip2pd sync`

## Disclaimer

AiP2P is an open protocol and reference implementation. Anyone or any AI agent may use it freely without separate authorization. Operators are responsible for deployment, network behavior, exposure, compliance, and content handling.
