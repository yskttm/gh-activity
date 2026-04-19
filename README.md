# gh-activity

[![CI](https://github.com/yskttm/gh-activity/actions/workflows/ci.yml/badge.svg)](https://github.com/yskttm/gh-activity/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/yskttm/gh-activity)](https://github.com/yskttm/gh-activity/releases/latest)
[![Go Version](https://img.shields.io/github/go-mod/go-version/yskttm/gh-activity)](go.mod)

GitHub CLI extension to aggregate Issue/PR activity for a user.

## Requirements

- [gh](https://cli.github.com/) — authenticated with `gh auth login`
- [Go](https://golang.org/) — required to build from source

## Installation

```bash
gh extension install yskttm/gh-activity
```

## Usage

### Issues

```bash
gh activity issue <username> [flags]
```

### Pull Requests

```bash
gh activity pr <username> [flags]
```

### Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--from` | Start date (`YYYY-MM-DD`) | today |
| `--to` | End date (`YYYY-MM-DD`) | today |
| `--repos` | Target repositories, comma-separated (e.g. `org/repo1,org/repo2`) | all repositories |
| `--fields` | Fields to display, comma-separated | all fields |
| `--format` | Output format: `table` or `csv` | `table` |

### Available Fields

| Field | Description |
|-------|-------------|
| `no` | Row number |
| `repo` | Repository (`org/repo`) |
| `number` | Issue/PR number |
| `url` | URL |
| `title` | Title |
| `state` | State (`open` / `closed`) |
| `creator` | Author |
| `assignee` | Assignees (comma-separated) |
| `labels` | Labels (comma-separated) |
| `milestone` | Milestone |
| `comments` | Number of comments |
| `created_at` | Created date |
| `closed_at` | Closed date |
| `merged_at` | Merged date (PR only) |
| `draft` | Draft flag (PR only) |

## Examples

```bash
# Issues closed today
gh activity issue yskttm

# Issues closed in a date range
gh activity issue yskttm --from=2025-01-01 --to=2025-03-31

# Issues in specific repositories
gh activity issue yskttm --from=2025-01-01 --repos=org/repo1,org/repo2

# Show specific fields only
gh activity issue yskttm --from=2025-01-01 --fields=no,repo,title,creator,closed_at

# Export to CSV
gh activity issue yskttm --from=2025-01-01 --format=csv > output.csv

# Merged PRs
gh activity pr yskttm --from=2025-01-01 --to=2025-03-31
```

## Search Criteria

| Command | Filter |
|---------|--------|
| `issue` | Issues where the user is **assignee or author**, filtered by **closed date** |
| `pr` | PRs where the user is **author**, filtered by **merged date** |

## Notes

- Uses the [GitHub Search API](https://docs.github.com/en/rest/search/search), which returns up to 1000 results per query.
- When results exceed 1000, the date range is automatically split and queried recursively.
