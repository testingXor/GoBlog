package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const staticFolder = "static"

func allStaticPaths() (paths []string) {
	paths = []string{}
	err := filepath.Walk(staticFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Mode().IsRegular() {
			paths = append(paths, strings.TrimPrefix(path, staticFolder))
		}
		return nil
	})
	if err != nil {
		return
	}
	return
}

// Gets only called by registered paths
func serveStaticFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", fmt.Sprintf("public,max-age=%d,s-max-age=%d,stale-while-revalidate=%d", appConfig.Cache.Expiration, appConfig.Cache.Expiration/3, appConfig.Cache.Expiration))
	http.ServeFile(w, r, filepath.Join(staticFolder, r.URL.Path))
}
