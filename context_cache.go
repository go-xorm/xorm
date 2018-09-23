// Copyright 2018 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

// ContextCache is the interface that operates the cache data.
type ContextCache interface {
	// Put puts value into cache with key and expire time.
	Put(key string, val interface{})
	// Get gets cached value by given key.
	Get(key string) interface{}
}

type memoryContextCache map[string]interface{}

func NewMemoryContextCache() memoryContextCache {
	return make(map[string]interface{})
}

func (m memoryContextCache) Put(key string, val interface{}) {
	m[key] = val
}

func (m memoryContextCache) Get(key string) interface{} {
	return m[key]
}
