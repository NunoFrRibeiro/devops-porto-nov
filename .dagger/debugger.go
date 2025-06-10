package main

import (
	"context"

	"dagger/porto-meetup/internal/dagger"
)

// Debug broken tests
func (d *PortoMeetup) DebugTests(
	ctx context.Context,
	// LLM model used to debug tests
	// *optional
	// +default="gemini-2.0-flash"
	model string,
) (string, error) {
	prompt := dag.CurrentModule().
		Source().
		File("prompts/fix_tests.md")

	if _, counterErr := d.Buildcnp.CheckDirectory(ctx, d.Source.Directory("CounterBackend")); counterErr != nil {
		workspace := dag.Workspace(
			d.Source.Directory("CounterBackend"),
			d.Buildcnp.AsWorkspaceCheckable(),
		)

		env := dag.Env().
			WithWorkspaceInput("workspace", workspace, "workspace to read, write and test code").
			WithWorkspaceOutput("output", "workspace with fixes")

		return dag.LLM(dagger.LLMOpts{
			Model: model,
		}).WithEnv(env).
			WithPromptFile(prompt).
			Env().
			Output("fixed").
			AsWorkspace().
			Diff(ctx)
	}

	if _, adderErr := d.Buildcnp.CheckDirectory(ctx, d.Source.Directory("AdderBackend")); adderErr != nil {
		workspace := dag.Workspace(
			d.Source.Directory("AdderBackend"),
			d.Buildcnp.AsWorkspaceCheckable(),
		)

		env := dag.Env().
			WithWorkspaceInput("workspace", workspace, "workspace to read, write and test code").
			WithWorkspaceOutput("output", "workspace with fixes")

		return dag.LLM(dagger.LLMOpts{
			Model: model,
		}).WithEnv(env).
			WithPromptFile(prompt).
			Env().
			Output("fixed").
			AsWorkspace().
			Diff(ctx)
	}

	return "Nothing broken was found", nil
}
