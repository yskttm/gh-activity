# gh-activity

GitHub CLI を使って、メンバーの Issue / PR 活動を集計・出力するシェルスクリプト群。

---

## ファイル構成

```
gh-activity/
├── gh-activity.sh        # 共通関数ライブラリ（直接実行不可）
├── gh-issue-activity.sh  # Issue 専用スクリプト
└── gh-pr-activity.sh     # PR 専用スクリプト
```

---

## 仕様

### 引数

| 位置 | 引数 | 説明 |
|------|------|------|
| 1 | `username` | 対象の GitHub アカウント名 |
| 2 | `start-date` | 集計開始日（例: `2024-10-01`） |
| 3 | `end-date` | 集計終了日（例: `2025-03-31`） |
| 4〜 | `repo1 repo2 ...` | 対象リポジトリ（省略時は全リポジトリ対象） |

### 出力カラム

| カラム | 内容 |
|--------|------|
| No. | 連番 |
| Repo | リポジトリ名（`org/repo` 形式） |
| URL | Issue / PR の URL |
| Title | タイトル |
| Creator | 作成者 |
| Assignee | アサインされているユーザー（複数の場合はカンマ区切り） |
| Labels | ラベル（複数の場合はカンマ区切り） |

### ソート順

`Repo` → `Issue/PR番号` の昇順

### 検索条件

#### Issue（`gh-issue-activity.sh`）

- `assignee:USERNAME`（担当者）と `author:USERNAME`（起票者）の **OR 条件**で取得
- 2クエリに分けて取得し、`id` で重複排除してマージ
- 集計対象: `closed` 日が指定期間内のもの

#### PR（`gh-pr-activity.sh`）

- `author:USERNAME`（作成者）で取得
- 集計対象: `merged` 日が指定期間内のもの

### ページネーション・件数上限対応

- GitHub Search API の上限（100件/ページ、1000件/クエリ）に対応
- 1000件を超える場合は期間を自動的に半分に分割して再帰的に取得
- 分割後も1000件を超え、かつ1日単位まで分割できない場合は Warning を表示

---

## 使い方

### セットアップ

```bash
chmod +x gh-issue-activity.sh gh-pr-activity.sh
```

### Issue の集計

```bash
# 全リポジトリ対象
./gh-issue-activity.sh octocat 2024-10-01 2025-03-31

# リポジトリを絞る
./gh-issue-activity.sh octocat 2024-10-01 2025-03-31 org/repo1 org/repo2
```

### PR の集計

```bash
# 全リポジトリ対象
./gh-pr-activity.sh octocat 2024-10-01 2025-03-31

# リポジトリを絞る
./gh-pr-activity.sh octocat 2024-10-01 2025-03-31 org/repo1 org/repo2
```

### 出力イメージ

```
Fetching issues for octocat (2024-10-01 to 2025-03-31)...
Repos: org/repo1 org/repo2

Fetching issues by assignee...
Fetching issues by author...

Total: 4 issue(s)

No.  Repo       URL                                       Title                      Creator   Assignee          Labels
1    org/repo1  https://github.com/org/repo1/issues/142  Database connection error   octocat   octocat, alice    bug
2    org/repo1  https://github.com/org/repo1/issues/187  Fix login bug               octocat   octocat           bug, high-priority
3    org/repo2  https://github.com/org/repo2/issues/23   Add search feature request  octocat                     enhancement
4    org/repo2  https://github.com/org/repo2/issues/56   Update README               bob       octocat           docs
```

---

## コード

### gh-activity.sh（共通関数）

```bash
#!/bin/bash

# 共通関数ライブラリ（直接実行不可）

date_add_days() {
  local base=$1
  local days=$2
  if date -v +1d > /dev/null 2>&1; then
    date -j -v "+${days}d" -f "%Y-%m-%d" "$base" "+%Y-%m-%d"
  else
    date -d "${base} +${days} days" "+%Y-%m-%d"
  fi
}

date_to_epoch() {
  local d=$1
  if date -v +1d > /dev/null 2>&1; then
    date -j -f "%Y-%m-%d" "$d" "+%s"
  else
    date -d "$d" "+%s"
  fi
}

build_repo_query() {
  local repos=("$@")
  local repo_query=""
  for repo in "${repos[@]}"; do
    repo_query="${repo_query}+repo:${repo}"
  done
  echo "$repo_query"
}

check_count() {
  local query=$1
  gh api "search/issues?q=${query}&per_page=1" | jq '.total_count'
}

fetch_page() {
  local query=$1
  local page=$2
  gh api "search/issues?q=${query}&per_page=100&page=${page}"
}

fetch_with_split() {
  local query_base=$1
  local range_start=$2
  local range_end=$3

  local date_query="${query_base}+${DATE_FIELD}:${range_start}..${range_end}"
  local total
  total=$(check_count "$date_query")

  if [ "$total" -eq 0 ]; then
    echo "[]"
    return
  fi

  if [ "$total" -le 1000 ]; then
    local all_items="[]"
    local page=1

    while true; do
      local response
      response=$(fetch_page "$date_query" "$page")
      local items
      items=$(echo "$response" | jq '.items')
      local count
      count=$(echo "$items" | jq 'length')

      if [ "$count" -eq 0 ]; then
        break
      fi

      all_items=$(echo "$all_items $items" | jq -s 'add')

      if [ "$count" -lt 100 ]; then
        break
      fi

      page=$((page + 1))
      sleep 1
    done

    echo "$all_items"
    return
  fi

  # 1000件超 → 期間を半分に分割して再帰
  local start_epoch
  local end_epoch
  start_epoch=$(date_to_epoch "$range_start")
  end_epoch=$(date_to_epoch "$range_end")
  local diff_days=$(( (end_epoch - start_epoch) / 86400 ))

  if [ "$diff_days" -le 1 ]; then
    echo "Warning: ${range_start}..${range_end} has ${total} items but cannot split further. Capped at 1000." >&2
    fetch_with_split "$query_base" "$range_start" "$range_end"
    return
  fi

  local mid_days=$(( diff_days / 2 ))
  local mid
  mid=$(date_add_days "$range_start" "$mid_days")
  local mid_next
  mid_next=$(date_add_days "$mid" 1)

  local first_half
  local second_half
  first_half=$(fetch_with_split "$query_base" "$range_start" "$mid")
  second_half=$(fetch_with_split "$query_base" "$mid_next" "$range_end")

  echo "$first_half $second_half" | jq -s 'add | unique_by(.id)'
}

print_results() {
  local items=$1
  local total
  total=$(echo "$items" | jq 'length')

  if [ "$total" -eq 0 ]; then
    echo "No results found."
    exit 0
  fi

  echo "Total: ${total} item(s)"
  echo ""

  echo "$items" | jq -r '
    sort_by(.repository_url, .number) |
    ["No.", "Repo", "URL", "Title", "Creator", "Assignee", "Labels"],
    (to_entries[] | [
      (.key + 1 | tostring),
      (.value.repository_url | split("/") | .[-2:] | join("/")),
      .value.html_url,
      .value.title,
      .value.user.login,
      (.value.assignees | map(.login) | join(", ")),
      (.value.labels | map(.name) | join(", "))
    ])
    | @tsv
  ' | column -t -s $'\t'
}
```

### gh-issue-activity.sh

```bash
#!/bin/bash

# Usage: ./gh-issue-activity.sh <username> <start> <end> [repo1 repo2 ...]
# Example: ./gh-issue-activity.sh octocat 2024-10-01 2025-03-31 org/repo1 org/repo2

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "${SCRIPT_DIR}/gh-activity.sh"

USERNAME=$1
START=$2
END=$3
shift 3
REPOS=("$@")

if [ -z "$USERNAME" ] || [ -z "$START" ] || [ -z "$END" ]; then
  echo "Usage: $0 <username> <start-date> <end-date> [repo1 repo2 ...]"
  echo "Example: $0 octocat 2024-10-01 2025-03-31 org/repo1 org/repo2"
  exit 1
fi

REPO_QUERY=$(build_repo_query "${REPOS[@]}")
DATE_FIELD="closed"
export DATE_FIELD

echo "Fetching issues for ${USERNAME} (${START} to ${END})..."
if [ ${#REPOS[@]} -gt 0 ]; then
  echo "Repos: ${REPOS[*]}"
fi
echo ""

echo "Fetching issues by assignee..."
ITEMS_ASSIGNEE=$(fetch_with_split "assignee:${USERNAME}+type:issue${REPO_QUERY}" "$START" "$END")

echo "Fetching issues by author..."
ITEMS_AUTHOR=$(fetch_with_split "author:${USERNAME}+type:issue${REPO_QUERY}" "$START" "$END")

ALL_ITEMS=$(echo "$ITEMS_ASSIGNEE $ITEMS_AUTHOR" | jq -s 'add | unique_by(.id) | sort_by(.closed_at)')

print_results "$ALL_ITEMS"
```

### gh-pr-activity.sh

```bash
#!/bin/bash

# Usage: ./gh-pr-activity.sh <username> <start> <end> [repo1 repo2 ...]
# Example: ./gh-pr-activity.sh octocat 2024-10-01 2025-03-31 org/repo1 org/repo2

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "${SCRIPT_DIR}/gh-activity.sh"

USERNAME=$1
START=$2
END=$3
shift 3
REPOS=("$@")

if [ -z "$USERNAME" ] || [ -z "$START" ] || [ -z "$END" ]; then
  echo "Usage: $0 <username> <start-date> <end-date> [repo1 repo2 ...]"
  echo "Example: $0 octocat 2024-10-01 2025-03-31 org/repo1 org/repo2"
  exit 1
fi

REPO_QUERY=$(build_repo_query "${REPOS[@]}")
DATE_FIELD="merged"
export DATE_FIELD

echo "Fetching PRs for ${USERNAME} (${START} to ${END})..."
if [ ${#REPOS[@]} -gt 0 ]; then
  echo "Repos: ${REPOS[*]}"
fi
echo ""

ALL_ITEMS=$(fetch_with_split "author:${USERNAME}+type:pr${REPO_QUERY}" "$START" "$END")

print_results "$ALL_ITEMS"
```

---

## 依存関係

- `gh` — GitHub CLI（要認証済み）
- `jq` — JSON パーサー
- `column` — 表形式出力（macOS / Linux 標準）
