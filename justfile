# hubstaff-tui development tasks
# Run `just` to see all available recipes

set dotenv-load := false

# Project metadata
version := `cat VERSION 2>/dev/null || echo "dev"`
binary := "hubstaff-tui"

# Default recipe: show help
default:
    @just --list

mod build '.just/build'
mod test '.just/test'
mod lint '.just/lint'
mod install '.just/install'
mod dev '.just/dev'

# Run all checks (build + vet + lint + test)
check: build::build lint::vet lint::lint test::test
