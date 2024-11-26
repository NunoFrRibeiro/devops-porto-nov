package main

import (
	"context"
	"dagger/porto-meetup/internal/dagger"
	"fmt"
)

// Run the projects unit tests
func (m *PortoMeetup) UnitTests(
	ctx context.Context,
	source *dagger.Directory,
) (string, error) {
	return dag.Golang().Test(ctx, dagger.GolangTestOpts{
		Source: source,
		Args:   []string{"./..."},
	})
}

// Runs GolangCILint against the source
func (m *PortoMeetup) Lint(
	ctx context.Context,
	source *dagger.Directory,
) (string, error) {
	return dag.Golang().
		WithProject(source).
		GolangciLint(ctx, dagger.GolangGolangciLintOpts{})
}

// Builds the source binary
func (m *PortoMeetup) Build(
	source *dagger.Directory,
) *dagger.Directory {
	return dag.Golang().Build([]string{}, dagger.GolangBuildOpts{
		Source: source,
	})
}

// Returns the source binary
func (m *PortoMeetup) Binary(
	source *dagger.Directory,
	binaryName string,
) *dagger.File {
	binary := m.Build(source)
	return binary.File(binaryName)
}

// Returns a container with the built binary
func (m *PortoMeetup) Container(
	source *dagger.Directory,
	// Port to open on container
	// +required
	port int,
	binaryName string,
) *dagger.Container {
	binary := m.Binary(source, binaryName)
	platform := m.Arch
	binaryStr := fmt.Sprintf("/bin/%s", binaryName)

	return dag.Container(dagger.ContainerOpts{
		Platform: dagger.Platform(platform),
	}).
		From("ubuntu:24.10").
		WithFile(binaryStr, binary).
		WithEntrypoint([]string{binaryStr}).
		WithExposedPort(port)
}

func (m *PortoMeetup) Serve(
	source *dagger.Directory,
	// Port to open on container
	// +required
	port int,
	binaryName string,
) *dagger.Service {
	return m.Container(source, port, binaryName).AsService()
}
