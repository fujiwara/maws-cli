package maws

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/aws-sdk-go-v2/service/sts/types"
	"golang.org/x/sync/semaphore"
)

type Option struct {
	Config       string
	MaxParallels int64
	Commands     []string
}

func Run(ctx context.Context, opt Option) (int64, error) {

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return 0, err
	}

	cfg, err := LoadConfig(opt.Config)
	if err != nil {
		return 0, err
	}
	if err := cfg.restrictCommand(opt.Commands); err != nil {
		return 0, err
	}

	var wg sync.WaitGroup
	var errCount int64
	sem := semaphore.NewWeighted(opt.MaxParallels)
	for _, role := range cfg.Roles {
		wg.Add(1)
		sem.Acquire(ctx, 1)
		go func(role string) {
			defer wg.Done()
			defer sem.Release(1)
			creds, err := assumeRole(ctx, awsCfg, role)
			if err != nil {
				log.Printf("[warn] failed to assume role for %s %s", role, err)
				atomic.AddInt64(&errCount, 1)
				return
			}
			if err := runFor(ctx, awsCfg.Region, creds, opt.Commands); err != nil {
				log.Printf("[warn] failed to run for role %s %s", role, err)
				atomic.AddInt64(&errCount, 1)
				return
			}
		}(role)
	}
	wg.Wait()
	return errCount, nil
}

func runFor(ctx context.Context, region string, creds *types.Credentials, commands []string) error {
	log.Printf("[debug] %s %s", *creds.AccessKeyId, commands)

	cmd := exec.CommandContext(ctx, "aws", commands...)
	envs := make([]string, 0)
	for _, env := range os.Environ() {
		switch env {
		case "AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_SESSION_TOKEN", "AWS_PROFILE", "AS_DEFAULT_PROFILE", "AWS_PAGER":
		default:
			envs = append(envs, env)
		}
	}
	envs = append(envs,
		"AWS_DEFAULT_REGION="+region,
		"AWS_ACCESS_KEY_ID="+*creds.AccessKeyId,
		"AWS_SECRET_ACCESS_KEY="+*creds.SecretAccessKey,
		"AWS_SESSION_TOKEN="+*creds.SessionToken,
		"AWS_PAGER=",
	)
	cmd.Env = envs
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Wait()
}

func assumeRole(ctx context.Context, awsCfg aws.Config, role string) (*types.Credentials, error) {
	log.Printf("[debug] assume role to %s", role)
	client := sts.NewFromConfig(awsCfg)
	input := &sts.AssumeRoleInput{
		RoleArn:         aws.String(role),
		RoleSessionName: aws.String(fmt.Sprintf("maws-%d", time.Now().Unix())),
	}
	result, err := client.AssumeRole(ctx, input)
	if err != nil {
		return nil, err
	}
	return result.Credentials, nil
}
