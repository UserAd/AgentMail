#!/usr/bin/env bash

set -e

JSON_MODE=false
BRANCH_NAME=""
ARGS=()
i=1
while [ $i -le $# ]; do
    arg="${!i}"
    case "$arg" in
        --json)
            JSON_MODE=true
            ;;
        --branch-name)
            if [ $((i + 1)) -gt $# ]; then
                echo 'Error: --branch-name requires a value' >&2
                exit 1
            fi
            i=$((i + 1))
            next_arg="${!i}"
            if [[ "$next_arg" == --* ]]; then
                echo 'Error: --branch-name requires a value' >&2
                exit 1
            fi
            BRANCH_NAME="$next_arg"
            ;;
        --help|-h)
            echo "Usage: $0 [--json] --branch-name <name> <feature_description>"
            echo ""
            echo "Options:"
            echo "  --json                Output in JSON format"
            echo "  --branch-name <name>  Branch name for the feature (required)"
            echo "  --help, -h            Show this help message"
            echo ""
            echo "Examples:"
            echo "  $0 --branch-name 'user-auth' 'Add user authentication system'"
            echo "  $0 --json --branch-name 'oauth2-api' 'Implement OAuth2 integration'"
            exit 0
            ;;
        *)
            ARGS+=("$arg")
            ;;
    esac
    i=$((i + 1))
done

FEATURE_DESCRIPTION="${ARGS[*]}"
if [ -z "$FEATURE_DESCRIPTION" ]; then
    echo "Usage: $0 [--json] --branch-name <name> <feature_description>" >&2
    exit 1
fi

if [ -z "$BRANCH_NAME" ]; then
    echo "Error: --branch-name is required. Please provide a branch name for the feature." >&2
    exit 1
fi

# Function to find the repository root by searching for existing project markers
find_repo_root() {
    local dir="$1"
    while [ "$dir" != "/" ]; do
        if [ -d "$dir/.git" ] || [ -d "$dir/.specify" ]; then
            echo "$dir"
            return 0
        fi
        dir="$(dirname "$dir")"
    done
    return 1
}

# Function to clean and format a branch name (preserves / for prefixes like feature/, tech/)
clean_branch_name() {
    local name="$1"
    echo "$name" | tr '[:upper:]' '[:lower:]' | sed 's|[^a-z0-9/]|-|g' | sed -E 's/-+/-/g' | sed -E 's|/+|/|g' | sed 's|^[/-]*||' | sed 's|[/-]*$||'
}

# Translate branch name to directory-safe name (/ → -)
branch_to_dirname() {
    echo "$1" | sed 's|/|-|g' | sed -E 's/-+/-/g' | sed 's/^-//' | sed 's/-$//'
}

# Resolve repository root
SCRIPT_DIR="$(CDPATH="" cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if git rev-parse --show-toplevel >/dev/null 2>&1; then
    REPO_ROOT=$(git rev-parse --show-toplevel)
    HAS_GIT=true
else
    REPO_ROOT="$(find_repo_root "$SCRIPT_DIR")"
    if [ -z "$REPO_ROOT" ]; then
        echo "Error: Could not determine repository root. Please run this script from within the repository." >&2
        exit 1
    fi
    HAS_GIT=false
fi

cd "$REPO_ROOT"

SPECS_DIR="$REPO_ROOT/specs"
mkdir -p "$SPECS_DIR"

# Clean the branch name
BRANCH_NAME=$(clean_branch_name "$BRANCH_NAME")

# GitHub enforces a 244-byte limit on branch names
MAX_BRANCH_LENGTH=244
if [ ${#BRANCH_NAME} -gt $MAX_BRANCH_LENGTH ]; then
    TRUNCATED_NAME=$(echo "$BRANCH_NAME" | cut -c1-$MAX_BRANCH_LENGTH)
    TRUNCATED_NAME=$(echo "$TRUNCATED_NAME" | sed 's|[-/]*$||')

    ORIGINAL_BRANCH_NAME="$BRANCH_NAME"
    BRANCH_NAME="$TRUNCATED_NAME"

    >&2 echo "[specify] Warning: Branch name exceeded GitHub's 244-byte limit"
    >&2 echo "[specify] Original: $ORIGINAL_BRANCH_NAME (${#ORIGINAL_BRANCH_NAME} bytes)"
    >&2 echo "[specify] Truncated to: $BRANCH_NAME (${#BRANCH_NAME} bytes)"
fi

if [ "$HAS_GIT" = true ]; then
    if git show-ref --verify --quiet "refs/heads/$BRANCH_NAME" 2>/dev/null; then
        >&2 echo "[specify] Branch '$BRANCH_NAME' already exists, switching to it"
        git checkout "$BRANCH_NAME"
    else
        git checkout -b "$BRANCH_NAME"
    fi
else
    >&2 echo "[specify] Warning: Git repository not detected; skipped branch creation for $BRANCH_NAME"
fi

DIR_NAME=$(branch_to_dirname "$BRANCH_NAME")
FEATURE_DIR="$SPECS_DIR/$DIR_NAME"
mkdir -p "$FEATURE_DIR"

TEMPLATE="$REPO_ROOT/.specify/templates/spec-template.md"
SPEC_FILE="$FEATURE_DIR/spec.md"
if [ ! -f "$SPEC_FILE" ]; then
    if [ -f "$TEMPLATE" ]; then cp "$TEMPLATE" "$SPEC_FILE"; else touch "$SPEC_FILE"; fi
fi

# Set the SPECIFY_FEATURE environment variable for the current session
export SPECIFY_FEATURE="$BRANCH_NAME"

if $JSON_MODE; then
    printf '{"BRANCH_NAME":"%s","SPEC_FILE":"%s"}\n' "$BRANCH_NAME" "$SPEC_FILE"
else
    echo "BRANCH_NAME: $BRANCH_NAME"
    echo "SPEC_FILE: $SPEC_FILE"
    echo "SPECIFY_FEATURE environment variable set to: $BRANCH_NAME"
fi
