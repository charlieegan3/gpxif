GO_TEST_ARGS := ""
FILE_PATTERN := 'go\|Makefile'

# Define recipes
test:
    go test ./... {{ GO_TEST_ARGS }}

test_watch:
    find . | grep '{{ FILE_PATTERN }}' | entr bash -r -c 'clear; just test'

build:
    go build -o gpxif main.go

gomod2nix:
    #!/usr/bin/env bash
    old_content=$(cat gomod2nix.toml)
    rm gomod2nix.toml
    gomod2nix
    if [[ "$old_content" != "$(cat gomod2nix.toml)" ]]; then
      echo "Error: The file contents have changed."
      exit 1
    fi
