// Copyright 2015 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

//Start Procedure
//    funcName：procedure function name
//    inLen：input parameter length
//    outLen：output parameter length
func (session *Session) StartProcedure(funcName string, inLen, outLen int) (p *Procedure) {
	return callProcedure(session.engine, funcName, inLen, outLen)
}
