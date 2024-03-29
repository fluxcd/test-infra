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
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// CreatedAtTimeLayout is a time layout for the 'createdat' label/tag on cloud
// resources.
const CreatedAtTimeLayout = "x2006-01-02_15h04m05s"

// RunCommandOptions is used to configure the RunCommand execution.
type RunCommandOptions struct {
	// Shell is the name of the shell program used to run the command.
	Shell string
	// EnvVars is the environment variables used with the command.
	EnvVars []string
	// StdoutOnly can be enabled to only capture the stdout of the command
	// output.
	StdoutOnly bool
	// Timeout is timeout for the command execution.
	Timeout time.Duration
	// AttachConsole attaches the stdout and stderr of the command to the
	// console.
	AttachConsole bool
}

// defaultRunCommandOptions adds default options of RunCommandOptions.
func defaultRunCommandOptions(o *RunCommandOptions) {
	if o.Shell == "" {
		o.Shell = "bash"
	}
	if o.Timeout == 0 {
		o.Timeout = 5 * time.Minute
	}
}

// RunCommand executes the given command in a given directory.
func RunCommand(ctx context.Context, dir, command string, opts RunCommandOptions) error {
	output, err := RunCommandWithOutput(ctx, dir, command, opts)
	if err != nil {
		return fmt.Errorf("failed to run command %s: %v", string(output), err)
	}
	return nil
}

// RunCommandWithOutput executes the given command and returns the output.
func RunCommandWithOutput(ctx context.Context, dir, command string, opts RunCommandOptions) ([]byte, error) {
	defaultRunCommandOptions(&opts)

	// If ctx has deadline, pass the context to the command, else create a new
	// context with timeout from the run command option.
	var timeoutCtx context.Context
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		timeoutCtx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	} else {
		timeoutCtx = ctx
	}
	cmd := exec.CommandContext(timeoutCtx, opts.Shell, "-c", command)
	cmd.Dir = dir
	// Add env vars.
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, opts.EnvVars...)

	// Create writers to attach to the command.
	outWriters := []io.Writer{}
	errWriters := []io.Writer{}

	// Append the writers with an output buffer to capture the command writes.
	// If attach console is requested, append stdout. Append stderr only if
	// StdoutOnly is not requested.
	var output bytes.Buffer
	outWriters = append(outWriters, &output)
	if !opts.StdoutOnly {
		errWriters = append(errWriters, &output)
	}
	if opts.AttachConsole {
		outWriters = append(outWriters, os.Stdout)
		if !opts.StdoutOnly {
			errWriters = append(errWriters, os.Stderr)
		}
	}

	outWr := io.MultiWriter(outWriters...)
	errWr := io.MultiWriter(errWriters...)

	// Assign writers to the command based on the configuration.
	cmd.Stdout = outWr
	if !opts.StdoutOnly {
		cmd.Stderr = errWr
	}

	err := cmd.Run()
	return output.Bytes(), err
}

// CreateAndPushImages randomly generates test images with the given tags and
// pushes them to the given test repositories.
func CreateAndPushImages(repos map[string]string, tags []string) error {
	// TODO: Build and push concurrently.
	for _, repo := range repos {
		for _, tag := range tags {
			imgRef := repo + ":" + tag
			ref, err := name.ParseReference(imgRef)
			if err != nil {
				return err
			}

			// Use the login credentials from the host docker/podman client config.
			opts := []remote.Option{
				remote.WithAuthFromKeychain(authn.DefaultKeychain),
			}

			// Create a random image.
			img, err := random.Image(1024, 1)
			if err != nil {
				return err
			}

			log.Printf("pushing test image %s\n", ref.String())
			if err := remote.Write(ref, img, opts...); err != nil {
				return err
			}
		}
	}
	return nil
}

// RetagAndPush retags local image based on the remoteImage and pushes the remoteImage
func RetagAndPush(ctx context.Context, localImage, remoteImage string) error {
	log.Printf("pushing flux test image %s\n", remoteImage)
	// Retag local image and push.
	if err := RunCommand(ctx, "./",
		fmt.Sprintf("docker tag %s %s", localImage, remoteImage),
		RunCommandOptions{},
	); err != nil {
		return err
	}

	return RunCommand(ctx, "./",
		fmt.Sprintf("docker push %s", remoteImage),
		RunCommandOptions{},
	)
}

// ParseCreatedAtTime parses 'createdat' label/tag on resources. The time value
// is in a custom format due to the label/tag value restrictions on various
// cloud platforms. See tf-modules/utils/tags for details about the custom
// format.
func ParseCreatedAtTime(createdat string) (time.Time, error) {
	return time.Parse(CreatedAtTimeLayout, createdat)
}
