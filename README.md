# Pull Request Description Generator

This Go program automates the process of generating GitHub pull request descriptions based on the changes made in each file. It uses OpenAI's GPT-3.5-turbo model to generate the descriptions and Jira issue links based on the PR title.

## Installation

To install the program, clone the repository and build the binary:

```sh
git clone https://github.com/your-repo/pull-request-description-generator.git
cd pull-request-description-generator
go build
```

## Usage

Before running the program, you'll need to set the following environment variables:

- `GITHUB_TOKEN`: Your GitHub access token.
- `OPENAI_TOKEN`: Your OpenAI access token.
- `OWNER`: The GitHub owner (username or organization) of the repository.
- `REPO`: The GitHub repository name.
- `PR_NUMBER`: The pull request number.

You can also use flags to provide the required information:

```
./pull-request-description-generator --gh-token <GITHUB_TOKEN> --openai-token <OPENAI_TOKEN> --owner <OWNER> --repo <REPO> --pr-number <PR_NUMBER>
```

Optional flags:

- `--test`: Test mode. The generated description will be printed to the console without updating the pull request.
- `--skip-files`: Comma-separated list of files to skip when generating the description (default: "go.mod,go.sum,.pb.go").

After running the program, the pull request description will be updated with the generated content.

## Dependencies

- [go-flags](https://github.com/jessevdk/go-flags): A Go library for command line flag parsing.
- [go-openai](https://github.com/sashabaranov/go-openai): A Go client for the OpenAI API.

## Functions

- `getDiffContent`: Fetches the diff content from the GitHub API.
- `parseGitDiffAndSplitPerFile`: Parses the git diff and splits it into a slice of FileDiff.
- `getFilenameFromDiffHeader`: Extracts the filename from a diff header.
- `generatePRDescription`: Generates the pull request description using the OpenAI API.
- `getPullRequestTitle`: Fetches the pull request title from the GitHub API.
- `generateJiraLinkByTitle`: Generates a Jira issue link based on the PR title.
- `updatePullRequestDescription`: Updates the pull request description on GitHub.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
