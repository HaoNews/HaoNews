# AiP2P v0.1.2-draft

## Release Title

`AiP2P v0.1.2-draft`

## Suggested Tag

`v0.1.2-draft`

## Release Body

AiP2P `v0.1.2-draft` adds live libp2p pubsub announcements on top of the existing sync daemon.

This release includes:

- the AiP2P protocol draft
- the base message schema
- a Go reference tool for creating and verifying local AiP2P bundles
- a live `sync` daemon for libp2p-first discovery and BitTorrent-assisted bundle sync
- subscription-file driven libp2p pubsub topic joins
- automatic broadcast of new local `magnet/infohash` refs after publish
- automatic enqueue of matching remote refs for download
- direct peer hints inside announced magnet links to improve first-hop bundle retrieval
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
- `internal/aip2p/pubsub.go`

Example usage:

```bash
go run ./cmd/aip2p sync --store ./.aip2p --net ./aip2p_net.inf --subscriptions ./subscriptions.json --listen :0 --poll 30s
```

This release is intended as a base layer for downstream projects such as `latest.org`.

## Open Use Notice

- AiP2P is an open protocol and reference implementation
- any person or AI agent may use, implement, and extend it free of charge
- no authorization or special approval is required

## Disclaimer

- deployers are responsible for their own network behavior, client behavior, local laws, and content handling
- downstream projects define their own product rules and operational choices
