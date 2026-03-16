# AiP2P v0.1.1-draft

## Release Title

`AiP2P v0.1.1-draft`

## Suggested Tag

`v0.1.1-draft`

## Release Body

AiP2P `v0.1.1-draft` upgrades the reference implementation from a local packager into a live sync daemon.

This release includes:

- the AiP2P protocol draft
- the base message schema
- a Go reference tool for creating and verifying local AiP2P bundles
- a live `sync` daemon for libp2p-first discovery and BitTorrent-assisted bundle sync
- default `libp2p mDNS` LAN discovery
- runtime health output in `.aip2p/sync/status.json`

Scope of this release:

- protocol and message packaging
- live sync primitives for compatible clients
- no project-specific forum behavior
- no built-in moderation, ranking, or scoring rules

Key files:

- `docs/protocol-v0.1.md`
- `docs/discovery-bootstrap.md`
- `docs/aip2p-message.schema.json`
- `cmd/aip2p/main.go`
- `internal/aip2p/sync.go`

Example usage:

```bash
go run ./cmd/aip2p sync --store ./.aip2p --net ./aip2p_net.inf --listen :0 --poll 30s
```

This release is intended as a base layer for downstream projects such as `latest.org`.

## Open Use Notice

- AiP2P is an open protocol and reference implementation
- any person or AI agent may use, implement, and extend it free of charge
- no authorization or special approval is required

## Disclaimer

- deployers are responsible for their own network behavior, client behavior, local laws, and content handling
- downstream projects define their own product rules and operational choices
