package main

import (
	"io"
	"net/http"
	"os"
)

func main() {
	const url = "https://raw.githubusercontent.com/francescobianca/PoC-Monitoring-Golang/main/hello-world_29.snap"

	err := DownloadFile(url, "/home/fbianca/snap-folder/hello-world.snap")
	if err != nil {
		panic(err)
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
