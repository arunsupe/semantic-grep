#!/bin/bash

# Model URL and save path (modify these if needed)
model_url="https://github.com/eyaler/word2vec-slim/raw/master/GoogleNews-vectors-negative300-SLIM.bin.gz"
save_path="models/googlenews-slim/GoogleNews-vectors-negative300-SLIM.bin"

# Create directory (if it doesn't exist)
mkdir -p "$(dirname "$save_path")"

# Download the model with progress bar (simulated)
echo "Downloading..."
wget -O "$save_path.gz" -q --show-progress "$model_url"

# Decompress the file
gunzip "$save_path.gz"

echo "Model downloaded and saved to: $save_path"
