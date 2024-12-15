package main

import (
	"encoding/json"
	"fmt"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
	"io/ioutil"
	"log"
	"net/smtp"
	"os/exec"
	"strings"
)

// SMTPConfig holds the SMTP server configuration
type SMTPConfig struct {
	SMTPHost      string `json:"smtp_host"`
	SMTPPort      string `json:"smtp_port"`
	FromEmail     string `json:"from_email"`
	EmailPassword string `json:"email_password"`
	ToEmail       string `json:"to_email"`
}

// Send email function
func sendEmail(config SMTPConfig, subject, body string) {
	// Email content
	subjectLine := "Subject: " + subject + "\n"
	message := []byte(subjectLine + "\n" + body)

	// Set up authentication information.
	auth := smtp.PlainAuth("", config.FromEmail, config.EmailPassword, config.SMTPHost)

	// Send the email
	err := smtp.SendMail(config.SMTPHost+":"+config.SMTPPort, auth, config.FromEmail, []string{config.ToEmail}, message)
	if err != nil {
		log.Fatalf("Error sending email: %v\n", err)
	} else {
		fmt.Println("Alert email sent successfully!")
	}
}

// ReadSMTPConfig reads the SMTP configuration from a file
func ReadSMTPConfig(filePath string) (SMTPConfig, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return SMTPConfig{}, fmt.Errorf("could not read config file: %w", err)
	}

	var config SMTPConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		return SMTPConfig{}, fmt.Errorf("could not parse config file: %w", err)
	}

	return config, nil
}

// GetFanSpeeds returns the fan speeds using the 'sensors' command on Linux
func GetFanSpeeds() (string, error) {
	// Run the 'sensors' command (make sure lm-sensors is installed)
	cmd := exec.Command("sensors")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Error fetching fan speeds: %w", err)
	}
	return string(output), nil
}

func main() {
	// Read SMTP configuration from config file
	config, err := ReadSMTPConfig("config.json")
	if err != nil {
		log.Fatalf("Error reading SMTP config: %v\n", err)
	}

	// Thresholds
	const (
		maxTemp            = 90.0 // Max temperature in °C
		minTemp            = 80.0 // Min temperature in °C
		minFanSpeed        = 3500 // Min fan speed in RPM
		maxFanSpeed        = 5000 // Max fan speed in RPM
		maxClockSpeed      = 3.20 // Max clock speed in GHz
		cpuUsageThreshold  = 80.0 // Max CPU usage in %
		memUsageThreshold  = 80.0 // Max memory usage in %
		diskUsageThreshold = 50.0 // Max disk usage in %
	)

	alertMessage := ""

	// Monitor CPU Temperature (using sensors command for Linux)
	temps, err := GetCPUTemperature()
	if err != nil {
		log.Fatalf("Error fetching CPU temperature: %v\n", err)
	}
	if temps > maxTemp || temps < minTemp {
		alertMessage += fmt.Sprintf("Alert: CPU Temperature is out of safe range: %.2f°C\n", temps)
	} else {
		fmt.Printf("CPU Temperature: %.2f°C (Safe)\n", temps)
	}

	// Monitor Fan Speeds (using external sensors command)
	fanSpeeds, err := GetFanSpeeds()
	if err != nil {
		log.Fatalf("Error fetching fan speeds: %v\n", err)
	}
	// Checking if fan speed data is in range
	if strings.Contains(fanSpeeds, "fan1") {
		alertMessage += fmt.Sprintf("Fan speed info:\n%s\n", fanSpeeds)
	}

	// Monitor CPU Clock Speed (using CPU Info method)
	clockSpeeds, err := cpu.Info()
	if err != nil {
		log.Fatalf("Error fetching CPU clock speed: %v\n", err)
	}
	for _, cpuInfo := range clockSpeeds {
		// Assuming the CPU has a frequency field available
		if cpuInfo.Mhz/1000.0 < maxClockSpeed {
			alertMessage += fmt.Sprintf("Alert: CPU Clock Speed is below 3.20 GHz: %.2f GHz\n", cpuInfo.Mhz/1000.0)
		} else {
			fmt.Printf("CPU Clock Speed: %.2f GHz (Safe)\n", cpuInfo.Mhz/1000.0)
		}
	}

	// Monitor CPU Usage
	cpuUsage, err := cpu.Percent(0, true)
	if err != nil {
		log.Fatalf("Error fetching CPU usage: %v\n", err)
	}
	for i, usage := range cpuUsage {
		if usage > cpuUsageThreshold {
			alertMessage += fmt.Sprintf("Alert: CPU Core %d usage is above 80%%: %.2f%%\n", i, usage)
		} else {
			fmt.Printf("CPU Core %d usage: %.2f%% (Safe)\n", i, usage)
		}
	}

	// Monitor Memory Usage
	memStats, err := mem.VirtualMemory()
	if err != nil {
		log.Fatalf("Error fetching memory stats: %v\n", err)
	}
	if memStats.UsedPercent > memUsageThreshold {
		alertMessage += fmt.Sprintf("Alert: Memory usage is above 80%%: %.2f%%\n", memStats.UsedPercent)
	} else {
		fmt.Printf("Memory usage: %.2f%% (Safe)\n", memStats.UsedPercent)
	}

	// Monitor Disk Usage
	diskStats, err := disk.Usage("/")
	if err != nil {
		log.Fatalf("Error fetching disk usage: %v\n", err)
	}
	if diskStats.UsedPercent > diskUsageThreshold {
		alertMessage += fmt.Sprintf("Alert: Disk usage is above 50%%: %.2f%%\n", diskStats.UsedPercent)
	} else {
		fmt.Printf("Disk usage: %.2f%% (Safe)\n", diskStats.UsedPercent)
	}

	// Send an email if any alert message exists
	if alertMessage != "" {
		sendEmail(config, "System Alert: Resource Usage Exceeded", alertMessage)
	}
}

// GetCPUTemperature uses the 'sensors' command for Linux to fetch CPU temperature
//func GetCPUTemperature() (float64, error) {
//	// Run the 'sensors' command
//	cmd := exec.Command("osx-cpu-temp") //for mac  brew install osx-cpu-temp
//
//	//for linux lm-sensors
//	//for windows wmic
//	output, err := cmd.Output()
//	if err != nil {
//		return 0, fmt.Errorf("Error fetching CPU temperature: %w", err)
//	}
//
//	// Parse the output to find the temperature
//	for _, line := range strings.Split(string(output), "\n") {
//		if strings.Contains(line, "Core 0") {
//			// Example: Core 0:      +45.0°C  (high = +80.0°C, crit = +100.0°C)
//			parts := strings.Fields(line)
//			if len(parts) > 1 {
//				// Convert temperature to float64
//				var temp float64
//				_, err := fmt.Sscanf(parts[1], "%f", &temp)
//				if err == nil {
//					return temp, nil
//				}
//			}
//		}
//	}
//
//	return 0, fmt.Errorf("could not find CPU temperature")
//}

// GetCPUTemperature uses the 'osx-cpu-temp' command for macOS to fetch CPU temperature
func GetCPUTemperature() (float64, error) {
	// Run the 'osx-cpu-temp' command
	cmd := exec.Command("osx-cpu-temp")
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("Error fetching CPU temperature: %w", err)
	}

	// Print the raw output for debugging
	fmt.Printf("Raw output: %s\n", string(output))

	// Proceed with parsing the output
	var temp float64
	_, err = fmt.Sscanf(string(output), "+%f°C", &temp)
	if err != nil {
		return 0, fmt.Errorf("Error parsing CPU temperature: %w", err)
	}

	return temp, nil

}
