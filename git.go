package gogit

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/sk409/goconst"
)

type Git struct {
	RootDirectoryPath string
	gitBinPath        string
}

func NewGit(rootDirectoryPath, gitBinPath string) *Git {
	return &Git{
		RootDirectoryPath: rootDirectoryPath,
		gitBinPath:        gitBinPath,
	}
}

func (g *Git) Clone(src, dst string) error {
	command := gitCommand(g.RootDirectoryPath, g.gitBinPath, "clone", src, dst)
	return command.Run()
}

func (g *Git) Init(path string) error {
	command := gitCommand(filepath.Join(g.RootDirectoryPath, path), g.gitBinPath, "init")
	return command.Run()
}

func (g *Git) InitBare(path string) error {
	command := gitCommand(filepath.Join(g.RootDirectoryPath, path), g.gitBinPath, "init", "--bare")
	return command.Run()
}

func (g *Git) LsFiles(path string) ([]byte, error) {
	command := gitCommand(filepath.Join(g.RootDirectoryPath, path), g.gitBinPath, "ls-files")
	return command.Output()
}

func (g *Git) UploadPack(path string, args ...string) ([]byte, error) {
	args = append([]string{rpcUploadPack}, args...)
	command := gitCommand(filepath.Join(g.RootDirectoryPath, path), g.gitBinPath, args...)
	return command.Output()
}

func (g *Git) ReceivePack(path string, args ...string) ([]byte, error) {
	args = append([]string{rpcReceivePack}, args...)
	command := gitCommand(filepath.Join(g.RootDirectoryPath, path), g.gitBinPath, args...)
	return command.Output()
}

func (g *Git) Refs(path, service string, w http.ResponseWriter) error {
	var refs []byte
	var err error
	args := []string{statelessRPC, advertiseRefs, "."}
	if service == rpcUploadPack {
		refs, err = g.UploadPack(path, args...)
	} else {
		refs, err = g.ReceivePack(path, args...)
	}
	if err != nil {
		return err
	}
	w.Header().Set(goconst.HTTP_HEADER_CONTENT_TYPE, fmt.Sprintf("application/x-git-%s-advertisement", service))
	w.WriteHeader(http.StatusOK)
	head := "# service=git-" + service + "\n"
	size := fmt.Sprintf("%04s", strconv.FormatInt(int64(len(head)+4), 16))
	w.Write([]byte(size + head))
	w.Write([]byte("0000"))
	w.Write(refs)
	return nil
}

func (g *Git) RPC(path, service string, w http.ResponseWriter, r *http.Request) error {
	var body io.ReadCloser
	var err error
	if r.Header.Get(goconst.HTTP_HEADER_CONTENT_ENCODING) == goconst.HTTP_HEADER_CONTENT_ENCODING_GZIP {
		body, err = gzip.NewReader(r.Body)
		if err != nil {
			return err
		}
		defer body.Close()
	} else {
		body = r.Body
	}
	repositoryDirectoryPath := filepath.Join(
		g.RootDirectoryPath,
		path,
	)
	args := []string{service, statelessRPC, "."}
	command := gitCommand(repositoryDirectoryPath, g.gitBinPath, args...)
	defer cleanUpProcessGroup(command)
	stdin, err := command.StdinPipe()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}
	defer stdin.Close()
	stdout, err := command.StdoutPipe()
	if err != nil {
		return err
	}
	defer stdout.Close()
	err = command.Start()
	if err != nil {
		return err
	}
	bufferIn := bufferPool.Get().([]byte)
	defer bufferPool.Put(bufferIn)
	if _, err := io.CopyBuffer(stdin, body, bufferIn); err != nil {
		return err
	}
	stdin.Close()
	w.Header().Set(goconst.HTTP_HEADER_CONTENT_TYPE, fmt.Sprintf("application/x-git-%s-result", service))
	bufferOut := bufferPool.Get().([]byte)
	defer bufferPool.Put(bufferOut)
	if _, err := io.CopyBuffer(w, stdout, bufferOut); err != nil {
		return err
	}
	if err = command.Wait(); err != nil {
		return err
	}
	return nil
}
