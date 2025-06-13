package main

import (
	"context"
	"strings"

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

		suggestion, err := dag.LLM(dagger.LLMOpts{
			Model: model,
		}).WithEnv(env).
			WithPromptFile(prompt).
			Env().
			Output("fixed").
			AsWorkspace().
			Diff(ctx)
		if err != nil {
			return "", err
		}
		markupSuggestion := ""
		codeSuggestions := parseDiff(suggestion)
		for _, codeSuggestion := range codeSuggestions {
			markupSuggestion = "```suggestion\n" + strings.Join(
				codeSuggestion.Suggestion,
				"\n",
			) + "\n```"
		}

		return markupSuggestion, nil
	}

	if _, adderErr := d.Buildcnp.CheckDirectory(ctx, d.Source.Directory("AdderBackend")); adderErr != nil {
		workspace := dag.Workspace(
			d.Source.Directory("AdderBackend"),
			d.Buildcnp.AsWorkspaceCheckable(),
		)

		env := dag.Env().
			WithWorkspaceInput("workspace", workspace, "workspace to read, write and test code").
			WithWorkspaceOutput("output", "workspace with fixes")

		suggestion, err := dag.LLM(dagger.LLMOpts{
			Model: model,
		}).WithEnv(env).
			WithPromptFile(prompt).
			Env().
			Output("fixed").
			AsWorkspace().
			Diff(ctx)
		if err != nil {
			return "", err
		}
		markupSuggestion := ""
		codeSuggestions := parseDiff(suggestion)
		for _, codeSuggestion := range codeSuggestions {
			markupSuggestion = "```suggestion\n" + strings.Join(
				codeSuggestion.Suggestion,
				"\n",
			) + "\n```"
		}

		return markupSuggestion, nil
	}

	return "Nothing broken was found", nil
}

func (d *PortoMeetup) DebugTestsPR(
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

	d, err = New(gitSource, "", "", "")
	if err != nil {
		return err
	}

	suggestionDiff, err := d.DebugTests(ctx, model)
	if err != nil {
		return err
	}
	codeSuggestions := parseDiff(suggestionDiff)
	for _, codeSuggestion := range codeSuggestions {
		markupSuggestion := "```suggestion\n" + strings.Join(
			codeSuggestion.Suggestion,
			"\n",
		) + "\n```"
		err := githubIssue.WritePullRequestCodeComment(
			ctx,
			GH_REPO,
			pr,
			commit,
			markupSuggestion,
			codeSuggestion.File,
			"RIGHT",
			codeSuggestion.Line,
		)
		if err != nil {
			return err
		}
	}
	return nil
}
