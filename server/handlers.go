package server

import (
	"encoding/json"
	"fast/fileservice"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strings"
)

func RegisterHandlers() {
	http.Handle("/", &srvHanler{})
}

type srvHanler struct{}

func NewHandler() *srvHanler {
	return &srvHanler{}
}

func (srv srvHanler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	files := make([]fileservice.FileEntry, 0)
	p := strings.TrimPrefix(r.URL.Path, "/")

	switch r.Method {
	case http.MethodGet:
		keyword := r.URL.Query().Get("keyword")
		if keyword != "" {
			log.Printf("[search] - Keyword: [%s] <-- Addr: [%s]\n", keyword, r.RemoteAddr)
			files = fileservice.FS.Search(keyword)
		} else {
			if p == "" {
				files = fileservice.FS.Dir(p)
				log.Printf("[browse] - Dir: [/] - Addr: [%s]\n", r.RemoteAddr)
			} else if file, ok := fileservice.FS.File(p); ok {
				if file.IsDir {
					log.Printf("[browse] - Dir: [/%s] - Addr: [%s]\n", p, r.RemoteAddr)
					files = fileservice.FS.Dir(p)
				} else {
					log.Printf("[download] - File: [/%s] --> Addr: [%s] Status: [start]\n", p, r.RemoteAddr)

					w.Header().Set("Content-Disposition", "attachment; filename="+file.Name)
					w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
					http.ServeFile(w, r, fileservice.FS.RealPath(file))
					log.Printf("[download] - File: [/%s] --> Addr: [%s] Status: [done]\n", p, r.RemoteAddr)
					return
				}
			}
		}
	case http.MethodPost:
		if !fileservice.FS.Upload {
			http.Error(w, "Upload is disabled", http.StatusForbidden)
			return
		}
		files = fileservice.FS.Dir(p)
		saveFiles(r, w, p)
		http.Redirect(w, r, p, http.StatusFound)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	rs := make([]fileservice.FileEntry, 0)
	for _, f := range files {
		rs = append(rs, fileservice.FileEntry{
			Name:    f.Name,
			Path:    path.Join(f.Path, f.Name),
			Size:    f.Size,
			Type:    f.Type,
			IsDir:   f.IsDir,
			ModTime: f.ModTime,
		})
	}
	filesJson, err := json.Marshal(rs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data := &ViewData{
		Files:     rs,
		Upload:    fileservice.FS.Upload,
		JsonFiles: string(filesJson),
		Style:     getCSS(),
		Script:    getScript(),
	}
	err = rootTemplate.Lookup("index.gohtml").Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func saveFiles(r *http.Request, w http.ResponseWriter, p string) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	mulfiles := r.MultipartForm.File["files"]

	for _, fileHeader := range mulfiles {
		file, err := fileHeader.Open()
		if err != nil {
			log.Printf("Error opening file: %v\n", err)
			continue
		}
		defer file.Close()
		saveFile(p, fileHeader, w, r, file)
	}
}

func saveFile(p string, fileHeader *multipart.FileHeader, w http.ResponseWriter, r *http.Request, file multipart.File) {
	fullPath := path.Join(p, fileHeader.Filename)
	_, ok := fileservice.FS.File(fullPath)
	if ok {
		ext := path.Ext(fullPath)
		fn := path.Base(fullPath)
		fn = strings.TrimSuffix(fn, ext)
		fullPath = fmt.Sprintf("%s (copy)%s", fn, ext)
	}

	realpath := path.Join(fileservice.FS.RootDir, fullPath)
	dst, err := os.Create(realpath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	info, err := os.Stat(realpath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fileservice.FS.Insert(&fileservice.FileEntry{
		Name:    fileHeader.Filename,
		Path:    p,
		IsDir:   false,
		Size:    fileservice.FileSize(info.Size()),
		Type:    strings.TrimPrefix(path.Ext(fileHeader.Filename), "."),
		ModTime: fileservice.Timestamp(info.ModTime().Unix()),
	})
	log.Printf("[upload] - File: [/%s] <-- Addr: [%s]\n", fullPath, r.RemoteAddr)
}
