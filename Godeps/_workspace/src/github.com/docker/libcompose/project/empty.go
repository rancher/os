package project

type EmptyService struct {
}

func (e *EmptyService) Create() error {
	return nil
}

func (e *EmptyService) Build() error {
	return nil
}

func (e *EmptyService) Up() error {
	return nil
}

func (e *EmptyService) Start() error {
	return nil
}

func (e *EmptyService) Down() error {
	return nil
}

func (e *EmptyService) Delete() error {
	return nil
}

func (e *EmptyService) Restart() error {
	return nil
}

func (e *EmptyService) Log() error {
	return nil
}

func (e *EmptyService) Pull() error {
	return nil
}

func (e *EmptyService) Kill() error {
	return nil
}

func (e *EmptyService) Containers() ([]Container, error) {
	return []Container{}, nil
}

func (e *EmptyService) Scale(count int) error {
	return nil
}

func (e *EmptyService) Info() (InfoSet, error) {
	return InfoSet{}, nil
}
