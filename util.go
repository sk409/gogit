package gogit

import (
	"os/exec"
	"syscall"
)

func gitCommand(repositoryPath string, gitBinPath string, args ...string) *exec.Cmd {
	command := exec.Command(gitBinPath, args...)
	command.Dir = repositoryPath
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	return command
}

func cleanUpProcessGroup(cmd *exec.Cmd) {
	if cmd == nil {
		return
	}
	if process := cmd.Process; process != nil && process.Pid > 0 {
		syscall.Kill(-process.Pid, syscall.SIGTERM)
	}
	cmd.Wait()
}
