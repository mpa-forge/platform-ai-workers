package shell

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
)

type Runner struct{}

type Result struct {
	Stdout string
	Stderr string
}

func (Runner) Run(ctx context.Context, dir string, env []string, name string, args ...string) (Result, error) {
	return Runner{}.RunInput(ctx, dir, env, nil, name, args...)
}

func (Runner) RunInput(ctx context.Context, dir string, env []string, stdin io.Reader, name string, args ...string) (Result, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	if len(env) > 0 {
		cmd.Env = env
	}
	if stdin != nil {
		cmd.Stdin = stdin
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return Result{
			Stdout: stdout.String(),
			Stderr: stderr.String(),
		}, fmt.Errorf("%s %v: %w; stderr=%s", name, args, err, stderr.String())
	}

	return Result{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}, nil
}
