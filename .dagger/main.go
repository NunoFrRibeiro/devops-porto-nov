package main

import (
	"dagger/porto-meetup/internal/dagger"
	"runtime"
)

type PortoMeetup struct {
	// Project Source Directory
	// +private
	Source *dagger.Directory
	// If needed specify the architecture
	// +private
	Arch string
	// If needed specify the OS
	// +private
	OS string
}

func New(
	// Project Source Directory
	// +defaultPath="/"
	// +optional
	// +ignore=[".github", "tmp"]
	source *dagger.Directory,

	// Checkout the repository (at the designated ref) and use it as the source directory instead of the local one.
	// +optional
	ref string,

	// If needed specify the architecture
	// +optional
	arch string,

	// If needed specify the OS
	// +optional
	os string,
) (*PortoMeetup, error) {
	if source == nil && ref != "" {
		source = dag.Git("https://github.com/NunoFrRibeiro/devops-porto-nov.git").
			Ref(ref).
			Tree()
	} else if arch == "" {
		arch = runtime.GOARCH
	} else if os == "" {
		os = runtime.GOOS
	}

	return &PortoMeetup{
		Source: source,
		Arch:   arch,
		OS:     os,
	}, nil
}
