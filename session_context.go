// Copyright 2018 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"context"
)

type cacheContext struct {
	context.Context
	values map[string]interface{}
}

func (c *cacheContext) Done() <-chan struct{} {
	for k := range c.values {
		delete(c.values, k)
	}
	return nil
}

func (c *cacheContext) Value(key interface{}) interface{} {
	return c.values[key.(string)]
}

func WithCacher(ctx context.Context) context.Context {
	return &cacheContext{
		Context: ctx,
		values:  make(map[string]interface{}),
	}
}
