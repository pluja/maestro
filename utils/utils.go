package utils

import (
	"bufio"
	"bytes"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
	"fmt"
	"net/url"
)

// GetLinuxDistro retrieves the OS version for macOS or Linux.
func GetLinuxDistro() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		// Execute the sw_vers command for macOS
		cmd := exec.Command("sw_vers", "-productVersion")
		var out bytes.Buffer
		cmd.Stdout = &out
		err := cmd.Run()
		if err != nil {
			return "", err
		}

		// Read the output
		scanner := bufio.NewScanner(&out)
		for scanner.Scan() {
			line := scanner.Text()
			return strings.TrimSpace(line), nil
		}

		if err := scanner.Err(); err != nil {
			return "", err
		}

	case "linux":
		// Open the /etc/os-release file for Linux
		file, err := os.Open("/etc/os-release")
		if err != nil {
			return "", err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "PRETTY_NAME=") {
				return strings.Trim(line[len("PRETTY_NAME="):], "\""), nil
			}
		}

		if err := scanner.Err(); err != nil {
			return "", err
		}
	}

	return "Unknown", nil
}

func GetDirFileList() (string, error) {
	// Get the current working directory
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error getting current directory: %w", err)
	}

	// Read the contents of the directory
	files, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("error reading directory: %w", err)
	}

	// StringBuilder to accumulate file details
	var builder strings.Builder

	// Convert the file list to a string
	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			return "", fmt.Errorf("error getting info for file %s: %w", file.Name(), err)
		}
		builder.WriteString(fmt.Sprintf("Filename: %v, Size: %v bytes, IsDir: %t, Mode: %v\n", file.Name(), info.Size(), file.IsDir(), info.Mode()))
	}

	return builder.String(), nil
}

func GetContext(disableFiles bool) (string, error) {
	// Get the current working directory
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error getting current directory: %w", err)
	}

	context := "Linux distro: "
	distro, err := GetLinuxDistro()
	if err != nil {
		return "", fmt.Errorf("error getting linux distro: %w", err)
	}
	context += distro

	// User Username, UID, GID
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}

	context += fmt.Sprintf("\nUsername: %s, UID: %s, GID: %s", currentUser.Username, currentUser.Uid, currentUser.Gid)
	if disableFiles {
		context += fmt.Sprintf("\nCurrent working directory: %s", dir)
		files, err := GetDirFileList()
		if err != nil {
			return "", fmt.Errorf("error getting files: %w", err)
		}
		context += files
	}

	return context, nil
}

// compareVersion compares two version strings, returning true if v1 is greater than v2
func CompareVersion(v1, v2 string) bool {
	v1Parts := strings.Split(v1, ".")
	v2Parts := strings.Split(v2, ".")
	for i := 0; i < len(v1Parts) && i < len(v2Parts); i++ {
		if v1Parts[i] > v2Parts[i] {
			return true
		} else if v1Parts[i] < v2Parts[i] {
			return false
		}
	}
	return len(v1Parts) >= len(v2Parts)
}

func SanitizeEndpoint(endpoint string) string {
	if !strings.HasPrefix(endpoint, "http") && !strings.HasPrefix(endpoint, "https") {
		endpoint = fmt.Sprintf("http://%s", endpoint)
	}
	endpoint = strings.ReplaceAll(endpoint, "/api/chat", "")
	endpoint = strings.TrimSuffix(endpoint, "/")

	// Get only the host and scheme
	url, err := url.Parse(endpoint)
	if err != nil {
		return endpoint
	}
	endpoint = fmt.Sprintf("%s://%s", url.Scheme, url.Host)

	return endpoint
}
