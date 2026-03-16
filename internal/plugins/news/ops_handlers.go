package newsplugin

import (
	"net/http"
	"strings"
	"time"
)

func (a *App) handleNetwork(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/network" {
		http.NotFound(w, r)
		return
	}
	index, err := a.index()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	syncStatus, err := a.syncRuntimeStatus()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	supervisorStatus, err := a.syncSupervisorStatus()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	netCfg, err := a.networkBootstrap()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	anchors, hasLANBTMatch, lanBTOverall := a.lanBTStatus(r.Context(), netCfg)
	data := NetworkPageData{
		Project:       displayProjectName(a.project),
		Version:       a.version,
		ListenAddr:    a.httpListenAddr(),
		PageNav:       a.pageNav("/network"),
		Now:           time.Now(),
		NodeStatus:    a.nodeStatus(index),
		SyncStatus:    syncStatus,
		Supervisor:    supervisorStatus,
		LANPeers:      append([]string(nil), netCfg.LANPeers...),
		LANBTAnchors:  anchors,
		LANBTHasMatch: hasLANBTMatch,
		LANBTOverall:  lanBTOverall,
	}
	if err := a.templates.ExecuteTemplate(w, "network.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *App) handleAPINetworkBootstrap(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/network/bootstrap" {
		http.NotFound(w, r)
		return
	}
	syncStatus, err := a.syncRuntimeStatus()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !syncStatus.LibP2P.Enabled || strings.TrimSpace(syncStatus.LibP2P.PeerID) == "" {
		http.Error(w, "libp2p sync daemon is not online on this node", http.StatusServiceUnavailable)
		return
	}
	dialAddrs := dialableLibP2PAddrs(syncStatus, requestBootstrapHost(r))
	if len(dialAddrs) == 0 {
		http.Error(w, "no dialable libp2p addresses available on this node", http.StatusServiceUnavailable)
		return
	}
	writeJSON(w, http.StatusOK, NetworkBootstrapResponse{
		Project:         a.project,
		Version:         a.version,
		NetworkID:       syncStatus.NetworkID,
		PeerID:          syncStatus.LibP2P.PeerID,
		ListenAddrs:     append([]string(nil), syncStatus.LibP2P.ListenAddrs...),
		DialAddrs:       dialAddrs,
		BitTorrentNodes: dialableBitTorrentNodes(syncStatus, requestBootstrapHost(r)),
	})
}
