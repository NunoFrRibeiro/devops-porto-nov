package main

import (
	"context"

	"dagger/porto-meetup/internal/dagger"
)

type Kube struct {
	// +private
	K3s *dagger.K3S
}

func (k *Kube) Service(
	ctx context.Context,
) (*dagger.Service, error) {
	server := k.K3s.Server()

	server, err := server.Start(ctx)
	if err != nil {
		return nil, err
	}

	err = k.Deploy(ctx, k.K3s.Config(dagger.K3SConfigOpts{
		Local: false,
	}))

	return dag.Proxy().
		WithService(server, "adder", 8081, 30000, dagger.ProxyWithServiceOpts{
			IsTCP: true,
		}).
		WithService(server, "counter", 8080, 30001, dagger.ProxyWithServiceOpts{
			IsTCP: true,
		}).Service(), nil
}

func (k *Kube) Deploy(
	ctx context.Context,
	kubeConfig *dagger.File,
) error {
	_, err := dag.Container().From("bitnami/kubectl:1.31.0-debian-12-r4").
		WithUser("root").
		WithFile("/.kube/config", kubeConfig).
		WithEnvVariable("KUBECONFIG", "/.kube/config").
		WithExec([]string{"chown", "1001:0", "/.kube/config"}).
		WithExec([]string{
			"bash",
			"-c",
			"apt update && apt install -y curl",
		}).
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
		}).Sync(ctx)
	return err
}
