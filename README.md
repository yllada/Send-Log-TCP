# Send Log TCP

Welcome to the Send Log TCP project! This project is designed to facilitate the sending of log data over TCP connections. It is a simple yet powerful tool for developers who need to transmit log information between systems.

## Features

- **Easy to Use**: Simple setup and configuration.
- **Reliable**: Ensures log data is sent reliably over TCP.
- **Flexible**: Can be integrated into various applications and systems.
- **Lightweight**: Minimal resource usage.

## Installation

To install the Send Log TCP tool, follow these steps:

1. Clone the repository:
    ```sh
    git clone https://github.com/yourusername/send-log-tcp.git
    ```
2. Navigate to the project directory:
    ```sh
    cd send-log-tcp
    ```
3. Install the required dependencies:
    ```sh
    npm install
    ```

## Usage

To use the Send Log TCP tool, follow these steps:

1. Start the TCP server:
    ```sh
    npm start
    ```
2. Configure your application to send log data to the TCP server.

## Configuration

You can configure the TCP server by editing the `.env.example` file and rename `.env`. Here are some of the key settings:

- `LOG_AGENT_ADDRESS`: The direcctions and port on which the TCP server will listen.
- `LOG_MESSAGE`: The log message.

Example `config.json`:
```go
    LOG_AGENT_ADDRESS=127.0.0.1:7003
    LOG_MESSAGE=log_message
```

## Contributing

We welcome contributions to the Send Log TCP project! If you have any ideas, suggestions, or bug reports, please open an issue or submit a pull request.

## License

This project is licensed under the Apache License. See the [LICENSE](LICENSE) file for more details.

## Contact

For any questions or inquiries, please contact us at [yadian.llada@gmail.com](mailto:yadian.llada@gmail.com).

Thank you for using Send Log TCP!