#!/bin/bash

echo -e "Start running the script..."
cd ../

echo -e "Checking Linux dependencies..."

# Check for required packages
check_dependency() {
    if ! pkg-config --exists "$1" 2>/dev/null; then
        echo -e "\033[31mError: $1 is not installed.\033[0m"
        echo -e "Install it with: sudo apt install $2"
        exit 1
    fi
}

check_dependency "gtk+-3.0" "libgtk-3-dev"
check_dependency "webkit2gtk-4.0" "libwebkit2gtk-4.0-dev"

echo -e "\033[32mAll dependencies satisfied!\033[0m"

echo -e "Start building the app for Linux platform..."
wails build --clean --platform linux/amd64

if [ $? -eq 0 ]; then
    echo -e "\033[32mBuild successful!\033[0m"
    echo -e "Output: build/bin/Sendlog-Syslog"
else
    echo -e "\033[31mBuild failed!\033[0m"
    exit 1
fi

echo -e "End running the script!"
