package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

func main() {
	const url = "https://raw.githubusercontent.com/francescobianca/PoC-Monitoring-Golang/main/hello-world_29.snap"

	err := DownloadFile(url, "/home/fbianca/snap-folder/hello-world.snap")
	if err != nil {
		panic(err)
	}

	// Calling Sleep method
	time.Sleep(10 * time.Second)

	/*install, err := exec.Command("snap install /home/fbianca/snap-folder/hello-world.snap", "--dangerous").Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", install)*/

	args := []string{"install", "/home/fbianca/snap-folder/hello-world.snap", "--dangerous"}
	output, err := RunCMD("snap", args, true)

	if err != nil {
		fmt.Println("Error:", output)
	} else {
		fmt.Println("Result:", output)
	}

}

// DownloadFile will download a url and store it in local filepath.
// It writes to the destination file as it downloads it, without
// loading the entire file into memory.
func DownloadFile(url string, filepath string) error {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
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
