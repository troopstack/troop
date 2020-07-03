package cache

import (
	"sync"

	"github.com/troopstack/troop/src/model"
)

type SafeScouts struct {
	sync.RWMutex
	M map[string]*model.ScoutInfo
}

var Scouts = NewSafeScouts()

func NewSafeScouts() *SafeScouts {
	return &SafeScouts{M: make(map[string]*model.ScoutInfo)}
}

func (S *SafeScouts) Put(req *model.ScoutInfo) {
	if scoutInfo, exists := S.Get(req.Hostname); !exists ||
		scoutInfo.ScoutVersion != req.ScoutVersion ||
		scoutInfo.Type != req.Type ||
		scoutInfo.IP != req.IP ||
		scoutInfo.PubKey != req.PubKey ||
		scoutInfo.AES != req.AES ||
		scoutInfo.Status != req.Status {
		S.Lock()
		S.M[req.Hostname] = req
		S.Unlock()
	}
}

func (S *SafeScouts) Get(hostname string) (*model.ScoutInfo, bool) {
	S.RLock()
	defer S.RUnlock()
	val, exists := S.M[hostname]
	return val, exists
}

func (S *SafeScouts) Delete(hostname string) {
	S.Lock()
	defer S.Unlock()
	delete(S.M, hostname)
}

func (S *SafeScouts) Keys() []string {
	S.RLock()
	defer S.RUnlock()
	count := len(S.M)
	keys := make([]string, count)
	i := 0
	for hostname := range S.M {
		keys[i] = hostname
		i++
	}
	return keys
}

func (S *SafeScouts) All() []*model.ScoutInfo {
	S.RLock()
	defer S.RUnlock()
	count := len(S.M)
	keys := make([]*model.ScoutInfo, count)
	i := 0
	for hostname := range S.M {
		host := S.M[hostname]
		keys[i] = host
		i++
	}
	return keys
}

func (S *SafeScouts) UnacceptedHosts() []*model.ScoutInfo {
	S.RLock()
	defer S.RUnlock()
	count := len(S.M)
	keys := make([]*model.ScoutInfo, count)
	i := 0
	for hostname := range S.M {
		host := S.M[hostname]
		if host.Status != "unaccepted" {
			continue
		}
		keys[i] = host
		i++
	}
	return keys
}

func (S *SafeScouts) DeniedHosts() []*model.ScoutInfo {
	S.RLock()
	defer S.RUnlock()
	count := len(S.M)
	keys := make([]*model.ScoutInfo, count)
	i := 0
	for hostname := range S.M {
		host := S.M[hostname]
		if host.Status != "denied" {
			continue
		}
		keys[i] = host
		i++
	}
	return keys
}
