package gogit

import "sync"

var (
	bufferPool = sync.Pool{New: func() interface{} { return make([]byte, 32*1024) }}
)

const (
	RPCUploadPack  = "upload-pack"
	RPCReceivePack = "receive-pack"
	statelessRPC   = "--stateless-rpc"
	advertiseRefs  = "--advertise-refs"
)
