#!/usr/bin/env bash

curl -fsSL https://ollama.com/install.sh | sh

# kill any running ollama process so that the tests can start from a clean state
sudo systemctl stop ollama.service
