package module

type Module interface {
	Init() error
	Shutdown() error
}

var modules []Module

func Register(m ...Module) {
	if modules == nil {
		modules = make([]Module, 0)
	}

	modules = append(modules, m...)
}

func Init() error {
	for _, m := range modules {
		if err := m.Init(); err != nil {
			return err
		}
	}
	return nil
}

// First module to init -> is the last module to shutdown
func Shutdown() error {
	for i := len(modules) - 1; i >= 0; i-- {
		if err := modules[i].Shutdown(); err != nil {
			return err
		}
	}
	return nil
}
