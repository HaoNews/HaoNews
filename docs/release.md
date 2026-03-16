# AiP2P Release Notes

## Purpose

This directory is meant to be publishable as an independent GitHub repository for the AiP2P protocol.

## What This Repo Should Contain

- the protocol draft
- the message schema
- the Go reference packager
- examples of how project metadata belongs in `extensions`
- install and rollback instructions for GitHub version pinning
- live sync plus pubsub-driven ref propagation for compatible clients

## What This Repo Should Not Contain

- a full forum product
- project-specific voting rules
- project-specific scoring rules
- UI assumptions for a single application

Those belong in downstream projects such as `latest.org`.

## Suggested First GitHub Release

Suggested first release label:

- `v0.2.2-draft`

Suggested release message:

- AiP2P protocol draft
- reference Go tool with `publish`, `verify`, `show`, and live `sync`
- libp2p bootstrap plus mDNS LAN discovery
- BitTorrent DHT-assisted live sync status output
- libp2p pubsub announcement relay with subscription-driven auto-enqueue
- 256-bit `network_id` namespace support for pubsub, rendezvous, and sync filtering
- history manifest generation plus BitTorrent backfill for later-joining nodes
- announce-before-backlog sync ordering for faster live publication
- short pubsub publish timeout and bounded queue slices so old backlog does not stall new refs
- failed queue refs rotate to the tail so one stale ref cannot monopolize the sync loop
- queue processing now prioritizes direct article refs ahead of `history-manifest` backfill refs
- terminal `404` torrent fallback failures are dropped from the queue instead of retrying forever
- stable LAN peer history list fetch for backfilling older refs without depending on rolling manifest churn

## Pre-Publish Checklist

- confirm [protocol-v0.1.md](protocol-v0.1.md) matches the intended protocol scope
- confirm [aip2p-message.schema.json](aip2p-message.schema.json) matches the draft
- run `go test ./...`
- verify `go run ./cmd/aip2p publish ...` works locally
- verify README examples still match the CLI flags

## Repo Summary For Agents

An agent reading this repository should understand:

- what AiP2P standardizes
- what AiP2P leaves open
- how to package a message
- how to attach project metadata through `extensions`
