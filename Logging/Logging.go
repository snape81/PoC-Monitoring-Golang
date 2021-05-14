package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func main() {
	fmt.Println("Logging Command!")

	args := []string{"$PATH"}
	output, err := RunCMD("echo", args, true)

	if err != nil {
		fmt.Println("Error:", output)
	} else {
		fmt.Println("Result:", output)
	}

	args1 := []string{"/usr/bin/journalctl"}
	output1, err1 := RunCMD("sudo", args1, true)

	if err1 != nil {
		fmt.Println("Error:", output1)
	} else {
		fmt.Println("Result:", output1)
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
