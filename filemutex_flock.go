// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux netbsd openbsd

package filemutex

import (
	"sync"
	"syscall"
)

const (
	mkdirPerm = 0750
)

// FileMutex is similar to sync.RWMutex, but also synchronizes across processes.
// This implementation is based on flock syscall.
type FileMutex struct {
	mu   sync.RWMutex
	fd   int
	path string
}

func New(filename string) (*FileMutex, error) {
	fd, err := syscall.Open(filename, syscall.O_CREAT|syscall.O_RDONLY, mkdirPerm)
	if err != nil {
		return nil, err
	}
	return &FileMutex{fd: fd, path: filename}, nil
}

func (m *FileMutex) Lock() {
	m.mu.Lock()
	if err := syscall.Flock(m.fd, syscall.LOCK_EX); err != nil {
		panic(err)
	}
}

func (m *FileMutex) Unlock() {
	if err := syscall.Flock(m.fd, syscall.LOCK_UN); err != nil {
		panic(err)
	}
	m.mu.Unlock()
}

func (m *FileMutex) RLock() {
	m.mu.RLock()
	if err := syscall.Flock(m.fd, syscall.LOCK_SH); err != nil {
		panic(err)
	}
}

func (m *FileMutex) RUnlock() {
	if err := syscall.Flock(m.fd, syscall.LOCK_UN); err != nil {
		panic(err)
	}
	m.mu.RUnlock()
}

// Close does an Unlock() combined with closing and unlinking the associated
// lock file. You should create a New() FileMutex for every Lock() attempt if
// using Close().
func (m *FileMutex) Close() {
	if err := syscall.Flock(m.fd, syscall.LOCK_UN); err != nil {
		panic(err)
	}
	syscall.Close(m.fd)
	syscall.Unlink(m.path)
	m.mu.Unlock()
}
