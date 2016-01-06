package network

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"
	"time"

	"github.com/alexgear/checker/config"
)

func Ping(ief string) bool {
	var target string
	if ief == "wifi" {
		target = "8.8.4.4:53"
	} else {
		target = "8.8.8.8:53"
	}
	d := net.Dialer{Timeout: 5 * time.Second}
	conn, err := d.Dial("tcp", target)
	if err != nil {
		log.Printf("Failed to initiate tcp connection: %s\n", err.Error())
		return false
	}
	//log.Println(conn.LocalAddr(), conn.RemoteAddr())
	defer conn.Close()
	return true
}

func InitNetwork() error {
	lines, err := exec.Command("nmcli", "dev", "wifi", "list", "ifname", "wlo1").Output()
	if err != nil {
		return fmt.Errorf("Failed to list connections: %s", err.Error())
	}
	var signalStrength string
	for i, line := range strings.Split(string(lines), "\n") {
		fields := strings.Fields(line)
		if i > 0 && len(fields) > 0 && fields[0] == "*" {
			signalStrength = fields[6]
			log.Printf("Signal strength = %s\n", signalStrength)
			break
		}
	}
	if signalStrength == "" {
		log.Println("Disconnected, trying to reconnect")
		exec.Command("nmcli", "connection", "delete", config.C.SSID).Output()
		_, err := exec.Command("nmcli",
			"dev", "wifi",
			"connect", config.C.SSID,
			"password", config.C.Password,
			"name", config.C.SSID,
			"ifname", config.C.WifiIef).Output()
		if err != nil {
			fmt.Errorf("Failed to reconnect: %s", err.Error())
		}
		exec.Command("sudo", "ip", "route", "add", "8.8.4.4", "via", config.C.WifiGw, "dev", config.C.WifiIef).Output()
	}
	exec.Command("sudo", "ip", "route", "add", "8.8.4.4", "via", config.C.WifiGw, "dev", config.C.WifiIef).Output()
	exec.Command("sudo", "ip", "route", "add", "8.8.8.8", "via", config.C.LanGw, "dev", config.C.LanIef).Output()
	return nil
}
