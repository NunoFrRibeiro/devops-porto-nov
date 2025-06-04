package main

import (
	"context"

	"dagger/workspace/internal/dagger"
)

type Checkable interface {
	dagger.DaggerObject
	CheckDirectory(ctx context.Context, source *dagger.Directory) (string, error)
	FormatFile(source *dagger.Directory, filePath string) *dagger.Directory
}

// Place to do the work and check
type Workspace struct {
	Work *dagger.Directory
	// +private
	Start *dagger.Directory
	// +private
	Checker Checkable
}

func New(
	// Initial state of the workspace
	source *dagger.Directory,
	// Checker used for testing
	checker Checkable,
) *Workspace {
	return &Workspace{
		Start:   source,
		Work:    source,
		Checker: checker,
	}
}

// Read the contents of the of the workspace at a given source
func (w *Workspace) Read(
	ctx context.Context,
	// Path to write the file
	path string,
) (string, error) {
	return w.Work.File(path).Contents(ctx)
}

// Write the contents of a file in the workspace at a given source
func (w *Workspace) Write(
	ctx context.Context,
	// Path to write the file
	path string,
	// Contents to write to the file
	contents string,
) *Workspace {
	w.Work = w.Work.WithNewFile(path, contents)
	w.Work = w.Checker.FormatFile(w.Work, path)
	return w
}

// Reset the workspace to the original state
func (w *Workspace) Reset() *Workspace {
	w.Work = w.Start
	return w
}

// Run tests in the workspace
func (w *Workspace) Check(ctx context.Context) (string, error) {
	return w.Checker.CheckDirectory(ctx, w.Work)
}

// Display the diff made in the Workspace
func (w *Workspace) Diff(ctx context.Context) (string, error) {
	return dag.Container().From("alpine").
		WithDirectory("/start", w.Start).
		WithDirectory("/finish", w.Work).
		WithExec([]string{
			"diff",
			"-rN",
			"/start",
			"/finish",
		}, dagger.ContainerWithExecOpts{
			Expect: dagger.ReturnTypeAny,
		}).Stdout(ctx)
}
