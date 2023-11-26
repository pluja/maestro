package utils

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"strings"
)

func GetLinuxDistro() (string, error) {
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
	if !disableFiles {
		context += fmt.Sprintf("\nCurrent working directory: %s", dir)
		files, err := GetDirFileList()
		if err != nil {
			return "", fmt.Errorf("error getting files: %w", err)
		}
		context += files
	}

	return context, nil
}
