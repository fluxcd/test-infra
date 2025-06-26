/*
Copyright 2022 The Flux authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tftestenv

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	install "github.com/hashicorp/hc-install"
	"github.com/hashicorp/hc-install/fs"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/hc-install/src"
	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2/klogr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	runtimeLog "sigs.k8s.io/controller-runtime/pkg/log"
)

var shutdownSignals = []os.Signal{os.Interrupt, syscall.SIGTERM}

// Environment encapsulates a Kubernetes test environment.
type Environment struct {
	client.Client
	ClientGo *kubernetes.Clientset
	Config   *rest.Config

	// CreateKubeconfig provides the terraform state output which is used to
	// construct kubeconfig.
	CreateKubeconfig CreateKubeconfig

	tf       *tfexec.Terraform
	retain   bool
	existing bool
	verbose  bool
	buildDir string
	// tfApplyOptions are the terraform apply options to use when running
	// terraform apply.
	tfApplyOptions []tfexec.ApplyOption
	// tfDestroyOptions are the terraform destroy options to use when running
	// terraform destroy.
	tfDestroyOptions []tfexec.DestroyOption
}

// createKubeconfig create a kubeconfig for the target cluster and writes to
// the given path using the contextual values from the infrastructure state.
type CreateKubeconfig func(ctx context.Context, state map[string]*tfjson.StateOutput, kcPath string) error

// EnvironmentOption is used to configure the Environment.
type EnvironmentOption func(*Environment)

// WithRetain configures the Environment to retain the created or existing
// infrastructure.
func WithRetain(retain bool) EnvironmentOption {
	return func(e *Environment) {
		e.retain = retain
	}
}

// WithExisting configures the Environment to use the existing infrastructure.
// By default, the environment set up would fail if the terraform state is not
// clean.
func WithExisting(existing bool) EnvironmentOption {
	return func(e *Environment) {
		e.existing = existing
	}
}

// WithVerbose configures the terraform executor to run in verbose mode.
func WithVerbose(verbose bool) EnvironmentOption {
	return func(e *Environment) {
		e.verbose = verbose
	}
}

// WithCreateKubeconfig configures how kubeconfig is constructured using the
// output state of the terraform infrastructure.
func WithCreateKubeconfig(create CreateKubeconfig) EnvironmentOption {
	return func(e *Environment) {
		e.CreateKubeconfig = create
	}
}

// WithBuildDir sets the build directory for the environment. Defaults to
// "build".
func WithBuildDir(dir string) EnvironmentOption {
	return func(e *Environment) {
		e.buildDir = dir
	}
}

// WithTfApplyOptions configures terraform apply options.
func WithTfApplyOptions(opts ...tfexec.ApplyOption) EnvironmentOption {
	return func(e *Environment) {
		e.tfApplyOptions = append(e.tfApplyOptions, opts...)
	}
}

// WithTfDestroyOptions configures terraform destroy options.
func WithTfDestroyOptions(opts ...tfexec.DestroyOption) EnvironmentOption {
	return func(e *Environment) {
		e.tfDestroyOptions = append(e.tfDestroyOptions, opts...)
	}
}

// New finds or downloads terraform binary, uses it to run terraform in the
// given terraformPath to create a kubernetes cluster. A kubeconfig of the
// created cluster is constructed at the given kubeconfigPath which is then used
// to construct a kubernetes client that can be used in the tests.
func New(ctx context.Context, scheme *runtime.Scheme, terraformPath string, kubeconfigPath string, opts ...EnvironmentOption) (*Environment, error) {
	// Set a default logger if not set already.
	runtimeLog.SetLogger(klogr.New())

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	env := &Environment{
		buildDir: "build", // Default build dir.
	}

	// Process the options.
	for _, opt := range opts {
		opt(env)
	}

	// Prepare build environment.
	cwd, err := os.Getwd()
	if err != nil {
		return env, fmt.Errorf("failed to get the current working directory: %w", err)
	}
	buildDir := filepath.Join(cwd, env.buildDir)
	if err := os.MkdirAll(buildDir, os.ModePerm); err != nil {
		return env, fmt.Errorf("failed to create build directory: %w", err)
	}

	env.tf, err = setUpTerraform(ctx, terraformPath, buildDir)
	if err != nil {
		return env, fmt.Errorf("could not create terraform instance: %w", err)
	}

	if env.verbose {
		env.tf.SetStdout(os.Stdout)
		env.tf.SetStderr(os.Stderr)
	}

	log.Println("Init Terraform")
	err = env.tf.Init(ctx, tfexec.Upgrade(true))
	if err != nil {
		return env, fmt.Errorf("error running init: %w", err)
	}

	// Exit the test when existing state is found if -existing flag is false.
	if !env.existing {
		log.Println("Checking for an empty Terraform state")
		state, err := env.tf.Show(ctx)
		if err != nil {
			return env, fmt.Errorf("could not read state: %v", err)
		}
		if state.Values != nil {
			log.Println("Found existing resources, likely from previous unsuccessful run, cleaning up...")
			return env, fmt.Errorf("expected an empty state but got existing resources")
		}
	}

	// Set up signal handling to gracefully stop the environment.
	sigs := make(chan os.Signal, 2)
	signal.Notify(sigs, shutdownSignals...)
	go func() {
		// Cancel the resource provisioning on first signal.
		s := <-sigs
		log.Println("Received signal:", s)
		infoMsg := "Attempting to gracefully stop terraform"
		if !env.retain {
			infoMsg += " and clean up"
		}
		log.Println(infoMsg)
		cancel()

		// Exit on second signal.
		<-sigs
		log.Println("Force stop")
		os.Exit(1)
	}()

	if err := env.createAndConfigure(ctx, scheme, kubeconfigPath); err != nil {
		// Clean up the partially provisioned resources on failure based on the
		// environment configuation. In CI, this would ensure that if the CI job
		// is cancelled, the resources get cleaned up.
		err = errors.Join(err, env.Stop(context.Background()))
		return env, fmt.Errorf("error running apply: %v", err)
	}

	return env, nil
}

// setUpTerraform finds or downloads terraform binary and returns Terraform
// which can be used to run terraform operations.
func setUpTerraform(ctx context.Context, terraformPath string, buildDir string) (*tfexec.Terraform, error) {
	// Find or download terraform binary.
	i := install.NewInstaller()
	execPath, err := i.Ensure(ctx, []src.Source{
		&fs.AnyVersion{
			Product: &product.Terraform,
		},
		&releases.LatestVersion{
			Product:    product.Terraform,
			InstallDir: buildDir,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("terraform exec path not found: %w", err)
	}
	log.Println("Terraform binary: ", execPath)

	return tfexec.NewTerraform(terraformPath, execPath)
}

// createAndConfigure creates the resources and configures the Environment with
// the created resource.
func (env *Environment) createAndConfigure(ctx context.Context, scheme *runtime.Scheme, kubeconfigPath string) error {
	// Apply Terraform, read the output values and construct kubeconfig.
	log.Println("Applying Terraform")
	err := env.tf.Apply(ctx, env.tfApplyOptions...)
	if err != nil {
		return fmt.Errorf("error running apply: %v", err)
	}
	state, err := env.tf.Show(ctx)
	if err != nil {
		return fmt.Errorf("could not read state: %v", err)
	}
	outputs := state.Values.Outputs
	if err = env.CreateKubeconfig(ctx, outputs, kubeconfigPath); err != nil {
		return fmt.Errorf("failed to create kubeconfig: %w", err)
	}

	// Create kube client.
	kubeCfg, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return fmt.Errorf("failed to build rest config: %w", err)
	}
	env.Client, err = client.New(kubeCfg, client.Options{Scheme: scheme})
	if err != nil {
		return fmt.Errorf("failed to create new client: %w", err)
	}
	env.ClientGo, err = kubernetes.NewForConfig(kubeCfg)
	if err != nil {
		return fmt.Errorf("failed to create client-go client: %w", err)
	}

	return nil
}

// Stop tears down the test infrastructure created by the environment.
func (env *Environment) Stop(ctx context.Context) error {
	if !env.retain {
		log.Println("Destroying environment...")
		if ferr := env.tf.Destroy(ctx, env.tfDestroyOptions...); ferr != nil {
			return fmt.Errorf("could not destroy infrastructure: %w", ferr)
		}
	}
	return nil
}

// State queries and returns the current state output of terraform.
func (env *Environment) StateOutput(ctx context.Context) (map[string]*tfjson.StateOutput, error) {
	state, err := env.tf.Show(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not read state: %v", err)
	}
	return state.Values.Outputs, nil
}

// Destroy configures a new Environment with the given configurations for
// terraform and runs terraform destroy. Ideally, this need not be used as the
// testenv New() handles graceful cleanup when shutdown signals are received.
// But in case the whole process gets terminated, use this to just perform a
// destroy of any created infrastructure.
// This can be used as the last step in CI to always run irrespective of success
// or failure of the test run to make sure the test infrastructure is destroyed.
// One such scenario is when the cloud provider takes longer than the usual time
// to provision the infrastructure and the test binary execution reaches timeout
// and the whole process gets terminated. This can be run in a separate step in
// CI to destroy the infrastructure.
func Destroy(ctx context.Context, terraformPath string, opts ...EnvironmentOption) error {
	// Set a default logger if not set already.
	runtimeLog.SetLogger(klogr.New())

	env := &Environment{
		buildDir: "build", // Default build dir.
	}

	// Process the options.
	for _, opt := range opts {
		opt(env)
	}

	// Assume that the initial test run created the build directory.
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get the current working directory: %w", err)
	}
	buildDir := filepath.Join(cwd, env.buildDir)

	env.tf, err = setUpTerraform(ctx, terraformPath, buildDir)
	if err != nil {
		return fmt.Errorf("could not create terraform instance: %w", err)
	}

	if env.verbose {
		env.tf.SetStdout(os.Stdout)
		env.tf.SetStderr(os.Stderr)
	}

	log.Println("Terraform destroy...")
	return env.tf.Destroy(ctx, env.tfDestroyOptions...)
}
