package business

import (
	"os/exec"
	"runtime"
	"log"
)

var commands = map[string]string{
	"windows": "start",
	"darwin":  "open",
	"linux":   "xdg-open",
}

var Version = "0.1.0"

// Open calls the OS default program for uri
func Open(uri string) error {
	run, ok := commands[runtime.GOOS]
	if !ok {
		log.Println(ok)
	}

	cmd := exec.Command(run, uri)
	return cmd.Start()
}