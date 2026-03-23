# SendLog Syslog

<div align="center">

[![CI](https://github.com/Y3LCorp/Send-Log-TCP/actions/workflows/ci.yml/badge.svg)](https://github.com/Y3LCorp/Send-Log-TCP/actions/workflows/ci.yml)
[![Release](https://github.com/Y3LCorp/Send-Log-TCP/actions/workflows/release.yml/badge.svg)](https://github.com/Y3LCorp/Send-Log-TCP/actions/workflows/release.yml)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![GitHub release](https://img.shields.io/github/v/release/Y3LCorp/Send-Log-TCP)](https://github.com/Y3LCorp/Send-Log-TCP/releases/latest)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20macOS%20%7C%20Linux-blue)](https://github.com/Y3LCorp/Send-Log-TCP/releases)

**A modern, cross-platform desktop application for sending syslog messages to remote servers via TCP or UDP.**

Built with [Wails v2](https://wails.io/) (Go + Web Technologies) | RFC 5424 & RFC 3164 Compliant

[Download](#-download) • [Features](#-features) • [Documentation](#-documentation) • [Contributing](#-contributing)

</div>

---

## 📋 Table of Contents

- [Features](#-features)
- [Download](#-download)
- [Quick Start](#-quick-start)
- [Usage](#-usage)
- [Development](#-development)
- [Documentation](#-documentation)
- [Contributing](#-contributing)
- [Security](#-security)
- [License](#-license)

---

## ✨ Features

- **Multi-Protocol Support**: Send logs via TCP or UDP
- **TLS/SSL Encryption**: Secure syslog transmission over TLS (RFC 5425)
  - Support for self-signed certificates
  - Configurable certificate verification
  - TLS 1.2 and 1.3 support
- **RFC Compliance**: Supports both RFC 5424 (modern) and RFC 3164 (legacy) formats
- **TCP Framing** (RFC 6587): 
  - ⭐ **Octet Counting** (recommended): Efficient length-prefixed framing
  - **Non-Transparent Framing**: LF-delimited for legacy compatibility
- **Professional Implementation**:
  - Modular architecture with separated concerns
  - Efficient memory usage with pre-calculated buffers
  - Comprehensive validation (UTF-8, message length, framing rules)
  - Robust error handling with detailed messages
- **Modern UI**: Built with Next.js, React, and shadcn/ui components
- **Dark/Light Theme**: System-aware theme switching
- **Real-time Connection**: Test and maintain persistent connections
- **Cross-Platform**: Windows, macOS (Intel & Apple Silicon), and Linux
- **Fully Tested**: Comprehensive test suite with benchmarks

### Tech Stack

| Layer | Technologies |
|-------|-------------|
| **Backend** | Go 1.22+, Wails v2 |
| **Frontend** | Next.js, React, TypeScript |
| **UI** | shadcn/ui, Tailwind CSS |
| **Forms** | React Hook Form, Zod |

---

## 📥 Download

Pre-built binaries are available in the [Releases](https://github.com/Y3LCorp/Send-Log-TCP/releases) page.

### Windows
Download `Sendlog-Syslog-windows-amd64.exe` and run it directly.

### macOS
Download the appropriate version:
- **Apple Silicon (M1/M2/M3/M4):** `Sendlog-Syslog-darwin-arm64`
- **Intel:** `Sendlog-Syslog-darwin-amd64`

### Linux

<details>
<summary><b>DEB Package (Recommended for Debian/Ubuntu)</b></summary>

```bash
# Download and install - automatically handles dependencies
sudo dpkg -i sendlog-syslog_*_amd64.deb
sudo apt-get install -f  # Install missing dependencies if any
```
</details>

<details>
<summary><b>AppImage (Portable)</b></summary>

```bash
chmod +x Sendlog-Syslog-*-x86_64.AppImage
./Sendlog-Syslog-*-x86_64.AppImage
```
</details>

<details>
<summary><b>Raw Binary</b></summary>

```bash
# Install dependencies first
sudo apt install -y libgtk-3-0 libwebkit2gtk-4.1-0  # Ubuntu 24.04+
# or
sudo apt install -y libgtk-3-0 libwebkit2gtk-4.0-37  # Ubuntu 22.04

# Then run the binary
chmod +x Sendlog-Syslog-linux-amd64
./Sendlog-Syslog-linux-amd64
```
</details>

---

## 🚀 Quick Start

1. **Download** the appropriate binary for your platform from [Releases](https://github.com/Y3LCorp/Send-Log-TCP/releases/latest)
2. **Run** the application
3. **Configure** your syslog server connection:
   - Enter the server IP address
   - Set the port (default: `514` for plain, `6514` for TLS)
   - Select protocol (TCP or UDP)
4. **Send** your log messages!

---

## 📖 Usage

### Connection Configuration

| Setting | Description | Default |
|---------|-------------|---------|
| **IP Address** | Syslog server address (IPv4 or IPv6) | - |
| **Port** | Server port | 514 (plain) / 6514 (TLS) |
| **Protocol** | TCP or UDP | TCP |
| **TLS/SSL** | Enable encrypted connection (TCP only) | Off |
| **Verify Certificate** | Validate server certificate | On |

### Message Configuration

| Setting | Description | Options |
|---------|-------------|---------|
| **RFC Format** | Syslog message format | RFC 5424 (modern) / RFC 3164 (legacy) |
| **Facility** | Message category | 0-23 (e.g., 16 = local0) |
| **Severity** | Message priority | 0-7 (e.g., 6 = informational) |
| **Hostname** | Source hostname | Auto-detected or custom |
| **App Name** | Application identifier | Custom |

### 🔒 TLS/SSL Security

| Feature | Description |
|---------|-------------|
| **Port 6514** | Standard port for syslog over TLS (RFC 5425) |
| **Certificate Verification** | ✅ Enabled: Validates using system CA (production) |
| | ⚠️ Disabled: Accepts self-signed certs (testing) |
| **TLS Versions** | TLS 1.2 and TLS 1.3 supported |

> **Note:** TLS is only available for TCP connections. UDP does not support encryption.

---

## 🛠️ Development
```

3. Install frontend dependencies:
```bash
cd frontend
pnpm install
```

## 🚀 Development

Run in development mode:
```bash
wails dev
```

## 🏗️ Building

### Linux (Ubuntu/Debian)
```bash
# First time: Install dependencies
./scripts/install-linux-deps.sh

# Build
./scripts/build-linux.sh
```

### Windows
```bash
./scripts/build-windows.sh
```

### macOS (ARM)
```bash
./scripts/build-macos-arm.sh
```

### macOS (Intel)
```bash
./scripts/build-macos-intel.sh
```

## 📝 Usage

1. **Configure Connection**:
   - Enter server IP address
   - Set port (default: 514 for plain, 6514 for TLS)
   - Select protocol (TCP/UDP)
   - **TLS Options** (TCP only):
     - Enable "Use TLS/SSL" for encrypted connections
     - Disable "Verify Certificate" to accept self-signed certificates
   - Connect to server

2. **Configure Message Format**:
   - Choose RFC format (5424 or 3164)
   - Set facility (0-23)
   - Set severity (0-7)
   - Optional: custom hostname and app name

3. **Send Messages**:
   - Enter log messages (one per line)
   - Click "Send Syslog Messages"

### 🔒 TLS/SSL Security Notes

- **Port 6514**: Standard port for syslog over TLS (RFC 5425)
- **Certificate Verification**: 
  - ✅ **Enabled**: Verifies server certificate using system CA (recommended for production)
  - ⚠️ **Disabled**: Accepts self-signed certificates (useful for testing/development)
- **TLS Versions**: Supports TLS 1.2 and TLS 1.3
- **UDP**: TLS is only available for TCP connections

## 🔧 Configuration

### Window Settings (main.go)
```go
Width:     900  // Fixed width
Height:    780  // Fixed height
MinWidth:  900
MinHeight: 780
MaxWidth:  1100
MaxHeight: 900
```

### Syslog Facilities
- 0-15: System facilities
- 16-23: Local use (local0-local7)

### Severity Levels
- 0: Emergency
- 1: Alert
- 2: Critical
- 3: Error
- 4: Warning
- 5: Notice
- 6: Informational
- 7: Debug

## � Technical Documentation

For detailed technical documentation about the RFC 6587 implementation, see [TECHNICAL_DOCUMENTATION.md](TECHNICAL_DOCUMENTATION.md).

## 🧪 Testing

Run the test suite:
```bash
go test -v
```

Run benchmarks:
```bash
go test -bench=. -benchmem
```

---

## 📚 Documentation

- **[Technical Documentation](TECHNICAL_DOCUMENTATION.md)** - RFC 6587 implementation details
- **[Changelog](CHANGELOG.md)** - Version history and release notes
- **[License](LICENSE)** - Apache 2.0 License

### Reference Standards

| RFC | Description |
|-----|-------------|
| [RFC 5424](https://datatracker.ietf.org/doc/html/rfc5424) | The Syslog Protocol |
| [RFC 3164](https://datatracker.ietf.org/doc/html/rfc3164) | BSD Syslog Protocol (legacy) |
| [RFC 5425](https://datatracker.ietf.org/doc/html/rfc5425) | TLS Transport Mapping for Syslog |
| [RFC 6587](https://datatracker.ietf.org/doc/html/rfc6587) | Transmission of Syslog Messages over TCP |

---

## 🤝 Contributing

Contributions are welcome! Here's how you can help:

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Commit** your changes using [Conventional Commits](https://www.conventionalcommits.org/):
   - `feat:` new feature
   - `fix:` bug fix
   - `docs:` documentation
   - `refactor:` code refactoring
4. **Push** to your branch (`git push origin feature/amazing-feature`)
5. **Open** a Pull Request

### Reporting Issues

Found a bug or have a feature request? [Open an issue](https://github.com/Y3LCorp/Send-Log-TCP/issues/new/choose)!

---

## � Security

Security is a top priority for SendLog Syslog. Key security features:

- **TLS 1.2/1.3** encryption support
- **mTLS** (Mutual TLS) for client authentication
- **Certificate verification** enabled by default
- **No telemetry** or data collection

For reporting security vulnerabilities, please see our [Security Policy](SECURITY.md).

---

## �📄 License

This project is licensed under the **Apache License 2.0** - see the [LICENSE](LICENSE) file for details.

---

## � Authors

<table>
  <tr>
    <td align="center">
      <a href="https://github.com/yllada">
        <img src="https://github.com/yllada.png" width="100px;" alt="Yadian Llada Lopez"/>
        <br />
        <sub><b>Yadian Llada Lopez</b></sub>
      </a>
      <br />
      <a href="mailto:yadian.llada@gmail.com">📧 Email</a>
    </td>
    <td align="center">
      <a href="https://github.com/JocLRojas">
        <img src="https://github.com/JocLRojas.png" width="100px;" alt="JocLRojas"/>
        <br />
        <sub><b>JocLRojas</b></sub>
      </a>
    </td>
  </tr>
</table>

---

<div align="center">

**⭐ If you find this project useful, please consider giving it a star!**

Made with ❤️ using [Wails](https://wails.io/) and [Go](https://go.dev/)

</div>
