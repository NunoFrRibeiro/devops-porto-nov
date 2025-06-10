package main

import (
	"context"
	"fmt"
	"runtime"

	"dagger/buildcnp/internal/dagger"
)

type Buildcnp struct {
	Source *dagger.Directory
}

func New(
	source *dagger.Directory,
) *Buildcnp {
	return &Buildcnp{
		Source: source,
	}
}

// Run the projects unit tests
func (m *Buildcnp) UnitTests(
	ctx context.Context,
) (string, error) {
	return dag.Golang().
		WithSource(m.Source).
		Test(ctx)
}

// Runs GolangCILint against the source
func (m *Buildcnp) Lint(
	ctx context.Context,
) (string, error) {
	return dag.Golang().
		WithSource(m.Source).
		GolangciLint(ctx)
}

// Formatter
func (m *Buildcnp) Format() *dagger.Directory {
	return dag.Golang().
		WithSource(m.Source).
		Fmt().
		GolangciLintFix()
}

// Checker
func (m *Buildcnp) Check(
	ctx context.Context,
) (string, error) {
	lint, err := m.Lint(ctx)
	if err != nil {
		return "", err
	}
	test, err := m.UnitTests(ctx)
	if err != nil {
		return "", fmt.Errorf("Error is: %v", err)
	}
	return "Lint result: " + lint + "\n" + "Test result: " + test, nil
}

// Builds the source binary
func (m *Buildcnp) Build(
	source *dagger.Directory,
) *dagger.Directory {
	return dag.Golang().Build([]string{}, dagger.GolangBuildOpts{
		Source: source,
	})
}

// Returns the source binary
func (m *Buildcnp) Binary(
	source *dagger.Directory,
	binaryName string,
) *dagger.File {
	binary := m.Build(source)
	return binary.File(binaryName)
}

// Returns a container with the built binary
func (m *Buildcnp) Container(
	source *dagger.Directory,
	// Port to open on container
	// +required
	port int,
	binaryName string,
	// Architecture to build the container
	// +optional
	arch string,
) *dagger.Container {
	if arch == "" {
		arch = runtime.GOARCH
	}
	binary := m.Binary(source, binaryName)
	binaryStr := fmt.Sprintf("/bin/%s", arch)

	return dag.Container(dagger.ContainerOpts{
		Platform: dagger.Platform(arch),
	}).
		From("ubuntu:24.10").
		WithFile(binaryStr, binary).
		WithEntrypoint([]string{binaryStr}).
		WithExposedPort(port)
}

// Creates a service for a created binary
func (m *Buildcnp) Serve(
	source *dagger.Directory,
	// Port to open on container
	// +required
	port int,
	binaryName string,
	arch string,
) *dagger.Service {
	return m.Container(source, port, binaryName, arch).AsService()
}

// Packages the Helm charts
func (m *Buildcnp) PackageChart(
	// Name of chart to build a package
	chart *dagger.Directory,

	// Set the version of the chart
	// +optional
	version string,
) *dagger.File {
	return dag.Helm().Package(chart, dagger.HelmPackageOpts{
		Version: version,
	})
}

// Stateless checker
func (m *Buildcnp) CheckDirectory(
	ctx context.Context,
	// Directory to run checks on
	source *dagger.Directory,
) (string, error) {
	m.Source = source
	return m.Check(ctx)
}

// Stateless formatter
func (m *Buildcnp) FormatDirectory(
	// Directory to format
	source *dagger.Directory,
) *dagger.Directory {
	m.Source = source
	return m.Format()
}

// Stateless formatter
func (m *Buildcnp) FormatFile(
	// Directory with go module
	source *dagger.Directory,
	// File path to format
	path string,
) *dagger.Directory {
	return dag.
		Container().
		From("golang:1.24").
		WithExec([]string{"go", "install", "golang.org/x/tools/gopls@latest"}).
		WithWorkdir("/app").
		WithDirectory("/app", source).
		WithExec([]string{"gopls", "format", "-w", path}).
		WithExec([]string{"gopls", "imports", "-w", path}).
		Directory("/app")
}
