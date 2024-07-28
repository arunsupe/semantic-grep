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

# Prompt to either move to /usr/bin/sgrep, adding local path to $PATH or doing nothing
echo "[*] Where would you like to install sgrep?"
echo "1. /usr/bin/sgrep (will require sudo)"
echo "2. Add local path to \$PATH"
echo "3. Do nothing"
read -p "Enter your choice: " choice

INSTALL_PATH="$(pwd)/sgrep"
if [ $choice -eq 1 ]; then
    echo "[*] Installing sgrep in /usr/bin/sgrep, please enter your password as sudo is required."
    sudo cp sgrep /usr/bin/sgrep
    if [ $? -ne 0 ]; then
        echo "Failed to install sgrep"
        exit 1
    fi
    INSTALL_PATH="/usr/bin/sgrep"
    echo "[OK] sgrep installed in /usr/bin/sgrep"
elif [ $choice -eq 2 ]; then
    echo "[*] Adding local path to \$PATH"
    export PATH=$PATH:$(pwd)
    echo "[OK] Local path added to \$PATH: to make this permanent, add the following line to your shell configuration file (e.g. ~/.bashrc or ~/.zshrc):"
    echo "export PATH=\$PATH:$(pwd)"
elif [ $choice -eq 3 ]; then
    echo "[*] Skipping installation"
    echo "[OK] Skipped installation"
else
    echo "Invalid choice"
    echo "[*] Skipping installation"
    echo "[OK] Skipped installation"
fi

# Setting configuration path

CONFIG_PATH="./config.json"
# ask user if they want to install the model
echo "[*] Do you want to install the configuration to $HOME/.config/semantic-grep/?"
echo "1. Yes"
echo "2. No"
read -p "Enter your choice: " choice

if [ $choice -eq 1 ]; then
    echo "[*] Setting configuration path in $HOME/.config/semantic-grep/"
    mkdir -p "$HOME/.config/semantic-grep/"
    CONFIG_PATH="$HOME/.config/semantic-grep/config.json"
else
    echo "[*] Skipping configuration installation"
    echo "[OK] Skipped configuration installation"
fi

# Downloading the model
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

echo "[OK] Model downloaded"
# User prompt to move the model to $HOME/.config/semantic-grep/
MODEL_PATH="./models/googlenews-slim/GoogleNews-vectors-negative300-SLIM.bin"
echo "[*] Do you want to move the model to $HOME/.config/semantic-grep/?"
echo "1. Yes"
echo "2. No"
read -p "Enter your choice: " choice
if [ $choice -eq 1 ]; then
    cp -r models "$HOME/.config/semantic-grep/"
    echo "[OK] Model moved to $HOME/.config/semantic-grep/"
    MODEL_PATH="$HOME/.config/semantic-grep/models/googlenews-slim/GoogleNews-vectors-negative300-SLIM.bin"
else
    echo "[*] Skipping model installation"
    echo "[OK] Skipped model installation"
fi

# Defining model path and writing to configuration based on their paths
echo "[*] Setting model path"
CONFIG_STRING="{\"model_path\":\"$MODEL_PATH\"}"
echo $CONFIG_STRING > "$CONFIG_PATH"
if [ $? -ne 0 ]; then
    echo "Failed to set model path"
    exit 1
fi
echo "[OK] Model path set:"
cat "$CONFIG_PATH"

# Testing
echo "[*] Testing"
$INSTALL_PATH -h
if [ $? -ne 0 ]; then
    echo "Failed to test sgrep"
    exit 1
fi
echo "[OK] sgrep tested and working. Installation complete."