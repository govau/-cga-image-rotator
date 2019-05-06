package main

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type imageHandler struct {
	Path      string
	Extension string
	MIME      string
	TTL       int64 // seconds

	data [][]byte
}

func (ih *imageHandler) CreateHandler() (http.HandlerFunc, error) {
	files, err := ioutil.ReadDir(ih.Path)
	if err != nil {
		return nil, err
	}
	ih.data = nil
	for _, f := range files { // docs guarantee to be sorted alphabetically
		if strings.HasSuffix(f.Name(), "."+ih.Extension) {
			bs, err := ioutil.ReadFile(filepath.Join(ih.Path, f.Name()))
			if err != nil {
				return nil, err
			}
			buf := &bytes.Buffer{}
			gzw, err := gzip.NewWriterLevel(buf, gzip.BestCompression)
			if err != nil {
				return nil, err
			}
			_, err = gzw.Write(bs)
			if err != nil {
				return nil, err
			}
			err = gzw.Close()
			if err != nil {
				return nil, err
			}
			ih.data = append(ih.data, buf.Bytes())
		}
	}
	if len(ih.data) == 0 {
		return nil, errors.New("no files found")
	}
	return ih.handleIt, nil
}

func (ih *imageHandler) handleIt(w http.ResponseWriter, r *http.Request) {
	now := time.Now().Unix()
	idxToServe := (now / ih.TTL) % int64(len(ih.data))
	maxAge := (((now / ih.TTL) + 1) * ih.TTL) - now

	w.Header().Set("Content-Type", ih.MIME)
	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Set("Content-Length", strconv.Itoa(len(ih.data[idxToServe])))
	w.Header().Set("Cache-Control", "public, max-age="+strconv.Itoa(int(maxAge)))
	w.Write(ih.data[idxToServe])
}

func main() {
	ext := os.Getenv("EXTENSION")
	if ext == "" {
		log.Fatal("EXTENSION must be specified")
	}
	mime := os.Getenv("MIMETYPE")
	if mime == "" {
		log.Fatal("MIMETYPE must be specified")
	}
	dur, err := strconv.Atoi(os.Getenv("TTL"))
	if dur <= 0 || err != nil {
		log.Fatal("TTL must be specified, and a positive integer in seconds")
	}
	names := strings.Split(os.Getenv("ALLOWED_NAMES"), ",")
	for _, name := range names {
		if name == "" {
			log.Fatal("ALLOWED_NAMES must be specified and be a comma separated list of strings")
		}
		h, err := (&imageHandler{
			Path:      name,
			Extension: ext,
			MIME:      mime,
			TTL:       int64(dur),
		}).CreateHandler()
		if err != nil {
			log.Fatal(err)
		}
		http.HandleFunc(fmt.Sprintf("/%s.%s", name, ext), h)
	}
	log.Println("Serving...")
	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), nil))
}
