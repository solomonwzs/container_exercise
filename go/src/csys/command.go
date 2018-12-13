package csys

import (
	"os"
	"os/exec"

	"github.com/solomonwzs/goxutil/logger"
)

func SystemCmd(cmd string, arg ...string) (err error) {
	logger.Debug(cmd, arg)
	c := exec.Command(cmd, arg...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err = c.Run(); err != nil {
		return
	}
	c.Wait()
	return
}
