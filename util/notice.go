package util

import "os/exec"

func NotifyMac(title, message string) error {
	cmd := exec.Command("osascript", "-e", `display notification "`+message+`" with title "`+title+`"`)
	return cmd.Run()
}
