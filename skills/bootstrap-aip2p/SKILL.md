---
name: bootstrap-aip2p
description: Install, update, pin, or roll back the AiP2P GitHub repository, then verify the checked out version by running Go tests and the reference CLI help. Use when an AI agent needs to set up AiP2P from GitHub without human file copying.
---

# Bootstrap AiP2P

Use this skill when the agent needs to install or update the `AiP2P` repository directly from GitHub.

## Inputs To Decide First

- target directory
- version mode: `main`, latest tag, or fixed tag
- whether the goal is install, update, or rollback
- operating system: macOS, Linux, or Windows PowerShell

If the user does not specify a version, prefer the latest released tag. If no tag is requested and active development is desired, use `main`.

Prefer PowerShell-native commands on Windows. Do not give Unix shell substitutions such as `$(...)` to a PowerShell-only environment.

## Public Internet Helper Node

If the user wants AiP2P nodes in different private networks to connect more reliably, also read:

- [`docs/public-bootstrap-node.md`](../../docs/public-bootstrap-node.md)

Important boundary:

- this repository does not yet include a ready-made public bootstrap/rendezvous/relay server binary
- do not invent unsupported repository commands
- treat the public helper node as a separate deployment task, then write its final public multiaddrs into `aip2p_net.inf`

## Workflow

1. Clone the repository if it does not exist:

macOS / Linux:

```bash
git clone https://github.com/AiP2P/AiP2P.git
cd AiP2P
```

Windows PowerShell:

```powershell
git clone https://github.com/AiP2P/AiP2P.git
Set-Location AiP2P
```

2. Fetch the newest refs:

macOS / Linux:

```bash
git fetch --tags origin
```

Windows PowerShell:

```powershell
git fetch --tags origin
```

3. Choose one checkout mode:

- newest development:

  macOS / Linux:

```bash
git checkout main
git pull --ff-only origin main
```

  Windows PowerShell:

```powershell
git checkout main
git pull --ff-only origin main
```

- newest released tag:

  macOS / Linux:

```bash
git checkout "$(git tag --sort=-version:refname | head -n 1)"
```

  Windows PowerShell:

```powershell
$latestTag = git tag --sort=-version:refname | Select-Object -First 1
git checkout $latestTag
```

- exact pinned version:

```bash
git checkout <tag-or-commit>
```

4. Verify the checkout:

macOS / Linux:

```bash
go test ./...
go run ./cmd/aip2p
```

Windows PowerShell:

```powershell
go test ./...
go run ./cmd/aip2p
```

Expected CLI usage output currently includes:

- `publish`
- `verify`
- `show`

## Rollback

Prefer rolling back to a released tag:

macOS / Linux:

```bash
git fetch --tags origin
git checkout <older-tag>
go test ./...
```

Windows PowerShell:

```powershell
git fetch --tags origin
git checkout <older-tag>
go test ./...
```

## Agent Notes

- Do not invent unpublished commands.
- Treat this repository as protocol and reference tooling only.
- Downstream product behavior belongs in `Latest`, not here.
- For user-facing installation guidance, also read [`docs/install.md`](../../docs/install.md).
