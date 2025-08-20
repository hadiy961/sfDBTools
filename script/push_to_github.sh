#!/usr/bin/env bash
set -euo pipefail

# Robust script to push the repository to GitHub.
# Supports SSH (default if remote is SSH) or HTTPS with Personal Access Token.
# Usage examples:
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT" || exit 1

usage() {
  cat <<EOF
Usage: $(basename "$0") [options]

Options:
  -t TOKEN      GitHub Personal Access Token (or set GITHUB_TOKEN)
  -r REPO_URL   Remote repository URL (https or ssh). If omitted, uses 'origin' remote
  -b BRANCH     Branch to push (default: main)
  -m MESSAGE    Commit message if auto-committing changes
  --ssh         Force using SSH push (ignores token)
  -h            Show this help
EOF
  exit 1
}

TOKEN="${GITHUB_TOKEN:-}"
REPO_URL=""
BRANCH="main"
COMMIT_MSG=""
FORCE_SSH=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    -t) TOKEN="$2"; shift 2 ;;
    -r) REPO_URL="$2"; shift 2 ;;
    -b) BRANCH="$2"; shift 2 ;;
    -m) COMMIT_MSG="$2"; shift 2 ;;
    --ssh) FORCE_SSH=1; shift ;;
    -h|--help) usage ;;
    --) shift; break ;;
    *) echo "Unknown option: $1"; usage ;;
  esac
done

echo "Project: $PROJECT_ROOT"

# validate git repo
if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  echo "ERROR: Not a git repository (or git not installed)." >&2
  exit 2
fi

# determine remote if not provided
if [[ -z "$REPO_URL" ]]; then
  if git remote get-url origin >/dev/null 2>&1; then
    REPO_URL=$(git remote get-url origin)
    echo "Using existing 'origin' remote: ${REPO_URL//\$TOKEN/****}" >/dev/stderr || true
  else
    read -rp "No 'origin' remote found. Enter repository URL (https://... or git@...): " REPO_URL
  fi
fi

# decide mode (ssh or https)
if [[ $FORCE_SSH -eq 1 ]]; then
  USE_SSH=1
elif [[ "$REPO_URL" =~ ^git@|^ssh:// ]]; then
  USE_SSH=1
else
  USE_SSH=0
fi

# if we need token but it's missing, prompt interactively
if [[ $USE_SSH -eq 0 && -z "$TOKEN" ]]; then
  read -rsp "Enter GitHub Personal Access Token: " TOKEN
  echo
fi

# show git status and handle uncommitted changes
CHANGES=$(git status --porcelain)
if [[ -n "$CHANGES" ]]; then
  echo "Uncommitted changes detected:";
  git status --short
  if [[ -n "$COMMIT_MSG" ]]; then
    echo "Auto-adding and committing changes..."
    git add -A
    git commit -m "$COMMIT_MSG"
  else
    read -rp "There are uncommitted changes. Add and commit them with a message? (y/N) " yn
    if [[ "$yn" =~ ^[Yy]$ ]]; then
      read -rp "Commit message: " COMMIT_MSG
      git add -A
      git commit -m "$COMMIT_MSG"
    else
      echo "Aborting push. Commit or stash your changes and try again."; exit 3
    fi
  fi
fi

echo "Preparing to push branch '$BRANCH' to remote..."

if [[ $USE_SSH -eq 1 ]]; then
  echo "Using SSH push to remote: $REPO_URL"
  git push -u origin "$BRANCH"
else
  # Ensure repo URL is https and construct a push URL that embeds token
  # Acceptable input forms: https://github.com/owner/repo.git or https://github.com/owner/repo
  if [[ "$REPO_URL" =~ ^https:// ]]; then
    # strip possible protocol prefix
    url_no_proto="${REPO_URL#https://}"
    push_url="https://$TOKEN@${url_no_proto}"
    echo "Pushing via HTTPS (token secured) to ${REPO_URL%%\?*}"
    # push directly to URL to avoid permanently storing token in remote config
    git push "$push_url" "$BRANCH"
  else
    # try to convert ssh-style to https if possible (git@github.com:owner/repo.git)
    if [[ "$REPO_URL" =~ ^git@github.com:(.+)$ ]]; then
      repo_path="${BASH_REMATCH[1]}"
      push_url="https://$TOKEN@github.com/${repo_path}"
      echo "Converted SSH remote to HTTPS and will push to: https://github.com/${repo_path}"
      git push "$push_url" "$BRANCH"
    else
      echo "Cannot determine HTTPS push URL from: $REPO_URL" >&2
      exit 4
    fi
  fi
fi

if [[ $? -eq 0 ]]; then
  # mask repo url for display
  display_repo="$REPO_URL"
  display_repo="${display_repo/$TOKEN/****}"
  echo
  echo "✅ SUCCESS: pushed branch '$BRANCH' to $display_repo"
else
  echo
  echo "❌ ERROR: git push failed. Check token, permissions, branch name, and remote URL."
  exit 5
fi
