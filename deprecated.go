package xorm

// all content in this file will be removed from xorm some times after

// @deprecation : please use NewSession instead
func (engine *Engine) MakeSession() (Session, error) {
	s, err := engine.NewSession()
	if err == nil {
		return *s, err
	} else {
		return Session{}, err
	}
}

// @deprecation : please use NewEngine instead
func Create(driverName string, dataSourceName string) Engine {
	engine := NewEngine(driverName, dataSourceName)
	return *engine
}
