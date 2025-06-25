package main

import (
	"context"
	"fmt"
	"strings"

	"dagger/porto-meetup/internal/dagger"
)

// Debug broken tests
func (d *PortoMeetup) DebugLocal(
	ctx context.Context,
	// LLM model used to debug tests
	// *optional
	// +default="gemini-2.0-flash"
	model string,
) (string, error) {
	prompt := dag.CurrentModule().
		Source().
		File("prompts/fix_tests.md")

	// check if CounterBackend is broken
	if _, counterErr := d.Buildcnp.CheckDirectory(ctx, d.Source.Directory("CounterBackend")); counterErr != nil {
		workspace := dag.Workspace(
			d.Source.Directory("CounterBackend"),
			d.Buildcnp.AsWorkspaceCheckable(),
		)

		env := dag.Env().
			WithWorkspaceInput("workspace", workspace, "workspace to read, write and test the CounterBackend code").
			WithWorkspaceOutput("fixed", "workspace with fixes")

		return dag.LLM(dagger.LLMOpts{
			Model: model,
		}).WithEnv(env).
			WithPromptFile(prompt).
			Env().
			Output("fixed").
			AsWorkspace().
			Diff(ctx)
	}

	// check if AdderBackend is broken
	if _, adderErr := d.Buildcnp.CheckDirectory(ctx, d.Source.Directory("AdderBackend")); adderErr != nil {
		workspace := dag.Workspace(
			d.Source.Directory("AdderBackend"),
			d.Buildcnp.AsWorkspaceCheckable(),
		)

		env := dag.Env().
			WithWorkspaceInput("workspace", workspace, "workspace to read, write and test the AdderBackend code").
			WithWorkspaceOutput("fixed", "workspace with fixes")

		return dag.LLM(dagger.LLMOpts{
			Model: model,
		}).WithEnv(env).
			WithPromptFile(prompt).
			Env().
			Output("fixed").
			AsWorkspace().
			Diff(ctx)
	}

	return "", fmt.Errorf("Nothing broken was found")
}

func (d *PortoMeetup) DebugPR(
	ctx context.Context,
	// Token with permissions to comment on PR
	githubToken *dagger.Secret,
	// GitHub git commit
	commit string,
	// LLM model used to debug tests
	// *optional
	// +default="gemini-2.0-flash"
	model string,
) error {
	githubIssue := dag.GithubIssue(dagger.GithubIssueOpts{
		Token: githubToken,
	})
	gitRef := dag.Git(GH_REPO).Commit(commit)
	gitSource := gitRef.Tree()
	pr, err := githubIssue.GetPrForCommit(ctx, GH_REPO, commit)
	if err != nil {
		return err
	}

	d, err = New(gitSource, GH_REPO, "", "")
	if err != nil {
		return err
	}

	suggestionDiff, err := d.DebugLocal(ctx, model)
	if err != nil {
		return err
	}

	if suggestionDiff == "" {
		return fmt.Errorf("no suggestions found")
	}
	codeSuggestions := parseDiff(suggestionDiff)

	for _, codeSuggestion := range codeSuggestions {
		markupSuggestion := "```suggestion\n" + strings.Join(
			codeSuggestion.Suggestion,
			"\n",
		) + "\n```"
		err := githubIssue.WriteComment(ctx, GH_REPO, pr, markupSuggestion)
		// err := githubIssue.WritePullRequestCodeComment(
		// 	ctx,
		// 	GH_REPO,
		// 	pr,
		// 	commit,
		// 	markupSuggestion,
		// 	codeSuggestion.File,
		// 	"RIGHT",
		// 	codeSuggestion.Line,
		// )
		if err != nil {
			return err
		}
	}
	return nil
}
