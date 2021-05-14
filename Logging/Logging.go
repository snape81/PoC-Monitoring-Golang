package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func main() {
	fmt.Println("Logging Command!")

	args := []string{"journalctl"}
	output, err := RunCMD("sudo", args, true)

	if err != nil {
		fmt.Println("Error:", output)
	} else {
		fmt.Println("Result:", output)
	}

}

// RunCMD is a simple wrapper around terminal commands
func RunCMD(path string, args []string, debug bool) (out string, err error) {

	cmd := exec.Command(path, args...)

	var b []byte
	b, err = cmd.CombinedOutput()
	out = string(b)

	if debug {
		fmt.Println(strings.Join(cmd.Args[:], " "))

		if err != nil {
			fmt.Println("RunCMD ERROR")
			fmt.Println(out)
		}
	}

	return
}
