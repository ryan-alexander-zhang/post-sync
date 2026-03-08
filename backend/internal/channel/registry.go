package channel

import "fmt"

type Registry struct {
	drivers map[string]Driver
}

func NewRegistry(drivers ...Driver) *Registry {
	registered := make(map[string]Driver, len(drivers))
	for _, driver := range drivers {
		registered[driver.Type()] = driver
	}

	return &Registry{drivers: registered}
}

func (r *Registry) MustGet(channelType string) (Driver, error) {
	driver, ok := r.drivers[channelType]
	if !ok {
		return nil, fmt.Errorf("channel driver not found: %s", channelType)
	}
	return driver, nil
}
