# GPT-PullRequest-Updater

[![go-recipes](https://raw.githubusercontent.com/nikolaydubina/go-recipes/main/badge.svg?raw=true)](https://github.com/nikolaydubina/go-recipes)

This repository contains a tool for updating and reviewing GitHub pull requests using OpenAI's GPT language model. The project has two commands: `description` and `review`. The `description` command updates the pull request description with a high-level summary of the changes made. The `review` command creates individual comments for each file and an overall review summary comment.

## Requirements

- GitHub token with access to the desired repository
- OpenAI API token

## Installation

```bash
go install github.com/ravilushqa/gpt-pullrequest-updater/cmd/description@latest
go install github.com/ravilushqa/gpt-pullrequest-updater/cmd/review@latest
```

## Usage

### Review Command

Usage:
  ```
  review [OPTIONS]
  ```

Application Options:
  ```
      --gh-token=     GitHub token [$GITHUB_TOKEN]
      --openai-token= OpenAI token [$OPENAI_TOKEN]
      --owner=        GitHub owner [$OWNER]
      --repo=         GitHub repo [$REPO]
      --pr-number=    Pull request number [$PR_NUMBER]
      --openai-model= OpenAI model (default: gpt-3.5-turbo) [$OPENAI_MODEL]
      --test          Test mode [$TEST]
  ```

Help Options:
  ```
  -h, --help          Show this help message
  ```

Before running the command, make sure you have set the appropriate options or environment variables. The command line options will take precedence over the environment variables.

To run the `review` command, execute:

```
./review --gh-token=<GITHUB_TOKEN> --openai-token=<OPENAI_TOKEN> --owner=<OWNER> --repo=<REPO> --pr-number=<PR_NUMBER> --test
```

Replace `<GITHUB_TOKEN>`, `<OPENAI_TOKEN>`, `<OWNER>`, `<REPO>`, and `<PR_NUMBER>` with the appropriate values. If you want to enable test mode, add the `--test` flag.

### Description Command

The usage for the `description` command is similar to the `review` command. Replace `review` with `description` in the command above and execute.
Only difference is that `description` command has extra option `--jira-url` which is used to generate Jira links in the description.

## GitHub Action

This script can be used as a GitHub Action, allowing it to run automatically in your repository. To get started, add a new workflow file in your repository, such as: `.github/workflows/gpt_pullrequest_updater.yml`.

Here's an example of what the workflow file could look like:

```yaml
name: GPT Pull Request Updater

on:
   pull_request:
      types:
         - opened
         - synchronize

jobs:
   update_pull_request:
      runs-on: ubuntu-latest

      steps:
         - name: Set up Go
           uses: actions/setup-go@v2
           with:
              go-version: 1.19

         - name: Checkout GPT-PullRequest-Updater
           uses: actions/checkout@v2
           with:
              repository: ravilushqa/gpt-pullrequest-updater
              path: gpt-pullrequest-updater

         - name: Build description and review commands
           run: |
              cd gpt-pullrequest-updater
              make build

         - name: Update Pull Request Description
           run: |
              ./gpt-pullrequest-updater/bin/description
           env:
              GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
              OPENAI_TOKEN: ${{ secrets.OPENAI_TOKEN }}
              OWNER: ${{ github.repository_owner }}
              REPO: ${{ github.event.repository.name }}
              PR_NUMBER: ${{ github.event.number }}

         - name: Review Pull Request
           if: github.event.action == 'opened'
           run: |
              ./gpt-pullrequest-updater/bin/review
           env:
              GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
              OPENAI_TOKEN: ${{ secrets.OPENAI_TOKEN }}
              OWNER: ${{ github.repository_owner }}
              REPO: ${{ github.event.repository.name }}
              PR_NUMBER: ${{ github.event.number }}

```

Make sure to add your OpenAI API token to your repository secrets as `OPENAI_TOKEN`.

### Granting Permissions for GitHub Actions

In order to use this GitHub Action, you need to grant the necessary permissions to the GitHub token. To do this, follow these steps:

1. Go to the repository settings page: https://github.com/OWNER/REPO/settings
2. Navigate to the "Actions" tab on the left side of the settings page.
3. Scroll down to the "Workflow Permissions" section.

Select "Read and Write" permissions for the actions. This will provide your token with the necessary rights to modify your repository.
By following these steps, you'll grant the required permissions for the GPT-PullRequest-Updater GitHub Action to function properly, allowing it to update and review pull requests in your repository.

License
This project is licensed under the MIT License.
