package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/github"
	"github.com/urfave/cli/v2"
)

func main() {

	app := &cli.App{
		Name:  "drone-github-commit-status",
		Usage: "sends a commit status to github via drone",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "github_app_id",
				Usage:    "github app id",
				EnvVars:  []string{"PLUGIN_GITHUB_APP_ID"},
				Required: true,
			},
			&cli.StringFlag{
				Name:     "github_installation_id",
				Usage:    "github installation id",
				EnvVars:  []string{"PLUGIN_GITHUB_INSTALLATION_ID"},
				Required: true,
			},
			&cli.StringFlag{
				Name:     "github_app_private_key",
				Usage:    "github private key string",
				EnvVars:  []string{"PLUGIN_GITHUB_APP_PRIVATE_KEY"},
				Required: true,
			},
			&cli.StringFlag{
				Name:        "commit_state",
				Usage:       "State is the current state of the repository. Possible values are: pending, success, error, or failure",
				EnvVars:     []string{"PLUGIN_COMMIT_STATE"},
				DefaultText: "success",
			},
			&cli.StringFlag{
				Name:        "commit_context",
				Usage:       "A string label to differentiate this status from the statuses of other systems.",
				EnvVars:     []string{"PLUGIN_COMMIT_CONTEXT"},
				DefaultText: "drone-github-commit-status",
			},
			&cli.StringFlag{
				Name:     "commit_target_url",
				Usage:    "TargetURL is the URL of the page representing this status. It will be linked from the GitHub UI to allow users to see the source of the status.",
				EnvVars:  []string{"PLUGIN_COMMIT_TARGET_URL"},
				Required: false,
			},
			&cli.StringFlag{
				Name:     "commit_description",
				Usage:    "Description is a short high level summary of the status.",
				EnvVars:  []string{"PLUGIN_COMMIT_DESCRIPTION"},
				Required: false,
			},
		},
		Action: func(cCtx *cli.Context) error {

			drone_repo_owner := os.Getenv("DRONE_REPO_OWNER")
			drone_repo_name := os.Getenv("DRONE_REPO_NAME")
			drone_commit_sha := os.Getenv("DRONE_COMMIT_SHA")

			github_app_id, err := strconv.ParseInt(cCtx.String("github_app_id"), 10, 64)
			if err != nil {
				log.Fatal(err)
			}

			github_installation_id, err := strconv.ParseInt(cCtx.String("github_installation_id"), 10, 64)
			if err != nil {
				log.Fatal(err)
			}

			github_private_key := cCtx.String("github_app_private_key")
			status_state := cCtx.String("commit_state")
			status_context := cCtx.String("commit_context")
			status_target_url := cCtx.String("commit_target_url")
			status_description := cCtx.String("commit_description")

			status := github.RepoStatus{
				State:       &status_state,
				Context:     &status_context,
				TargetURL:   &status_target_url,
				Description: &status_description,
			}

			// Shared transport to reuse TCP connections.
			tr := http.DefaultTransport

			writePrivateKeyFile(github_private_key)
			itr, err := ghinstallation.NewKeyFromFile(tr, github_app_id, github_installation_id, "private-key.pem")
			if err != nil {
				log.Fatal(err)
			}

			// Use installation transport with github.com/google/go-github
			client := github.NewClient(&http.Client{Transport: itr})

			res, _, err := client.Repositories.CreateStatus(context.Background(), drone_repo_owner, drone_repo_name, drone_commit_sha, &status)
			if err != nil {
				log.Fatal(err)
			}

			json, err := json.Marshal(res)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println(string(json))

			return nil

		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func writePrivateKeyFile(privateKeyString string) {
	b := []byte(privateKeyString)
	err := os.WriteFile("private-key.pem", b, 0644)
	if err != nil {
		log.Fatal(err)
	}
}
