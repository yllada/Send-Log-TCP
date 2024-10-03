
# Send Log TCP/UDP

Welcome to the **Send Log TCP/UDP** project! This project allows you to send log data over both TCP and UDP connections, providing a lightweight and flexible logging solution for your applications.

## Key Features

- **Dual Protocol Support**: Send log data over both TCP and UDP.
- **Simple Setup**: Easy to install and configure for quick integration.
- **Reliable (TCP)**: Ensures reliable delivery of log data with automatic retries.
- **Fast and Lightweight (UDP)**: For scenarios where speed is prioritized over reliability.
- **Customizable**: Easily adapt to different log formats and server configurations.

## Installation

To install and set up the **Send Log TCP/UDP** tool, follow these steps:

1. **Clone the repository**:
    ```sh
    git clone https://github.com/yllada/send-log-tcp.git
    ```
2. **Navigate to the project directory**:
    ```sh
    cd send-log-tcp
    ```
3. **Install the Go dependencies**:
    ```sh
    go mod tidy
    ```

## Usage

### Running the TCP Log Sender

To send log data over a TCP connection, run the following command:

```bash
go run main.go --protocol tcp
```

### Running the UDP Log Sender

To send log data over a UDP connection, you can specify the protocol as follows:

```bash
go run main.go --protocol udp
```

### Example Code Snippet

Here's an example of how you can send logs using this tool in Go:

```go
package main

import (
    "fmt"
    "net"
    "os"
)

func main() {
    // Get the log agent address and message from the environment variables
    logAgentAddress := os.Getenv("LOG_AGENT_ADDRESS")
    logMessage := os.Getenv("LOG_MESSAGE")

    // Resolve TCP address and establish connection
    conn, err := net.Dial("tcp", logAgentAddress)
    if err != nil {
        fmt.Println("Error connecting:", err)
        return
    }
    defer conn.Close()

    // Send log message
    _, err = conn.Write([]byte(logMessage))
    if err != nil {
        fmt.Println("Error sending log:", err)
    } else {
        fmt.Println("Log sent successfully!")
    }
}
```

## Configuration

The configuration for this tool can be customized through environment variables. Update the `.env` file in the root of your project to adjust these settings:

| Variable           | Description                                    | Example Value        |
|--------------------|------------------------------------------------|----------------------|
| `LOG_AGENT_ADDRESS` | The IP address and port where logs will be sent. | `127.0.0.1:7003`     |
| `LOG_MESSAGE`       | The log message that you want to send.         | `"This is a log"`    |
| `PROTOCOL`          | The protocol to use for sending logs (`tcp` or `udp`). | `tcp`                |

### Sample `.env` File:

```bash
LOG_AGENT_ADDRESS=127.0.0.1:7003
LOG_MESSAGE="Application started successfully"
PROTOCOL=tcp
```

Make sure to rename the `.env.example` file to `.env` and set the appropriate values based on your environment.

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
