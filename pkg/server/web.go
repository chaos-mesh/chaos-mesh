package server

import (
	"net/http"
	"os"
	"path"
)

func (s *Server) WebServer(fs http.FileSystem) http.Handler {
	fsh := http.FileServer(fs)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if f, err := fs.Open(path.Clean(r.URL.Path)); err == nil {
			f.Close()
		} else if os.IsNotExist(err) {
			r.URL.Path = "/"
		}
		fsh.ServeHTTP(w, r)
	})
}

func (s *Server) web(prefix, root string) http.Handler {
	return http.StripPrefix(prefix, s.WebServer(http.Dir(root)))
}
