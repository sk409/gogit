package gogit

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/sk409/gofile"
)

type HTTPServer struct {
	PathPrefix        string
	rootDirectoryPath string
	gitBinPath        string
	git               *Git
}

func NewHTTPServer(rootDirectoryPath, gitBinPath string) *HTTPServer {
	return &HTTPServer{
		rootDirectoryPath: rootDirectoryPath,
		gitBinPath:        gitBinPath,
		git:               NewGit(rootDirectoryPath, gitBinPath),
	}
}

func (h *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.URL.Path, "info/refs") {
		h.handleInfoRefs(w, r)
	} else if strings.HasSuffix(r.URL.Path, rpcReceivePack) {
		//
		// readCloser, err := getReadCloser(r)
		// if err != nil {
		// 	return
		// }
		// requestBytes, err := ioutil.ReadAll(readCloser)
		// if err != nil {
		// 	return
		// }
		// fmt.Println("============================")
		// fmt.Println(string(requestBytes))
		// r.Body = ioutil.NopCloser(bytes.NewBuffer(requestBytes))
		//
		h.handleReceivePack(w, r)
	} else if strings.HasSuffix(r.URL.Path, rpcUploadPack) {
		h.handleUploadPack(w, r)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (h *HTTPServer) handleInfoRefs(w http.ResponseWriter, r *http.Request) {
	var service string
	if strings.Contains(r.URL.RawQuery, rpcUploadPack) {
		service = rpcUploadPack
	} else if strings.Contains(r.URL.RawQuery, rpcReceivePack) {
		service = rpcReceivePack
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	path := filepath.Join(
		strings.TrimPrefix(strings.Replace(r.URL.Path, "/info/refs", "", 1), h.PathPrefix),
	)
	repositoryDirectoryPath := filepath.Join(h.rootDirectoryPath, path)
	if !gofile.IsExist(repositoryDirectoryPath) {
		err := os.MkdirAll(repositoryDirectoryPath, 0755)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = h.git.InitBare(path)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	err := h.git.Refs(path, service, w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *HTTPServer) handleReceivePack(w http.ResponseWriter, r *http.Request) {
	h.handleRPC(w, r, rpcReceivePack)
}

func (h *HTTPServer) handleUploadPack(w http.ResponseWriter, r *http.Request) {
	h.handleRPC(w, r, rpcUploadPack)
}

func (h *HTTPServer) handleRPC(w http.ResponseWriter, r *http.Request, service string) {
	path := strings.TrimPrefix(strings.Replace(r.URL.Path, "/git-"+service, "", 1), h.PathPrefix)
	//err := h.git.RPCWithWriter(path, service, w, r)
	err := h.git.RPC(path, service, r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
