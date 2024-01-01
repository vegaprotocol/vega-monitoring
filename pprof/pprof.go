package pprof

import (
	"net/http"
	_ "net/http/pprof"
)

func StartPprofServer(listenInterface string) error {
	return http.ListenAndServe(listenInterface, nil)
}
