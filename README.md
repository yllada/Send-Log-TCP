
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
3. **Compile the fontend using npm or pnpm**:
    ```sh
    cd .\frontend\
    ```
    ```sh
    pnpm run build or npm run build
    ```

## Usage

### Running the TCP Log Sender

```sh
    go run main.go
```

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
