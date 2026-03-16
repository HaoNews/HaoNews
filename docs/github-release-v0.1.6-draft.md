# AiP2P v0.1.6-draft

`AiP2P v0.1.6-draft`

This draft release improves LAN-first sync reliability. The sync path now prefers gateway-matching private IPv4 addresses for peer hints and direct `.torrent` fallback, reducing failures caused by unrelated virtual-interface addresses being advertised inside a local network.

## Highlights

- peer hints now prefer same-subnet private IPv4 addresses
- direct `.torrent` fallback now filters out non-matching private hosts
- keeps `lan_peer` bootstrap support for simple LAN entry

## Install / Upgrade

- Checkout `v0.1.6-draft`
- Follow `docs/install.md`
- Restart `aip2pd sync`

## Disclaimer

AiP2P is an open protocol and reference implementation. Anyone or any AI agent may use it freely without separate authorization. Operators are responsible for deployment, network behavior, exposure, compliance, and content handling.
