# SendLog Syslog

A modern desktop application built with **Wails v2** for sending syslog messages to remote servers via TCP or UDP protocols.

## 🚀 Features

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
- **Desktop Optimized**: Fixed window size (950x720) for optimal UX
- **Fully Tested**: Comprehensive test suite with benchmarks

## 🎨 UI Components

The application features a clean, card-based interface with:
- Connection settings panel with real-time status
- Message configuration with facility and severity selectors
- Message composition area
- Toast notifications for feedback

## 🛠️ Tech Stack

### Backend
- **Go** - Application logic and syslog protocol implementation
- **Wails v2** - Go + Web frontend framework

### Frontend
- **Next.js** - React framework
- **TypeScript** - Type-safe development
- **shadcn/ui** - Modern UI components
- **Tailwind CSS** - Utility-first styling
- **React Hook Form** - Form management
- **Zod** - Schema validation

## 📦 Installation

### Prerequisites
- Go 1.21+
- Node.js 18+
- pnpm

### Setup

1. Clone the repository:
```bash
git clone https://github.com/yllada/Send-Log-TCP.git
cd Send-Log-TCP
```

2. Install Wails CLI:
```bash
./scripts/install-wails-cli.sh
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

## �📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 👤 Author

**Yadian Llada Lopez**
- Email: yadian.llada@gmail.com
- GitHub: [@yllada](https://github.com/yllada)

## 🤝 Contributing

Contributions, issues, and feature requests are welcome!
