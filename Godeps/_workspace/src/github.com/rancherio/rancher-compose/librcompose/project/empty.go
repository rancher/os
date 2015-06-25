package project

type EmptyService struct {
}

func (e *EmptyService) Create() error {
	return nil
}

func (e *EmptyService) Up() error {
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
func (e *EmptyService) Scale(count int) error {
	return nil
}
