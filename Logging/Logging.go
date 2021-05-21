package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"os/exec"
	"strings"
)

func main() {
	c := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return (&net.Dialer{}).DialContext(ctx, "unix", "/run/snapd.socket")
			},
		},
	}

	req, _ := http.NewRequest("GET", "http://localhost/v2/apps", nil)

	res, err := c.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	b, _ := httputil.DumpResponse(res, true)
	fmt.Println(string(b))
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
