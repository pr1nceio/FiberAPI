package utils

import (
	consul "github.com/hashicorp/consul/api"
	"sync"
	"time"
)

type SuperLock struct {
	localLock map[string]*sync.Mutex
	csul      *consul.KV
	sessionID string
}

func NewSuperLock(consul *consul.KV, consulSession string) *SuperLock {
	return &SuperLock{
		localLock: make(map[string]*sync.Mutex),
		csul:      consul,
		sessionID: consulSession,
	}
}

func (s *SuperLock) Lock(lockName string) {
	lock, ok := s.localLock[lockName]
	if !ok {
		lock = &sync.Mutex{}
		s.localLock[lockName] = lock
	}
	lock.Lock()
	if s.csul == nil {
		return
	}
	for {
		isAcq, _, err := s.csul.Acquire(&consul.KVPair{
			Key:     "sessions/fiberapi_lock/" + lockName,
			Value:   []byte(GetEnv("NOMAD_SHORT_ALLOC_ID", "default")),
			Session: s.sessionID,
		}, nil)
		if err == nil && isAcq {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func (s *SuperLock) Unlock(lockName string) {

	for {
		if s.csul == nil {
			break
		}
		isRel, _, err := s.csul.Release(&consul.KVPair{
			Key:     "sessions/fiberapi_lock/" + lockName,
			Value:   []byte(GetEnv("NOMAD_SHORT_ALLOC_ID", "default")),
			Session: s.sessionID,
		}, nil)
		if err == nil && isRel {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	lock, ok := s.localLock[lockName]
	if !ok {
		lock = &sync.Mutex{}
		s.localLock[lockName] = lock
	}
	lock.Unlock()
}

func (s *SuperLock) ReleaseAll() {
	for k := range s.localLock {
		s.Unlock(k)
		for {
			if s.csul == nil {
				break
			}
			isRel, _, err := s.csul.Release(&consul.KVPair{
				Key:     "sessions/fiberapi_lock/" + k,
				Value:   []byte(GetEnv("NOMAD_SHORT_ALLOC_ID", "default")),
				Session: s.sessionID,
			}, nil)
			if err == nil && isRel {
				break
			}
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func (s *SuperLock) ExposeConsul() *consul.KV {
	return s.csul
}
