package main

import (
	"context"

	"dagger/porto-meetup/internal/dagger"
)

var (
	APP           = "personal-blog"
	GH_REPO       = "https://github.com/NunoFrRibeiro/devops-porto-nov"
	DH_REPO       = "index.docker.io"
	COUNTER_IMAGE = "nunofilribeiro/counterbackend:v0.1.0"
	ADDER_IMAGE   = "nunofilribeiro/adderbackend:v0.1.0"
)

func (m *PortoMeetup) LintAll(
	ctx context.Context,
) (string, error) {
	adderResult, error := m.Lint(ctx, m.Source.Directory("AdderBackend"))
	if error != nil {
		return "", error
	}

	counterResult, error := m.Lint(ctx, m.Source.Directory("CounterBackend"))
	if error != nil {
		return "", error
	}

	result := adderResult + "\n" + counterResult
	return result, nil
}

func (m *PortoMeetup) TestAll(
	ctx context.Context,
) (string, error) {
	adderResult, err := m.UnitTests(ctx, m.Source.Directory("AdderBackend"))
	if err != nil {
		return "", err
	}

	counterResult, err := m.UnitTests(ctx, m.Source.Directory("CounterBackend"))
	if err != nil {
		return "", err
	}

	result := adderResult + "\n" + counterResult
	return result, nil
}

func (m *PortoMeetup) ServeAll(
	ctx context.Context,
	adderPort int,
	counterPort int,
) *dagger.Service {
	adderSource := m.Source.Directory("AdderBackend")
	adderAsService := m.Serve(adderSource, adderPort, "AdderBackend").WithHostname("AdderBackend")

	counterSource := m.Source.Directory("CounterBackend")
	counterAsService := m.Serve(counterSource, counterPort, "CounterBackend").WithHostname("CounterBackend")

	return dag.Proxy().
		WithService(adderAsService, "AdderBackend", adderPort, adderPort, dagger.ProxyWithServiceOpts{
			IsTCP: true,
		}).
		WithService(counterAsService, "CounterBackend", counterPort, counterPort, dagger.ProxyWithServiceOpts{
			IsTCP: true,
		}).
		Service()
}

func (m *PortoMeetup) Deploy(
	ctx context.Context,
	// Infisical Auth Client ID
	// +required
	infisicalId *dagger.Secret,
	// Infisical Auth Client Secret
	// +required
	infisicalSecret *dagger.Secret,
	// Infisical Project to fetch secrets
	// +required
	infisicalProject string,
) (string, error) {
	var result string
	if infisicalId != nil && infisicalProject != "" {
		registryUser, err := dag.Infisical(infisicalId, infisicalSecret).
			GetSecret("DH_USER", infisicalProject, "dev", dagger.InfisicalGetSecretOpts{
				SecretPath: "/flyio",
			}).
			Plaintext(ctx)
		if err != nil {
			return "", err
		}

		registryPass := dag.Infisical(infisicalId, infisicalSecret).
			GetSecret("DH_PASS", infisicalProject, "dev", dagger.InfisicalGetSecretOpts{
				SecretPath: "/",
			})

		counterImage := m.Container(m.Source.Directory("CounterBackend"), 8081, "CounterBackend")
		counterResult, err := dag.Container().
			WithRegistryAuth(DH_REPO, registryUser, registryPass).
			Publish(ctx, COUNTER_IMAGE, dagger.ContainerPublishOpts{
				PlatformVariants: []*dagger.Container{
					counterImage,
				},
			})

		adderImage := m.Container(m.Source.Directory("AdderBackend"), 8080, "AdderBackend")
		adderResult, err := dag.Container().
			WithRegistryAuth(DH_REPO, registryUser, registryPass).
			Publish(ctx, ADDER_IMAGE, dagger.ContainerPublishOpts{
				PlatformVariants: []*dagger.Container{
					adderImage,
				},
			})

		result = counterResult + "\n" + adderResult

		return result, nil
	}

	return result, nil
}
