package maws

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/pkg/errors"
	"golang.org/x/sync/semaphore"
)

func readRoleFile(name string) ([]string, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}

	var roles []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") {
			// comment
			continue
		}
		an, err := arn.Parse(line)
		if err != nil {
			log.Println("[warn] skip. invalid ARN:", line, err)
			continue
		}
		if an.Service != "iam" || !strings.HasPrefix(an.Resource, "role/") {
			log.Println("[warn] skip. not a IAM role ARN:", an.String())
			continue
		}
		roles = append(roles, an.String())
	}
	return roles, scanner.Err()
}

type Option struct {
	RoleFile     string
	MaxParallels int64
	Commands     []string
}

func Run(ctx context.Context, opt Option) (int64, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return 0, err
	}

	roles, err := readRoleFile(opt.RoleFile)
	if err != nil {
		return 0, err
	}
	var wg sync.WaitGroup
	var errCount int64
	sem := semaphore.NewWeighted(opt.MaxParallels)
	for _, role := range roles {
		wg.Add(1)
		sem.Acquire(ctx, 1)
		go func(role string) {
			defer wg.Done()
			defer sem.Release(1)
			creds, err := assumeRole(ctx, cfg, role)
			if err != nil {
				log.Printf("[warn] failed to assume role for %s %s", role, err)
				atomic.AddInt64(&errCount, 1)
				return
			}
			if err := runFor(ctx, cfg.Region, creds, opt.Commands); err != nil {
				log.Printf("[warn] failed to run for role %s %s", role, err)
				atomic.AddInt64(&errCount, 1)
				return
			}
		}(role)
	}
	wg.Wait()
	return errCount, nil
}

var restrictCommandPrefix = []string{
	"list-",
	"get-",
	"describe-",
}

func restrictCommand(commands []string) error {
	if len(commands) < 2 {
		return errors.New("insufficient commands")
	}

	// example: maws -- sts get-caller-identity
	// commands[0]: sts
	// commands[1]: get-caller-identity
	for _, prefix := range restrictCommandPrefix {
		if strings.HasPrefix(commands[1], prefix) {
			return nil
		}
	}
	return errors.Errorf(
		"%s %s is restricted. allowed command prefixes are %v",
		commands[0],
		commands[1],
		restrictCommandPrefix,
	)
}

func runFor(ctx context.Context, region string, creds *types.Credentials, commands []string) error {
	log.Printf("[debug] %s %s", *creds.AccessKeyId, commands)
	if err := restrictCommand(commands); err != nil {
		return err
	}

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

func assumeRole(ctx context.Context, cfg aws.Config, role string) (*types.Credentials, error) {
	log.Printf("[debug] assume role to %s", role)
	client := sts.NewFromConfig(cfg)
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
