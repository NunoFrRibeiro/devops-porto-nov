package main

import (
	"context"
	"fmt"

	"dagger/porto-meetup/internal/dagger"
)

var (
	APP           = "personal-blog"
	GH_REPO       = "https://github.com/NunoFrRibeiro/devops-porto-nov"
	DH_REPO       = "index.docker.io"
	COUNTER_IMAGE = "nunofilribeiro/counterbackend:v0.1.0"
	ADDER_IMAGE   = "nunofilribeiro/adderbackend:v0.1.0"
	GHCR          = "oci://ghcr.io/nunofrribeiro/devops-porto-nov"
)

// Runs GolangCILint for all sources
func (m *PortoMeetup) Lint(
	ctx context.Context,
) (string, error) {
	adderResult, error := m.Buildcnp.Lint(ctx, m.Source.Directory("AdderBackend"))
	if error != nil {
		return "", error
	}

	counterResult, error := m.Buildcnp.Lint(ctx, m.Source.Directory("CounterBackend"))
	if error != nil {
		return "", error
	}

	result := adderResult + "\n" + counterResult
	return result, nil
}

// Runs all tests
func (m *PortoMeetup) Test(
	ctx context.Context,
) (string, error) {
	adderResult, err := m.Buildcnp.UnitTests(ctx, m.Source.Directory("AdderBackend"))
	if err != nil {
		return "", err
	}

	counterResult, err := m.Buildcnp.UnitTests(ctx, m.Source.Directory("CounterBackend"))
	if err != nil {
		return "", err
	}

	result := adderResult + "\n" + counterResult
	return result, nil
}

func (m *PortoMeetup) Check(
	ctx context.Context,
	// Token with permissions to comment on PR
	githubToken *dagger.Secret,
	// GitHub git commit
	commit string,
	// LLM model used to debug tests
	// *optional
	// +default="gemini-2.0-flash"
	model string,
) (string, error) {
	lintResult, err := m.Lint(ctx)
	if err != nil {
		if githubToken != nil {
			debugPr := m.DebugPR(ctx, githubToken, commit, model)
			return "", fmt.Errorf("failed to lint.\nrunning debugger for %v %v", err, debugPr)
		}
		return "", err
	}

	testResult, err := m.Test(ctx)
	if err != nil {
		if githubToken != nil {
			debugPr := m.DebugPR(ctx, githubToken, commit, model)
			return "", fmt.Errorf("failed to lint.\nrunning debugger for %v %v", err, debugPr)
		}
		return "", err
	}

	return fmt.Sprintf("lint result: %s\ntest result: %s\n", lintResult, testResult), nil
}

// Creates a service to test changes made
func (m *PortoMeetup) ServeAll(
	ctx context.Context,
	adderPort int,
	counterPort int,
) *dagger.Service {
	adderSource := m.Source.Directory("AdderBackend")
	adderAsService := m.Buildcnp.Serve(adderSource, adderPort, "AdderBackend", m.Arch).
		WithHostname("AdderBackend")

	counterSource := m.Source.Directory("CounterBackend")
	counterAsService := m.Buildcnp.Serve(counterSource, counterPort, "CounterBackend", m.Arch).
		WithHostname("CounterBackend")

	return dag.Proxy().
		WithService(adderAsService, "AdderBackend", adderPort, adderPort, dagger.ProxyWithServiceOpts{
			IsTCP: true,
		}).
		WithService(counterAsService, "CounterBackend", counterPort, counterPort, dagger.ProxyWithServiceOpts{
			IsTCP: true,
		}).
		Service()
}

// Deploys the docker images to a registry
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
				SecretPath: "/",
			}).
			Plaintext(ctx)
		if err != nil {
			return "", err
		}

		registryPass := dag.Infisical(infisicalId, infisicalSecret).
			GetSecret("DH_PASS", infisicalProject, "dev", dagger.InfisicalGetSecretOpts{
				SecretPath: "/",
			})

		counterImage := m.Buildcnp.Container(
			m.Source.Directory("CounterBackend"),
			8081,
			"CounterBackend",
		)
		counterResult, err := dag.Container().
			WithRegistryAuth(DH_REPO, registryUser, registryPass).
			Publish(ctx, COUNTER_IMAGE, dagger.ContainerPublishOpts{
				PlatformVariants: []*dagger.Container{
					counterImage,
				},
			})

		adderImage := m.Buildcnp.Container(m.Source.Directory("AdderBackend"), 8080, "AdderBackend")
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

// Deploys the Helm Charts to the registry
func (m *PortoMeetup) DeployCharts(
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
) error {
	if infisicalId != nil && infisicalProject != "" {
		registryUser, err := dag.Infisical(infisicalId, infisicalSecret).
			GetSecret("GH_USER", infisicalProject, "dev", dagger.InfisicalGetSecretOpts{
				SecretPath: "/",
			}).
			Plaintext(ctx)
		if err != nil {
			return err
		}

		registryPass := dag.Infisical(infisicalId, infisicalSecret).
			GetSecret("GH_PASS", infisicalProject, "dev", dagger.InfisicalGetSecretOpts{
				SecretPath: "/",
			})

		counterChartDirectory := m.Source.Directory("helm/charts/counter")
		counterChart := m.Buildcnp.PackageChart(
			counterChartDirectory,
			dagger.BuildcnpPackageChartOpts{
				Version: "0.1.0",
			},
		)
		err = dag.Helm().
			WithRegistryAuth("ghcr.io", registryUser, registryPass).
			Push(ctx, counterChart, GHCR)
		if err != nil {
			return err
		}

		adderChartDirectory := m.Source.Directory("helm/charts/adder")
		adderChart := m.Buildcnp.PackageChart(adderChartDirectory, dagger.BuildcnpPackageChartOpts{
			Version: "0.1.0",
		})
		err = dag.Helm().
			WithRegistryAuth("ghcr.io", registryUser, registryPass).
			Push(ctx, adderChart, GHCR)
		if err != nil {
			return err
		}

	}
	return nil
}

// Tests the Helm charts deployment to a k3s Cluster
func (m *PortoMeetup) TestCharts(
	ctx context.Context,
) (string, error) {
	service, err := m.createCluster().KCDServer.Start(ctx)
	if err != nil {
		return "", err
	}

	kubeConfig := m.getConfig()

	return dag.Container().From("bitnami/kubectl:1.31.0-debian-12-r4").
		WithUser("root").
		WithExec([]string{
			"bash",
			"-c",
			"apt update && apt install -y curl",
		}).
		WithDirectory("/demo", m.Source).
		WithServiceBinding("k3scluster", service).
		WithFile("/.kube/config", kubeConfig).
		WithEnvVariable("KUBECONFIG", "/.kube/config").
		WithExec([]string{"chown", "1001:0", "/.kube/config"}).
		WithExec([]string{
			"bash",
			"-c",
			"curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash",
		}).
		WithExec([]string{
			"bash",
			"-c",
			"helm install --wait --debug adder oci://ghcr.io/nunofrribeiro/devops-porto-nov/adder",
		}).
		WithExec([]string{
			"bash",
			"-c",
			"helm install --wait --debug counter oci://ghcr.io/nunofrribeiro/devops-porto-nov/counter",
		}).Stdout(ctx)
}

func (m *PortoMeetup) createCluster() *PortoMeetup {
	kc := dag.K3S("TestCharts").Container()
	kc = kc.WithMountedCache("/var/lib/dagger", dag.CacheVolume("varlibdagger"))

	m.KCDServer = dag.K3S("TestCharts").WithContainer(kc).Server()
	return m
}

func (m *PortoMeetup) getConfig() *dagger.File {
	return dag.K3S("TestCharts").Config(dagger.K3SConfigOpts{
		Local: false,
	})
}

// Deploys k9s to a already created cluster
func (m *PortoMeetup) KNS() *dagger.Container {
	return dag.K3S("TestCharts").Kns().Terminal()
}
