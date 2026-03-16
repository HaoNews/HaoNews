package newsplugin

import (
	"context"
	"net/http"
)

func (a *App) HTTPListenAddr() string {
	return a.httpListenAddr()
}

func (a *App) NodeStatus(index Index) NodeStatus {
	return a.nodeStatus(index)
}

func (a *App) SyncRuntimeStatus() (SyncRuntimeStatus, error) {
	return a.syncRuntimeStatus()
}

func (a *App) SyncSupervisorStatus() (SyncSupervisorState, error) {
	return a.syncSupervisorStatus()
}

func (a *App) NetworkBootstrap() (NetworkBootstrapConfig, error) {
	return a.networkBootstrap()
}

func (a *App) LANBTStatus(ctx context.Context, cfg NetworkBootstrapConfig) ([]LANBTAnchorStatus, bool, string) {
	return a.lanBTStatus(ctx, cfg)
}

func RequestBootstrapHost(r *http.Request) string {
	return requestBootstrapHost(r)
}

func DialableLibP2PAddrs(status SyncRuntimeStatus, host string) []string {
	return dialableLibP2PAddrs(status, host)
}

func DialableBitTorrentNodes(status SyncRuntimeStatus, host string) []string {
	return dialableBitTorrentNodes(status, host)
}
