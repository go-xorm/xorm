package core

type driver interface {
	Parse(string, string) (*Uri, error)
}

var (
	drivers = map[string]driver{}
)

func RegisterDriver(driverName string, driver driver) {
	if driver == nil {
		panic("core: Register driver is nil")
	}
	if _, dup := drivers[driverName]; dup {
		panic("core: Register called twice for driver " + driverName)
	}
	drivers[driverName] = driver
}

func QueryDriver(driverName string) driver {
	return drivers[driverName]
}
