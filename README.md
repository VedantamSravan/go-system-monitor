
# Go System Monitoring Application

This Go application monitors system resources such as CPU temperature, fan speeds, CPU usage, memory usage, and disk usage. It sends an email alert if any of these resources exceed predefined thresholds. The SMTP configuration for sending emails is read from a configuration file (`config.json`).

## Features

- **CPU Temperature**: Monitors CPU temperature and checks if it falls within the safe range (80°C to 90°C).
- **Fan Speed**: Monitors fan speed and checks if it is within the safe range (3500 RPM to 5000 RPM).
- **CPU Clock Speed**: Monitors the CPU clock speed and checks if it is greater than 3.20 GHz.
- **CPU Usage**: Monitors the usage of each CPU core, ensuring it doesn't exceed 80%.
- **Memory Usage**: Monitors system memory usage, alerting if it exceeds 80%.
- **Disk Usage**: Monitors disk usage, alerting if it exceeds 50%.
- **Email Alerts**: Sends an email alert if any threshold is exceeded.

## Requirements

- Go 1.18+ 
- `github.com/shirou/gopsutil` for system monitoring
- A working SMTP server (e.g., Gmail) for sending email alerts

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/VedantamSravan/go-system-monitor.git
   ```

2. Navigate into the project directory:
   ```bash
   cd go-system-monitor
   ```

3. Install the dependencies:
   ```bash
   go mod tidy
   ```

4. Create a `config.json` file for SMTP server configuration.

## Configuration

Create a `config.json` file in the root directory of the project. Example configuration:

```json
{
  "smtp_host": "smtp.gmail.com",
  "smtp_port": "587",
  "from_email": "your-email@gmail.com",
  "email_password": "your-email-password",
  "to_email": "receiver-email@example.com"
}
```

- `smtp_host`: The host of your SMTP server (e.g., `smtp.gmail.com` for Gmail).
- `smtp_port`: The SMTP port (usually `587` for TLS).
- `from_email`: Your email address (used to send alerts).
- `email_password`: Your email password (or App Password for Gmail).
- `to_email`: The email address where alerts will be sent.

> **Important**: If using Gmail, ensure that "Less secure apps" is enabled or generate an App Password for added security.

## Usage

To start the monitoring application, run the following command:

```bash
go run main.go
```

The application will monitor your system and send email alerts if any of the thresholds are exceeded. 

### Example Output

- **CPU Temperature Alert**:
  ```
  Alert: CPU Temperature is out of safe range: 95.00°C
  ```

- **CPU Usage Alert**:
  ```
  Alert: CPU Core 0 usage is above 80%: 85.00%
  ```

- **Disk Usage Alert**:
  ```
  Alert: Disk usage is above 50%: 55.00%
  ```

### Example Safe Output

If everything is within safe limits, the application will print the status like:

```
CPU Temperature: 75.00°C (Safe)
Fan Speed: 4200 RPM (Safe)
CPU Clock Speed: 3.50 GHz (Safe)
CPU Core 0 usage: 45.00% (Safe)
Memory usage: 60.00% (Safe)
Disk usage: 40.00% (Safe)
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgements

- [gopsutil](https://github.com/shirou/gopsutil) - For system monitoring functionality.
- [Go](https://golang.org/) - The programming language used to build this application.

## Contributing

Feel free to fork the repository and submit pull requests. Contributions are welcome!
