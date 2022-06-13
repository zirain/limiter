package util

import (
	"strconv"

	"github.com/zirain/limiter/test/podexec"
)

func FortioLoad(c podexec.Client, podname string, n int, url string) error {
	cmds := []string{
		"fortio", "load",
		"-n", strconv.Itoa(n),
		url,
	}
	_, _, err := c.PodExecCommands(podname, "default", "fortio", cmds)
	return err
}
