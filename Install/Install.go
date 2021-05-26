package main

import (
	"bytes"
	"context"
	"encoding/json"
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
	urlLogin       = "/v2/login"
	typeJSON       = "application/json"
	typeAssertions = "application/x.ubuntu.assertion"
	baseURL        = "http://localhost"
	download_path  = "/home/claudio/snap-folder"
)

type Result struct {
	Id       string `json:"id"`
	Email    string `json:"email,omitempty"`
	Macaroon string `json:"macaroon,omitempty"`
}

type Wrapper struct {
	Type             string `json:"type"`
	Statuscode       string `json:"status-code,omitempty"`
	Status           string `json:"status,omitempty"`
	Result           Result `json:"result,omitempty"`
	WarningTimestamp string `json:"warning-timestamp,omitempty"`
	WarningCount     int    `json:"warning-count,omitempty"`
}

func main() {
	const url_snap = "https://raw.githubusercontent.com/snape81/PoC-Monitoring-Golang/main/hello-lhc_4.snap"
	const url_assert = "https://raw.githubusercontent.com/snape81/PoC-Monitoring-Golang/main/hello-lhc_4.assert"
	snapclient := NewClient(download_path)

	macaroon, err_login := snapclient.Login("claudio.starnoni@gmail.com", "Ubuntu001*")
	if err_login != nil {
		panic(err_login)
	}
	fmt.Println(macaroon)

	err_lhc := DownloadFile(url_snap, download_path+"/hello-lhc_4.snap")
	if err_lhc != nil {
		panic(err_lhc)
	}

	err := DownloadFile(url_assert, download_path+"/hello-lhc_4.assert")
	if err != nil {
		panic(err)
	}

	err_sideload := snapclient.SideloadInstall("hello-lhc", "4", macaroon)
	if err_sideload != nil {
		panic(err_sideload)
	}

	list, err_lhc := snapclient.List()
	println(string(list))

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

func (snap *Snapd) call(method, url, contentType string, body io.Reader, macaroon string) (*http.Response, error) {
	u := baseURL + url

	switch method {
	case "POST":
		req, err := http.NewRequest("POST", url, body)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", contentType)
		req.Header.Set("Authorization", "Macaroon root=\""+macaroon+"\",discharge=\"discharge-for-macaroon-authentication\"")

		return snap.client.Do(req)
	case "GET":
		return snap.client.Get(u)
	default:
		return nil, fmt.Errorf("unsupported method: %s", method)
	}
}

func (snap *Snapd) loginCall(method, url, contentType string, body io.Reader) (*http.Response, error) {
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
func (snap *Snapd) Ack(assertion []byte, macaroon string) error {
	call, err := snap.call("POST", urlAssertions, typeAssertions, bytes.NewReader(assertion), macaroon)
	fmt.Println("ACK")
	fmt.Println(call)
	return err
}

// Ack acknowledges a (snap) assertion
func (snap *Snapd) Login(email, password string) (string, error) {
	params := map[string]string{
		"email": email, "password": password,
	}
	data, err := json.Marshal(&params)
	if err != nil {
		return "nil", err
	}

	resp, err := snap.loginCall("POST", urlLogin, typeJSON, bytes.NewReader(data))

	defer resp.Body.Close()
	mywrapper := Wrapper{}
	err = json.NewDecoder(resp.Body).Decode(&mywrapper)
	fmt.Println(mywrapper)
	return mywrapper.Result.Macaroon, err
}

// InstallPath installs a snap from a local file
func (snap *Snapd) InstallPath(name, filePath, macaroon string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("cannot open: %q", filePath)
	}

	pr, pw := io.Pipe()
	mw := multipart.NewWriter(pw)
	go sendSnapFile(name, filePath, f, pw, mw)

	call, err := snap.call("POST", urlSnaps, mw.FormDataContentType(), pr, macaroon)
	fmt.Println("INSTALL ")
	fmt.Println(call)
	return err
}

// List the installed snaps
func (snap *Snapd) List() ([]byte, error) {
	resp, err := snap.loginCall("GET", urlSnaps, "application/json; charset=UTF-8", nil)
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(resp.Body)
}

// SideloadInstall side loads a snap by acknowledging the assertion and installing the snap
func (snap *Snapd) SideloadInstall(name, revision, macaroon string) error {
	assertsPath := path.Join(snap.downloadPath, fmt.Sprintf("%s_%s.assert", name, revision))
	snapPath := path.Join(snap.downloadPath, fmt.Sprintf("%s_%s.snap", name, revision))

	// acknowledge the snap assertion
	dataAssert, err := ioutil.ReadFile(assertsPath)
	if err != nil {
		return err
	}
	if err := snap.Ack(dataAssert, macaroon); err != nil {
		return err
	}

	// install the snap
	return snap.InstallPath(name, snapPath, macaroon)
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

func getJson(r http.Response, target interface{}) error {

	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}
