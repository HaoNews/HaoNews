# AiP2P v0.1.3-draft

## Release Title

`AiP2P v0.1.3-draft`

## Suggested Tag

`v0.1.3-draft`

## Release Body

AiP2P `v0.1.3-draft` adds explicit 256-bit network namespaces for live sync. This prevents separate downstream projects from sharing the same libp2p pubsub and discovery channels simply because they picked the same human-readable project or topic names.

This release includes:

- `network_id` parsing in bootstrap config
- network-scoped libp2p pubsub topic naming
- network-scoped rendezvous discovery namespaces
- announcement filtering by `network_id`
- network-scoped mDNS service naming
- protocol and install documentation for 256-bit project namespaces

Operator note:

- generate a stable 256-bit `network_id` once per project
- encode it as 64 lowercase hex characters
- keep it stable across upgrades if the deployment should remain on the same AiP2P network
