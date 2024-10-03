
# Send Log TCP/UDP

Welcome to the **Send SysLog TCP/UDP** project! This project allows you to send log data over both TCP and UDP connections, providing a lightweight and flexible logging solution for your applications.

## Key Features

- **Dual Protocol Support**: Send log data over both TCP and UDP.
- **Simple Setup**: Easy to install and configure for quick integration.
- **Reliable (TCP)**: Ensures reliable delivery of log data with automatic retries.
- **Fast and Lightweight (UDP)**: For scenarios where speed is prioritized over reliability.
- **Customizable**: Easily adapt to different log formats and server configurations.
- **Command-Line Options**: Send logs via command-line arguments for flexibility.

## Installation

To install and set up the **Send Log TCP/UDP** tool, follow these steps:

1. **Clone the repository**:
    ```sh
    git clone https://github.com/yllada/Send-Log-TCP.git
    ```
2. **Navigate to the project directory**:
    ```sh
    cd Send-Log-TCP
    ```
3. **Install the Go dependencies**:
    ```sh
    go mod tidy
    ```

## Usage

### Running the TCP Log Sender

To send log data over a TCP connection, run the following command:

```bash
go run main.go -protocol="tcp" -message="Your log message here"
```

### Running the UDP Log Sender

To send log data over a UDP connection, you can specify the protocol as follows:

```bash
go run main.go -protocol="udp" -message="Your log message here"
```

## Configuration

The configuration for this tool can be customized through flags variables:

| Variable                          | Description                                    | Example Value   |
|-----------------------------------|------------------------------------------------|-----------------|
| `address`         | The IP address and port where logs will be sent. | `127.0.0.1:7003`     |
| `protocol`          | The protocol to use for sending logs (`tcp` or `udp`). | `tcp`                |
| `facility`          | The Sylog facility where logs will be sent (`Local0` or `Local1` or `Local2` or `Local3` or `Local4` or `Local5` or `Local6` or `Local7`). | `Local0`                |
| `severity`          | The Syslog severity to use for sending logs (`Emergency` or `Alert` or `Critical` or `Error` or `Warning` or `Notice` or `Info` or `Debug`). | `Info`                |
| `hostname`          | The hostname to use for sending logs. | `HostName`                |
| `messages`          | The Syslog messages to use for sending logs. | `Comma-separated log messages`                |
| `interval`          | Interval for sending logs. | `(e.g., 5s)`                |

## Advanced Configuration

You can also adjust other settings, such as connection timeout, maximum retries (for TCP), and log format by extending the configuration. Future versions will include more advanced configuration options, such as batching logs and asynchronous sending.

## Contributing

We welcome contributions to the **Send Log TCP/UDP** project! If you have feature requests, ideas, or encounter bugs, feel free to open an issue or submit a pull request. Here's how you can contribute:

1. Fork the repository.
2. Create a feature branch (`git checkout -b feature/new-feature`).
3. Commit your changes (`git commit -m 'Add new feature'`).
4. Push the branch (`git push origin feature/new-feature`).
5. Open a pull request.

## License

This project is licensed under the Apache License 2.0. See the [LICENSE](LICENSE) file for more details.

## Contact

For any questions or inquiries, feel free to reach out via email at [yadian.llada@gmail.com](mailto:yadian.llada@gmail.com).

Thank you for using **Send Log TCP/UDP**!
