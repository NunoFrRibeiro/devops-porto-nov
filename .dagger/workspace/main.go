package main

import (
	"context"

	"dagger/workspace/internal/dagger"
)

type Workspace struct {
	Work  *dagger.Directory
	Start *dagger.Directory
}

func (w *Workspace) Read(
	ctx context.Context,
	// Path to write the file
	path string,
) (string, error) {
	return w.Work.File(path).Contents(ctx)
}

func (w *Workspace) Write()
