#!/bin/bash
set -e

# Full bootstrap: generate, init DB, migrate, seed, build, then run.
bash ./init_project.sh
bash ./run.sh "$@"
