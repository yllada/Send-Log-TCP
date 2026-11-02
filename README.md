# SendLog

Send syslog messages to a remote server — manually or on a fixed interval — and verify they arrive in your SIEM.

- TCP or UDP, with optional TLS/mTLS (RFC 5425)
- RFC 5424 and RFC 3164 message formats
- TCP framing: octet-counting (RFC 6587) or non-transparent
- Import messages from a file (`.json`, `.csv`, `.txt`)
- Continuous sending at a fixed rate with live statistics

One binary, several ways to use it:

- **`sendlog`** (no arguments) opens the GTK4 desktop GUI on Linux; on other platforms (or without a display) it opens the terminal UI.
- **`sendlog tui`** forces the terminal UI anywhere.
- **`sendlog send|connect|stress [flags]`** runs in command-line mode for scripts and automation.

## Build

```bash
# Terminal UI + CLI (static, no dependencies)
go build -tags nocgo -o sendlog ./cmd/sendlog

# Full binary with the GTK4 GUI (linux; needs CGO + GTK4/libadwaita headers)
go build -o sendlog ./cmd/sendlog
```

The GUI is linux-only and is compiled in when CGO is available. On any other platform, or with `CGO_ENABLED=0`, `sendlog` simply starts the terminal UI instead.

## Terminal UI

Run `./sendlog tui` (or `./sendlog` on a system without the GTK4 GUI) and navigate the screens with `↑/↓` and `enter`; go back with `esc`; quit with `ctrl+c`. Move between fields with `tab`.

## Desktop GUI (GTK4, Linux)

Run `./sendlog` with a display. It opens a single screen with three framed sections — **Connection**, **Message**, **Stress** — plus a live status area. The **Connect** button keeps a TCP/TLS socket open; **Send messages** and **Start stress** then reuse that connection until you disconnect. Importing a file fills the message editor, and stress runs show live counters and a progress bar.

### 1. Connection

Set the server host, port, and protocol:

| Field | Notes |
|---|---|
| Host / Port | Server address. `514` is the usual plain port, `6514` for TLS. |
| Protocol | `TCP` or `UDP`. |
| Framing | TCP only. `octet-counting` is the recommended default. |
| TLS | `off` or `on` (TCP only). |
| Verify | Whether to validate the server certificate. Disable only for self-signed/test setups. |
| CA cert / Cert / Key | Paths for the CA bundle and, for mTLS, the client certificate and key. |

Select the **Test** action and press `t` to check the connection before sending. The screen shows a clear connected / failed result with the reason.

### 2. Messages

Choose the message settings and the messages to send:

| Field | Notes |
|---|---|
| Format | `rfc5424` (modern) or `rfc3164` (legacy). |
| Facility | 0–23 (e.g. `local0` = 16). |
| Severity | 0–7 (e.g. `info` = 6, `err` = 3). |
| Hostname / Appname | Source name and application tag. Leave hostname empty to use the local one. |
| Raw PRI | Prepend `<PRI>` to the message verbatim, with no header. For replaying captured vendor lines. |

Add several messages with `a`, edit the selected one, remove it with `x`.

### 3. Import File

Type a file path and press `enter`. The file is parsed (`.json` arrays or NDJSON, `.csv` with a `message`/`msg`/`log` column, `.txt` one message per line) and a preview is shown: total lines, valid messages, and any errors. Choose **Append** to add to the current list or **Replace** to swap it in.

### 4. Send

Review the message count and the target, then start the send. You get live progress and, when done, the list of sent messages and any errors with their reasons, plus a final `sent / failed` summary. This is where you confirm each message reached the server — check your SIEM for them.

### 5. Stress

Send the same message automatically at a fixed rate — for example one log per second:

| Field | Notes |
|---|---|
| Message | What to send each tick. Defaults to your first composed message. |
| Rate | Messages per second. |
| Duration | Seconds to keep sending. Empty runs until you stop it. |

Press start, watch the live counters (sent, errors, rate, elapsed), and press stop when you are done. A final summary is shown.

## CLI (automation)

The same binary also runs as a plain command-line client for scripts and CI. Progress and stats go to stderr; add `--json` for a single machine-readable line on stdout. Exit code is `0` on success, `1` on error.

```bash
# Send one message over TCP (octet-counting framing, RFC 5424)
./sendlog send -a 10.0.0.5 -p 514 -t tcp -m "deploy finished"

# Send messages from a log file (.json / .csv / .txt)
./sendlog send -a 10.0.0.5 -p 514 -t udp -f events.txt

# mTLS (RFC 5425) with a client certificate
./sendlog send -a 10.0.0.5 -p 6514 -t tcp --tls \
  --cert client.pem --key client.key -m "secret log"

# Test that a connection works
./sendlog connect -a 10.0.0.5 -p 514 -t tcp --json

# Continuous send: 50 msg/s for 60 seconds (Ctrl+C stops early)
./sendlog stress -a 10.0.0.5 -p 514 -t tcp -m "load test" --rate 50 --duration 60
```

Run `./sendlog help` for the full flag reference.

## Verify in your SIEM

1. Point Connection at your syslog server and test it.
2. Compose one easy-to-recognize message, e.g. `sendlog smoke test <timestamp>`.
3. Send it and confirm it appears in the SIEM with the expected facility, severity, hostname, and app name.
4. Optionally run Stress at 1 message/s and watch the events stream in.

## Reference

| RFC | Description |
|---|---|
| [5424](https://datatracker.ietf.org/doc/html/rfc5424) | The Syslog Protocol (modern) |
| [3164](https://datatracker.ietf.org/doc/html/rfc3164) | BSD Syslog Protocol (legacy) |
| [5425](https://datatracker.ietf.org/doc/html/rfc5425) | TLS Transport Mapping for Syslog |
| [6587](https://datatracker.ietf.org/doc/html/rfc6587) | Transmission of Syslog over TCP (octet-counting) |

## License

[Apache 2.0](LICENSE)
