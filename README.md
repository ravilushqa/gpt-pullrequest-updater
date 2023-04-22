# GPT-PullRequest-Updater

This repository contains a tool for updating and reviewing GitHub pull requests using OpenAI's GPT language model. The project has two commands: `description` and `review`. The `description` command updates the pull request description with a high-level summary of the changes made. The `review` command creates individual comments for each file and an overall review summary comment.

## Requirements

- GitHub token with access to the desired repository
- OpenAI API token

## Installation

1. Clone the repository:

   ```
   git clone https://github.com/ravilushqa/gpt-pullrequest-updater.git
   ```

2. Navigate to the project root:

   ```
   cd gpt-pullrequest-updater
   ```

3. Build the commands:

   ```
   go build -o description ./cmd/description
   go build -o review ./cmd/review
   ```

## Usage

Before running the commands, make sure you have set the following environment variables:

- `GITHUB_TOKEN`: Your GitHub token
- `OPENAI_TOKEN`: Your OpenAI API token
- `OWNER`: The owner of the GitHub repository
- `REPO`: The name of the GitHub repository
- `PR_NUMBER`: The pull request number you want to update or review

### Description Command

The `description` command updates the pull request description with a high-level summary of the changes made. To run the command, execute:

```
./description
```

### Review Command

The `review` command creates individual comments for each file and an overall review summary comment. To run the command, execute:

```
./review
```

### Test Mode

Both commands support a test mode that prints the generated content to the console instead of updating the pull request. To enable test mode, set the `TEST` environment variable to `true`:

```
export TEST=true
```

Then, run the desired command as described above. The generated content will be printed to the console.

## License

This project is licensed under the MIT License.