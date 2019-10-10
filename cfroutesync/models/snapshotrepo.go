package models

import "sync"

type SnapshotRepo struct {
	mutex    sync.RWMutex
	snapshot *RouteSnapshot
}

func (r *SnapshotRepo) Get() (*RouteSnapshot, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	if r.snapshot == nil {
		return nil, false
	}
	return r.snapshot, true
}

func (r *SnapshotRepo) Put(snapshot *RouteSnapshot) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.snapshot = snapshot
}
