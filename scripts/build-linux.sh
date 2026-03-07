#!/bin/bash

echo -e "Start running the script..."
cd ../

echo -e "Checking Linux dependencies..."

# Check for required packages
check_dependency() {
    if ! pkg-config --exists "$1" 2>/dev/null; then
        echo -e "\033[31mError: $1 is not installed.\033[0m"
        echo -e "Install it with: sudo apt install $2"
        return 1
    fi
    return 0
}

check_dependency "gtk+-3.0" "libgtk-3-dev" || exit 1

# Ubuntu 24.04+ usa webkit2gtk-4.1, versiones anteriores usan 4.0
if pkg-config --exists "webkit2gtk-4.1" 2>/dev/null; then
    echo -e "Found webkit2gtk-4.1"
elif pkg-config --exists "webkit2gtk-4.0" 2>/dev/null; then
    echo -e "Found webkit2gtk-4.0"
else
    echo -e "\033[31mError: webkit2gtk not found.\033[0m"
    echo -e "Install with: sudo apt install libwebkit2gtk-4.1-dev"
    exit 1
fi

echo -e "\033[32mAll dependencies satisfied!\033[0m"

echo -e "Start building the app for Linux platform..."

# Detectar versión de webkit2gtk y usar el tag correcto
if pkg-config --exists "webkit2gtk-4.1" 2>/dev/null; then
    echo -e "Using webkit2gtk-4.1 (Ubuntu 24.04+)"
    wails build --clean --platform linux/amd64 -tags webkit2_41
else
    echo -e "Using webkit2gtk-4.0"
    wails build --clean --platform linux/amd64
fi

if [ $? -eq 0 ]; then
    echo -e "\033[32mBuild successful!\033[0m"
    echo -e "Output: build/bin/Sendlog-Syslog"
else
    echo -e "\033[31mBuild failed!\033[0m"
    exit 1
fi

echo -e "End running the script!"
