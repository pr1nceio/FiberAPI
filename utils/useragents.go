package utils

import (
	"fmt"
	"strings"
)

func ParseUserAgent(userAgent string) string {
	if !strings.Contains(userAgent, ")") {
		return userAgent
	}
	hw := strings.Split(userAgent, ")")
	hw = strings.Split(hw[0], "(")
	if len(hw) < 2 {
		return userAgent
	}
	hw = strings.Split(hw[1], ";")
	if len(hw) < 2 {
		return getOSByUAHW(hw)
	}
	return getOSByUAHW(hw)
}

func getOSByUAHW(data []string) string {
	var hw, ver string
	hw = data[0]
	if len(data) >= 2 {
		ver = data[1]
	}
	hw = strings.ToLower(hw)
	hw = strings.TrimSpace(hw)
	ver = strings.ToLower(ver)
	ver = strings.TrimSpace(ver)
	switch hw {
	case "macintosh": //MacOS
		{
			ver1 := strings.Split(ver, "os x")
			if len(ver1) > 1 {
				ver = strings.TrimSpace(ver1[1])
				return fmt.Sprintf("macOS %s", strings.ReplaceAll(ver, "_", "."))
			}
			return "macOS"
		}
	case "iphone":
		{
			ver1 := strings.Split(ver, "os ")
			if len(ver1) > 1 {
				ver1 = strings.Split(ver1[1], " ")
				if len(ver1) > 1 {
					ver = strings.TrimSpace(ver1[0])
					return fmt.Sprintf("iOS %s", strings.ReplaceAll(ver, "_", "."))
				}
			}
			return "iOS"
		}
	case "ipad":
		{
			ver1 := strings.Split(ver, "os ")
			if len(ver1) > 1 {
				ver1 = strings.Split(ver1[1], " ")
				if len(ver1) > 1 {
					ver = strings.TrimSpace(ver1[0])
					return fmt.Sprintf("iPadOS %s", strings.ReplaceAll(ver, "_", "."))
				}
			}
			return "iPadOS"
		}
	case "wayland": // Newer linux
		{
			return "Linux (Wayland)"
		}
	case "x11": // Linux, ChromeOS and Unix-like
		{
			if strings.Contains(ver, "cros ") {
				return "ChromeOS"
			}
			if strings.Contains(ver, "android ") {
				return "Android (Non-standard)"
			}
			return "Linux"
		}
	case "linux": // Mostly android (Chromium)
		{
			if strings.Contains(ver, "android ") {
				ver = strings.ReplaceAll(ver, "android ", "")
				return fmt.Sprintf("Android %s", strings.TrimSpace(ver))
			}
			if len(data) >= 3 && strings.Contains(strings.ToLower(data[2]), "android ") {
				ver = strings.ReplaceAll(strings.ToLower(data[2]), "android ", "")
				return fmt.Sprintf("Android %s", strings.TrimSpace(ver))
			}
			return "Linux (Non-standard)"
		}
	default:
		{
			// Windows
			if strings.HasPrefix(hw, "windows") {
				ver1 := strings.Split(hw, "windows")
				if len(ver1) > 1 {
					ver = strings.TrimSpace(strings.ReplaceAll(ver1[1], "nt", ""))
					ver2 := ""
					switch ver {
					case "10":
						fallthrough
					case "10.0":
						ver2 = "10"
					case "6.3":
						ver2 = "8.1"
					case "6.2":
						ver2 = "8"
					case "6.1":
						ver2 = "7"
					case "6.0":
						ver2 = "Vista"
					case "5.1":
						ver2 = "XP"
					default:
						ver2 = ""
					}
					return fmt.Sprintf("Windows %s", ver2)
				}
				return "Windows"
			}

			// Specific Android
			if strings.HasPrefix(hw, "android") {
				ver1 := strings.Split(hw, "android")
				if len(ver1) > 1 {
					ver = strings.TrimSpace(ver1[1])
					return fmt.Sprintf("Android %s", ver)
				}
				return "Android"
			}

			// Other
			return fmt.Sprintf("Unknown (%s; %s)", hw, ver)
		}
	}
}
