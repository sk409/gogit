package gogit

import (
	"compress/gzip"
	"io"
	"net/http"
	"os/exec"
	"syscall"

	"github.com/sk409/goconst"
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

func GetReadCloser(r *http.Request) (io.ReadCloser, error) {
	var body io.ReadCloser
	var err error
	if r.Header.Get(goconst.HTTP_HEADER_CONTENT_ENCODING) == goconst.HTTP_HEADER_CONTENT_ENCODING_GZIP {
		body, err = gzip.NewReader(r.Body)
		if err != nil {
			return nil, err
		}
		defer body.Close()
	} else {
		body = r.Body
	}
	return body, nil
}
