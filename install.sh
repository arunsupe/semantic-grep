#!/bin/bash

# golang check for version specified in go.mod
echo "[*] Checking golang"
go version
if [ $? -ne 0 ]; then
    echo "Failed to check golang version: do you have go installed?"
    exit 1
fi
echo "[OK] golang installed"

# Building and installing w2vgrep
echo "[*] Building and installing w2vgrep"
go  build -o w2vgrep
if [ $? -ne 0 ]; then
    echo "Failed to build w2vgrep"
    exit 1
fi
echo "[OK] w2vgrep built"

# Prompt to either move to /usr/bin/w2vgrep, adding local path to $PATH or doing nothing
echo "[*] Where would you like to install w2vgrep?"
echo "1. /usr/bin/w2vgrep (will require sudo)"
echo "2. Add local path to \$PATH"
echo "3. Do nothing"
read -p "Enter your choice: " choice

INSTALL_PATH="$(pwd)/w2vgrep"
if [ $choice -eq 1 ]; then
    echo "[*] Installing w2vgrep in /usr/bin/w2vgrep, please enter your password as sudo is required."
    sudo cp w2vgrep /usr/bin/w2vgrep
    if [ $? -ne 0 ]; then
        echo "Failed to install w2vgrep"
        exit 1
    fi
    INSTALL_PATH="/usr/bin/w2vgrep"
    echo "[OK] w2vgrep installed in /usr/bin/w2vgrep"
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
    echo "Failed to test w2vgrep"
    exit 1
fi
echo "[OK] w2vgrep tested and working. Installation complete."