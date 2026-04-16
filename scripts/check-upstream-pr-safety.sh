#!/usr/bin/env bash

set -euo pipefail

usage() {
  cat <<'EOF'
Usage: check-upstream-pr-safety.sh [base-ref] [target-ref] [extra-pattern]

Audits the diff between two refs for fork-only files and private content that
should not appear in an upstream PhotoPrism pull request.

Arguments:
  base-ref       Base ref to compare against. Default: develop
  target-ref     Ref to audit. Default: HEAD
  extra-pattern  Optional additional ERE regex for personal names, places,
                 or other repo-specific strings you do not want in public PRs.

Environment:
  UPSTREAM_PR_EXTRA_PATTERN
                 Optional extra regex used when the third argument is omitted.

Examples:
  ./scripts/check-upstream-pr-safety.sh
  ./scripts/check-upstream-pr-safety.sh develop origin/feature/metadata
  UPSTREAM_PR_EXTRA_PATTERN='name-one|name-two|town-name' \
    ./scripts/check-upstream-pr-safety.sh develop origin/feature/metadata
EOF
}

if [[ ${1:-} == "-h" ]] || [[ ${1:-} == "--help" ]]; then
  usage
  exit 0
fi

BASE_REF=${1:-develop}
TARGET_REF=${2:-HEAD}
EXTRA_PATTERN=${3:-${UPSTREAM_PR_EXTRA_PATTERN:-}}

git rev-parse --verify "${BASE_REF}^{commit}" >/dev/null
git rev-parse --verify "${TARGET_REF}^{commit}" >/dev/null

declare -a CHANGED_FILES=()
while IFS= read -r file; do
  CHANGED_FILES+=("${file}")
done < <(git diff --name-only "${BASE_REF}...${TARGET_REF}")

if [[ ${#CHANGED_FILES[@]} -eq 0 ]]; then
  echo "No files changed between ${BASE_REF} and ${TARGET_REF}."
  exit 0
fi

origin_url=$(git remote get-url origin 2>/dev/null || true)
origin_owner=""

if [[ ${origin_url} =~ ^git@github\.com:([^/]+)/ ]]; then
  origin_owner=${BASH_REMATCH[1]}
elif [[ ${origin_url} =~ ^https://github\.com/([^/]+)/ ]]; then
  origin_owner=${BASH_REMATCH[1]}
fi

declare -a TREE_FILES=()
for file in "${CHANGED_FILES[@]}"; do
  if git cat-file -e "${TARGET_REF}:${file}" 2>/dev/null; then
    TREE_FILES+=("${file}")
  fi
done

failed=0

report_match() {
  local label=$1
  shift

  if [[ $# -eq 0 ]]; then
    return 0
  fi

  failed=1
  echo
  echo "[$label]"
  printf '%s\n' "$@"
}

check_blocked_paths() {
  local -a matches=()
  local file

  for file in "${CHANGED_FILES[@]}"; do
    case "${file}" in
      README.md|AGENTS.md|.github/workflows/build-production.yml)
        matches+=("${file}")
        ;;
    esac
  done

  report_match "Fork-only files changed" "${matches[@]}"
}

check_content_pattern() {
  local label=$1
  local pattern=$2
  shift 2

  if [[ ${#TREE_FILES[@]} -eq 0 ]]; then
    return 0
  fi

  local output
  output=$(git grep -n -I -E "${pattern}" "${TARGET_REF}" -- "${TREE_FILES[@]}" || true)

  if [[ -n ${output} ]]; then
    report_match "${label}" "${output}"
  fi
}

check_blocked_paths
check_content_pattern "Absolute host paths" '/Users/|/Volumes/|/private/tmp/'
check_content_pattern "Package registry URLs" 'ghcr\.io/[^[:space:]")'\''>]+'
check_content_pattern "SSH GitHub remotes" 'git@github\.com:[^[:space:]")'\''>]+'

if [[ -n ${origin_owner} ]]; then
  check_content_pattern "Fork owner references" "github\\.com/${origin_owner}/|git@github\\.com:${origin_owner}/|ghcr\\.io/${origin_owner}/"
fi

if [[ -n ${EXTRA_PATTERN} ]]; then
  check_content_pattern "Extra private pattern matches" "${EXTRA_PATTERN}"
fi

if [[ ${failed} -eq 1 ]]; then
  echo
  echo "Upstream PR safety audit failed for ${TARGET_REF} against ${BASE_REF}."
  exit 1
fi

echo "Upstream PR safety audit passed for ${TARGET_REF} against ${BASE_REF}."
