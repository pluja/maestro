package utils

import (
	"fmt"
	"os"
	"os/exec"
)

func RunCommand(command string) (string, error) {
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%sError executing command: %s%s\n", ColorRed, err, ColorReset)
	}

	return "", nil
}
