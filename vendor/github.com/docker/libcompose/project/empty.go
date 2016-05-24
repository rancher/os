package project

import (
	"golang.org/x/net/context"

	"github.com/docker/libcompose/project/options"
)

// EmptyService is a struct that implements Service but does nothing.
type EmptyService struct {
}

// Create implements Service.Create but does nothing.
func (e *EmptyService) Create(ctx context.Context, options options.Create) error {
	return nil
}

// Build implements Service.Build but does nothing.
func (e *EmptyService) Build(ctx context.Context, buildOptions options.Build) error {
	return nil
}

// Up implements Service.Up but does nothing.
func (e *EmptyService) Up(ctx context.Context, options options.Up) error {
	return nil
}

// Start implements Service.Start but does nothing.
func (e *EmptyService) Start(ctx context.Context) error {
	return nil
}

// Stop implements Service.Stop() but does nothing.
func (e *EmptyService) Stop(ctx context.Context, timeout int) error {
	return nil
}

// Delete implements Service.Delete but does nothing.
func (e *EmptyService) Delete(ctx context.Context, options options.Delete) error {
	return nil
}

// Restart implements Service.Restart but does nothing.
func (e *EmptyService) Restart(ctx context.Context, timeout int) error {
	return nil
}

// Log implements Service.Log but does nothing.
func (e *EmptyService) Log(ctx context.Context, follow bool) error {
	return nil
}

// Pull implements Service.Pull but does nothing.
func (e *EmptyService) Pull(ctx context.Context) error {
	return nil
}

// Kill implements Service.Kill but does nothing.
func (e *EmptyService) Kill(ctx context.Context, signal string) error {
	return nil
}

// Containers implements Service.Containers but does nothing.
func (e *EmptyService) Containers(ctx context.Context) ([]Container, error) {
	return []Container{}, nil
}

// Scale implements Service.Scale but does nothing.
func (e *EmptyService) Scale(ctx context.Context, count int, timeout int) error {
	return nil
}

// Info implements Service.Info but does nothing.
func (e *EmptyService) Info(ctx context.Context, qFlag bool) (InfoSet, error) {
	return InfoSet{}, nil
}

// Pause implements Service.Pause but does nothing.
func (e *EmptyService) Pause(ctx context.Context) error {
	return nil
}

// Unpause implements Service.Pause but does nothing.
func (e *EmptyService) Unpause(ctx context.Context) error {
	return nil
}

// Run implements Service.Run but does nothing.
func (e *EmptyService) Run(ctx context.Context, commandParts []string) (int, error) {
	return 0, nil
}

// RemoveImage implements Service.RemoveImage but does nothing.
func (e *EmptyService) RemoveImage(ctx context.Context, imageType options.ImageType) error {
	return nil
}
