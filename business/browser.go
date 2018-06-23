package business

import (
	"os/exec"
	"runtime"
)

// Open calls the OS default program for uri
func Open(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd.exe", "/c", "start", url)
		break
	case "darwin":
		cmd = exec.Command("open", url)
		break
	case "linux":
		cmd = exec.Command("xdg-open", url)
		break
	default:
		break

	}
	return cmd.Start()
}