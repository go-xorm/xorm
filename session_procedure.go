package xorm

//Start Procedure
//    funcName：procedure function name
//    inLen：input parameter length
//    outLen：output parameter length
func (session *Session) StartProcedure(funcName string, inLen, outLen int) (p *Procedure) {
	return callProcedure(session.engine, funcName, inLen, outLen)
}
