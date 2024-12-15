package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/smtp"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/sensors"
)

// SMTPConfig holds the SMTP server configuration
type SMTPConfig struct {
	SMTPHost     string `json:"smtp_host"`
	SMTPPort     string `json:"smtp_port"`
	FromEmail    string `json:"from_email"`
	EmailPassword string `json:"email_password"`
	ToEmail      string `json:"to_email"`
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

func main() {
	// Read SMTP configuration from config file
	config, err := ReadSMTPConfig("config.json")
	if err != nil {
		log.Fatalf("Error reading SMTP config: %v\n", err)
	}

	// Thresholds
	const (
		maxTemp            = 90.0   // Max temperature in °C
		minTemp            = 80.0   // Min temperature in °C
		minFanSpeed        = 3500   // Min fan speed in RPM
		maxFanSpeed        = 5000   // Max fan speed in RPM
		maxClockSpeed      = 3.20   // Max clock speed in GHz
		cpuUsageThreshold  = 80.0   // Max CPU usage in %
		memUsageThreshold  = 80.0   // Max memory usage in %
		diskUsageThreshold = 50.0   // Max disk usage in %
	)

	alertMessage := ""

	// Monitor CPU Temperature
	temps, err := sensors.CPUTemperature()
	if err != nil {
		log.Fatalf("Error fetching CPU temperature: %v\n", err)
	}
	for _, temp := range temps {
		if temp.Temperature < minTemp || temp.Temperature > maxTemp {
			alertMessage += fmt.Sprintf("Alert: CPU Temperature is out of safe range: %.2f°C\n", temp.Temperature)
		} else {
			fmt.Printf("CPU Temperature: %.2f°C (Safe)\n", temp.Temperature)
		}
	}

	// Monitor Fan Speeds
	fans, err := sensors.FanSpeeds()
	if err != nil {
		log.Fatalf("Error fetching fan speeds: %v\n", err)
	}
	for _, fan := range fans {
		if fan.Value < minFanSpeed || fan.Value > maxFanSpeed {
			alertMessage += fmt.Sprintf("Alert: Fan speed is out of safe range: %.2f RPM\n", fan.Value)
		} else {
			fmt.Printf("Fan Speed: %.2f RPM (Safe)\n", fan.Value)
		}
	}

	// Monitor CPU Clock Speed
	clockSpeeds, err := cpu.BaseFrequency()
	if err != nil {
		log.Fatalf("Error fetching CPU clock speed: %v\n", err)
	}
	for _, speed := range clockSpeeds {
		if float64(speed)/1e9 < maxClockSpeed {
			alertMessage += fmt.Sprintf("Alert: CPU Clock Speed is below 3.20 GHz: %.2f GHz\n", float64(speed)/1e9)
		} else {
			fmt.Printf("CPU Clock Speed: %.2f GHz (Safe)\n", float64(speed)/1e9)
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
