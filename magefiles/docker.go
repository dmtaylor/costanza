//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/magefile/mage/mg"
)

const prodEnv = "prod"
const devEnv = "dev"
const prodDbDir = "/var/costanza/db_data"
const devDbDir = "/var/costanza_dev/db_data"

type Docker mg.Namespace

// Build builds the docker-compose image for the app with the given environment
func (Docker) Build(env string) error {
	if env == "" {
		env = devEnv
	}
	if env != prodEnv && env != devEnv {
		return fmt.Errorf("invalid environment: %s, only \"prod\" and \"dev\" are valid choices")
	}
	mg.Deps(Tests)
	dockerAppName := getDockerAppName(env)
	fmt.Println("tests passed: building")
	cmd := exec.Command("docker", "compose", "-f", "docker-compose.yml", "-f", "docker-compose."+env+".yml", "-p", dockerAppName, "build")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Run brings up docker compose application for given environment in foreground or background
func (Docker) Run(env string, background bool) error {
	if env == "" {
		env = devEnv
	}
	if env != prodEnv && env != devEnv {
		return fmt.Errorf("invalid environment: %s, only \"prod\" and \"dev\" are valid choices")
	}
	mg.Deps(Tests, mg.F(dbDir, env))
	dockerAppName := getDockerAppName(env)

	var cmd *exec.Cmd
	if background {
		cmd = exec.Command("docker", "compose", "-f", "docker-compose.yml", "-f", "docker-compose."+env+".yml", "-p", dockerAppName, "up", "-d")
	} else {
		cmd = exec.Command("docker", "compose", "-f", "docker-compose.yml", "-f", "docker-compose."+env+".yml", "-p", dockerAppName, "up")
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Restart force rebuilds & runs docker compose application with given env in foreground or background
func (Docker) Restart(env string, background bool) error {
	if env == "" {
		env = devEnv
	}
	if env != prodEnv && env != devEnv {
		return fmt.Errorf("invalid environment: %s, only \"prod\" and \"dev\" are valid choices")
	}
	mg.Deps(Tests, mg.F(dbDir, env))
	dockerAppName := getDockerAppName(env)

	cmd := exec.Command("docker", "compose", "-f", "docker-compose.yml", "-f", "docker-compose."+env+".yml", "-p", dockerAppName, "--", "build", "--no-cache")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to rebuild images: %w", err)
	}
	if background {
		cmd = exec.Command("docker", "compose", "-f", "docker-compose.yml", "-f", "docker-compose."+env+".yml", "-p", dockerAppName, "--", "up", "--build", "--force-recreate", "--no-deps", "-d")
	} else {
		cmd = exec.Command("docker", "compose", "-f", "docker-compose.yml", "-f", "docker-compose."+env+".yml", "-p", dockerAppName, "--", "up", "--build", "--force-recreate", "--no-deps")
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (Docker) Status(env string) error {
	if env == "" {
		env = devEnv
	}
	if env != prodEnv && env != devEnv {
		return fmt.Errorf("invalid environment: %s, only \"prod\" and \"dev\" are valid choices")
	}
	dockerAppName := getDockerAppName(env)
	cmd := exec.Command("docker", "compose", "-f", "docker-compose.yml", "-f", "docker-compose."+env+".yml", "-p", dockerAppName, "--", "ps")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func dbDir(env string) error {
	var dirname string
	if env == devEnv {
		dirname = devDbDir
	} else if env == prodEnv {
		dirname = prodDbDir
	} else {
		return fmt.Errorf("invalid environment: %s", env)
	}
	_, err := os.Stat(dirname)
	if os.IsNotExist(err) {
		cmd := exec.Command("sudo", "mkdir", "--parents", "--verbose", dirname)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	if err != nil {
		return fmt.Errorf("failed to get stat dir %s: %w", dirname, err)
	}
	return nil
}

func getDockerAppName(env string) string {
	if env == devEnv {
		return "costanza-dev"
	} else {
		return "costanza"
	}

}
