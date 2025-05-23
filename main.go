package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"
)

type Locals struct {
	AccountName string                 `hcl:"account_name,attr"`
	Remain      map[string]interface{} `hcl:",remain"`
}

type AccountConfig struct {
	Locals Locals `hcl:"locals,block"`
}

type Options struct {
	Debug                bool   `short:"d" long:"debug" description:"Enable debug output"`
	NonInteractive       bool   `short:"n" long:"non-interactive" description:"Disable Terragrunt interactive mode"`
	DisableProviderCache bool   `short:"p" long:"disable-provider-cache" description:"Disable provider cache"`
	AccountFile          string `short:"a" long:"account" description:"Account file" default:"account.hcl"`
	Cmd                  struct {
		Args []string `positional-arg-name:"CMD" required:"yes"`
	} `positional-args:"yes" required:"yes"`
}

var opts Options

func main() {
	_, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(1)
	}

	if opts.Debug {
		log.SetLevel(log.DebugLevel)
		log.SetFormatter(&log.TextFormatter{
			ForceColors:   true,
			FullTimestamp: false,
		})
	}

	if len(opts.Cmd.Args) == 0 {
		log.Errorf("Usage: terra-profile <command> [args...]")
		os.Exit(1)
	}

	cwd, err := os.Getwd()
	if err != nil {
		log.Errorf("Could not get current directory: %v\n", err)
		os.Exit(1)
	}

	hclPath, err := findAccountHCLFile(cwd)
	if err != nil {
		log.Errorf("Error: %v\n", err)
		os.Exit(1)
	}

	accountName, err := getAccountNameFromHCLFile(hclPath)
	if err != nil {
		log.Errorf("Error: %v\n", err)
		os.Exit(1)
	}

	if opts.Debug {
		log.Debugf("Found %s file in: %s\n", opts.AccountFile, filepath.Dir(hclPath))
		log.Debugf("HCL file: account_name=%s\n", accountName)
	}

	err = os.Setenv("AWS_PROFILE", accountName)
	if err != nil {
		log.Errorf("Error setting environment variable: %v\n", err)
		return
	}

	if opts.Debug {
		log.Debugf("Setting environment variable: AWS_PROFILE=%s\n", accountName)
	}

	if opts.NonInteractive {
		err = os.Setenv("TG_NON_INTERACTIVE", "true")
		if err != nil {
			log.Errorf("Error setting environment variable: %v\n", err)
			return
		}

		if opts.Debug {
			log.Debug("Setting environment variable: TG_NON_INTERACTIVE=true\n")
		}
	}

	if !opts.DisableProviderCache {
		err = os.Setenv("TG_PROVIDER_CACHE", "true")
		if err != nil {
			log.Errorf("Error setting environment variable: %v\n", err)
			return
		}

		if opts.Debug {
			log.Debug("Setting environment variable: TG_PROVIDER_CACHE=true\n")
		}
	}

	cmd := exec.Command(opts.Cmd.Args[0], opts.Cmd.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = os.Environ()

	if opts.Debug {
		log.Debugf("Executing: %v\n", cmd.Args)
	}

	err = cmd.Run()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.ExitCode())
		}

		log.Errorf("Error running command: %v\n", err)
		os.Exit(1)
	}
}

func findGitRoot(startDir string) (string, error) {
	dir := startDir

	for {
		gitRootDir := filepath.Join(dir, ".git")

		info, err := os.Stat(gitRootDir)
		if err == nil && info.IsDir() {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf(".git directory not found")
		}

		dir = parent
	}
}

func findAccountHCLFile(startDir string) (string, error) {
	dir := startDir

	gitRootDir, err := findGitRoot(dir)
	if err != nil {
		fmt.Println("Error:", err)
		return "", nil
	}

	for {
		path := filepath.Join(dir, opts.AccountFile)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}

		parent := filepath.Dir(dir)

		if parent == gitRootDir {
			log.Debugf("Breaking at Git root directory: %s", gitRootDir)
			break
		}

		if parent == dir {
			break
		}

		dir = parent
	}

	return "", fmt.Errorf("%s not found in any parent directory", opts.AccountFile)
}

func getAccountNameFromHCLFile(hclPath string) (string, error) {
	var config AccountConfig

	err := hclsimple.DecodeFile(hclPath, nil, &config)
	if err != nil {
		return "", fmt.Errorf("could not parse %s: %w", hclPath, err)
	}

	accountName := config.Locals.AccountName

	return accountName, nil
}
