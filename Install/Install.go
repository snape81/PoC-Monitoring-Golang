package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

const (
	socketFile     = "/run/snapd.socket"
	urlAssertions  = "/v2/assertions"
	urlSnaps       = "/v2/snaps"
	typeAssertions = "application/x.ubuntu.assertion"
	baseURL        = "http://localhost"
	download_path  = "/home/fbianca/snap-folder"
)

func main() {
	const url_snap = "https://raw.githubusercontent.com/snape81/PoC-Monitoring-Golang/main/hello-lhc_4.snap"
	const url_assert = "https://raw.githubusercontent.com/snape81/PoC-Monitoring-Golang/main/hello-lhc_4.assert"

	err_lhc := DownloadFile(url_snap, download_path+"/hello-lhc_4.snap")
	if err_lhc != nil {
		panic(err_lhc)
	}

	err := DownloadFile(url_assert, download_path+"/hello-lhc_4.assert")
	if err != nil {
		panic(err)
	}

	snapclient := NewClient(download_path)
	err_sideload := snapclient.SideloadInstall("hello-lhc", "4")
	if err_sideload != nil {
		panic(err_sideload)
	}

	/*args := []string{"install", "/home/fbianca/snap-folder/hello-world.snap", "--dangerous"}
	output, err_lhc := RunCMD("snap", args, true)

	if err_lhc != nil {
		fmt.Println("Error:", output)
	} else {
		fmt.Println("Result:", output)
	}*/

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

// Snapd service to access the snapd REST API
type Snapd struct {
	downloadPath string
	client       *http.Client
}

// NewClient returns a snapd API client
func NewClient(downloadPath string) *Snapd {
	return &Snapd{
		downloadPath: downloadPath,
		client: &http.Client{
			Transport: &http.Transport{
				DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
					return net.Dial("unix", socketFile)
				},
			},
		},
	}
}

func (snap *Snapd) call(method, url, contentType string, body io.Reader) (*http.Response, error) {
	u := baseURL + url

	switch method {
	case "POST":
		return snap.client.Post(u, contentType, body)
	case "GET":
		return snap.client.Get(u)
	default:
		return nil, fmt.Errorf("unsupported method: %s", method)
	}
}

// Ack acknowledges a (snap) assertion
func (snap *Snapd) Ack(assertion []byte) error {
	_, err := snap.call("POST", urlAssertions, typeAssertions, bytes.NewReader(assertion))
	return err
}

// InstallPath installs a snap from a local file
func (snap *Snapd) InstallPath(name, filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("cannot open: %q", filePath)
	}

	pr, pw := io.Pipe()
	mw := multipart.NewWriter(pw)
	go sendSnapFile(name, filePath, f, pw, mw)

	_, err = snap.call("POST", urlSnaps, mw.FormDataContentType(), pr)
	return err
}

// List the installed snaps
func (snap *Snapd) List() ([]byte, error) {
	resp, err := snap.call("GET", urlSnaps, "application/json; charset=UTF-8", nil)
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(resp.Body)
}

// SideloadInstall side loads a snap by acknowledging the assertion and installing the snap
func (snap *Snapd) SideloadInstall(name, revision string) error {
	//assertsPath := path.Join(snap.downloadPath, fmt.Sprintf("%s_%s.assert", name, revision))
	snapPath := path.Join(snap.downloadPath, fmt.Sprintf("%s_%s.snap", name, revision))

	// acknowledge the snap assertion
	/*dataAssert, err := ioutil.ReadFile(assertsPath)
	if err != nil {
		return err
	}
	if err := snap.Ack(dataAssert); err != nil {
		return err
	}*/

	// install the snap
	return snap.InstallPath(name, snapPath)
}

func sendSnapFile(name, snapPath string, snapFile *os.File, pw *io.PipeWriter, mw *multipart.Writer) {
	defer snapFile.Close()

	fields := []struct {
		name  string
		value string
	}{
		{"action", "install"},
		{"name", name},
		{"snap-path", snapPath},
	}
	for _, s := range fields {
		if s.value == "" {
			continue
		}
		if err := mw.WriteField(s.name, s.value); err != nil {
			pw.CloseWithError(err)
			return
		}
	}

	fw, err := mw.CreateFormFile("snap", filepath.Base(snapPath))
	if err != nil {
		pw.CloseWithError(err)
		return
	}

	_, err = io.Copy(fw, snapFile)
	if err != nil {
		pw.CloseWithError(err)
		return
	}

	mw.Close()
	pw.Close()
}
