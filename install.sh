#!/bin/bash

# golang check for version specified in go.mod
echo "[*] Checking golang"
go version
if [ $? -ne 0 ]; then
    echo "Failed to check golang version: do you have go installed?"
    exit 1
fi
echo "[OK] golang installed"

# Building and installing sgrep
echo "[*] Building and installing sgrep"
go  build -o sgrep
if [ $? -ne 0 ]; then
    echo "Failed to build sgrep"
    exit 1
fi
echo "[OK] sgrep built"

# Moving sgrep to /usr/bin/sgrep
sudo cp sgrep /usr/bin/sgrep
if [ $? -ne 0 ]; then
    echo "Failed to install sgrep"
    exit 1
fi
echo "[OK] sgrep installed in /usr/bin/sgrep"

# Setting configuration path
echo "[*] Setting configuration path in $HOME/.config/semantic-grep/"
mkdir -p "$HOME/.config/semantic-grep/"

# Downloading the model and moving to $HOME/.config/semantic-grep/
echo "[*] Downloading the model"
if [ -f "models/googlenews-slim/GoogleNews-vectors-negative300-SLIM.bin" ]; then
    echo "[OK] Model already downloaded"
else
    bash download-model.sh
    if [ $? -ne 0 ]; then
        echo "Failed to download the model"
        exit 1
    fi
fi

echo "[OK] Model downloaded - moving to $HOME/.config/semantic-grep/"
cp -r models "$HOME/.config/semantic-grep/"
echo "[OK] Model moved to $HOME/.config/semantic-grep/"

# Setting model path
echo "[*] Setting model path"
CONFIG_STRING="{\"model_path\":\"$HOME/.config/semantic-grep/models/googlenews-slim/GoogleNews-vectors-negative300-SLIM.bin\"}"
echo $CONFIG_STRING > "$HOME/.config/semantic-grep/config.json"
if [ $? -ne 0 ]; then
    echo "Failed to set model path"
    exit 1
fi
echo "[OK] Model path set:"
cat "$HOME/.config/semantic-grep/config.json"

# Testing
echo "[*] Testing"
sgrep -h
if [ $? -ne 0 ]; then
    echo "Failed to test sgrep"
    exit 1
fi
echo "[OK] sgrep tested and working. Installation complete."