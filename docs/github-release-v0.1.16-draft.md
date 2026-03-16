# AiP2P v0.1.16-draft

`AiP2P v0.1.16-draft`

This patch adds default `Trackerlist.inf` support for BitTorrent peer discovery.

Highlights:

- `aip2p sync` auto-loads `Trackerlist.inf` next to the net config
- magnet imports merge configured trackers
- `.torrent` imports merge configured trackers
- operators can override the tracker file with `--trackers`
