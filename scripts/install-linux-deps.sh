#!/bin/bash

# Script to install Linux dependencies for Wails
# Supports: Ubuntu/Debian, Fedora, Arch Linux

echo -e "\033[34m========================================\033[0m"
echo -e "\033[34m  SendLog Syslog - Linux Dependencies  \033[0m"
echo -e "\033[34m========================================\033[0m"

# Detect distribution
if [ -f /etc/os-release ]; then
    . /etc/os-release
    DISTRO=$ID
else
    echo -e "\033[31mCannot detect distribution.\033[0m"
    exit 1
fi

echo -e "Detected distribution: \033[32m$DISTRO\033[0m"

case $DISTRO in
    ubuntu|debian|linuxmint|pop)
        echo -e "Installing dependencies for Ubuntu/Debian..."
        sudo apt update
        # Detectar versión de Ubuntu para elegir el paquete correcto
        if apt-cache show libwebkit2gtk-4.1-dev &>/dev/null; then
            WEBKIT_PKG="libwebkit2gtk-4.1-dev"
        else
            WEBKIT_PKG="libwebkit2gtk-4.0-dev"
        fi
        echo -e "Using WebKit package: $WEBKIT_PKG"
        sudo apt install -y \
            libgtk-3-dev \
            $WEBKIT_PKG \
            build-essential \
            pkg-config \
            libappindicator3-dev
        ;;
    fedora)
        echo -e "Installing dependencies for Fedora..."
        sudo dnf install -y \
            gtk3-devel \
            webkit2gtk3-devel \
            gcc \
            pkg-config
        ;;
    arch|manjaro)
        echo -e "Installing dependencies for Arch Linux..."
        sudo pacman -S --noconfirm \
            gtk3 \
            webkit2gtk \
            base-devel \
            pkgconf
        ;;
    opensuse*|suse)
        echo -e "Installing dependencies for openSUSE..."
        sudo zypper install -y \
            gtk3-devel \
            webkit2gtk3-soup2-devel \
            gcc \
            pkg-config
        ;;
    *)
        echo -e "\033[33mUnknown distribution: $DISTRO\033[0m"
        echo -e "Please manually install:"
        echo -e "  - GTK 3 development libraries"
        echo -e "  - WebKitGTK 4.0 development libraries"
        echo -e "  - GCC/build-essential"
        echo -e "  - pkg-config"
        exit 1
        ;;
esac

# Verify installation
echo -e "\n\033[34mVerifying installation...\033[0m"

verify_pkg() {
    if pkg-config --exists "$1" 2>/dev/null; then
        echo -e "  ✓ $1: \033[32mInstalled\033[0m"
        return 0
    else
        echo -e "  ✗ $1: \033[31mNot found\033[0m"
        return 1
    fi
}

ALL_OK=true
verify_pkg "gtk+-3.0" || ALL_OK=false
# Verificar webkit2gtk-4.1 (Ubuntu 24.04+) o 4.0 (versiones anteriores)
if pkg-config --exists "webkit2gtk-4.1" 2>/dev/null; then
    verify_pkg "webkit2gtk-4.1" || ALL_OK=false
else
    verify_pkg "webkit2gtk-4.0" || ALL_OK=false
fi

if [ "$ALL_OK" = true ]; then
    echo -e "\n\033[32m✓ All dependencies installed successfully!\033[0m"
    echo -e "You can now build with: ./scripts/build-linux.sh"
else
    echo -e "\n\033[31m✗ Some dependencies are missing.\033[0m"
    exit 1
fi
