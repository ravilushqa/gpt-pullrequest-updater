# GPT Pull Request Updater

GPT Pull Request Updater is a command line tool that uses OpenAI's GPT-4 model to generate a comprehensive description of a GitHub pull request, including a summary of code changes. It then updates the pull request description with the generated summary.

## Prerequisites

To use the GPT Pull Request Updater, you need the following:
- A GitHub token with repository access
- An OpenAI API token

## Installation

1. Clone the repository:

```sh
git clone https://github.com/ravilushqa/gpt-pullrequest-updater.git
```

1. Change to the repository directory:

```sh
cd gpt-pullrequest-updater
```

1. Build the binary:

```sh
go build -o gpt-pr-updater
```

1. Add the binary to your PATH:

```sh
export PATH=$PATH:$(pwd)
```

## Usage

Run the GPT Pull Request Updater with the following command, providing the required flags:

```sh
gpt-pr-updater --gh-token GITHUB_TOKEN --openai-token OPENAI_TOKEN --owner OWNER --repo REPO --pr-number PR_NUMBER
```

### Flags

- `--gh-token` (required): Your GitHub token. Can also be set with the `GITHUB_TOKEN` environment variable.
- `--openai-token` (required): Your OpenAI API token. Can also be set with the `OPENAI_TOKEN` environment variable.
- `--owner` (required): GitHub repository owner. Can also be set with the `OWNER` environment variable.
- `--repo` (required): GitHub repository name. Can also be set with the `REPO` environment variable.
- `--pr-number` (required): The number of the pull request to update. Can also be set with the `PR_NUMBER` environment variable.
- `--test` (optional): If set, the tool will print the generated description to the console without updating the pull request. Can also be set with the `TEST` environment variable.

## Example

```sh
gpt-pr-updater --gh-token your_github_token --openai-token your_openai_token --owner ravilushqa --repo myrepo --pr-number 42
```

This command will fetch the pull request #42 from the `myrepo` repository, generate a summary of code changes using GPT-4, and update the pull request description with the generated summary.

## License

This project is released under the [MIT License](LICENSE).