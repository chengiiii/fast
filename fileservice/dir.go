package fileservice

import (
	"os"
	stpath "path"
	"strings"
	"sync"

	"github.com/google/btree"
)

type FileEntry struct {
	Name    string
	Path    string
	Size    FileSize
	Type    string
	ModTime Timestamp
	IsDir   bool
}

type FileList []FileEntry

type Timestamp int64

type FileSize int64

func (f *FileEntry) Less(item btree.Item) bool {
	if f.Path == item.(*FileEntry).Path {
		return f.Name < item.(*FileEntry).Name
	}
	return f.Path < item.(*FileEntry).Path
}

type Fs struct {
	RootDir string
	root    *btree.BTree
	mutex   *sync.Mutex
	depth   int
	Upload  bool
}

func (fs *Fs) createTree(rootPath string, depth int) {
	if rootPath == "" {
		panic("rootDir is empty")
	}
	if fs.depth != 0 && depth > fs.depth {
		return
	}
	rootPath = stpath.Clean(rootPath)
	dirEntrys, err := os.ReadDir(rootPath)
	if err != nil {
		panic(err)
	}

	if rootPath == "." {
		rootPath = ""
	}

	for _, dirEntry := range dirEntrys {
		fileinfo, err := dirEntry.Info()
		if err != nil {
			continue
		}

		filetype := strings.TrimPrefix(stpath.Ext(dirEntry.Name()), ".")
		if filetype == "" {
			if fileinfo.IsDir() {
				filetype = "dir"
			} else {
				filetype = "file"
			}
		}
		filetype = strings.ToUpper(filetype)
		fs.Insert(&FileEntry{
			Name:    dirEntry.Name(),
			Path:    strings.TrimPrefix(strings.TrimPrefix(rootPath, stpath.Clean(fs.RootDir)), "/"),
			Size:    FileSize(fileinfo.Size()),
			Type:    filetype,
			IsDir:   dirEntry.IsDir(),
			ModTime: Timestamp(fileinfo.ModTime().Unix()),
		})
		if dirEntry.IsDir() {
			fs.createTree(stpath.Join(rootPath, dirEntry.Name()), depth+1)
		}
	}
}

func (fs *Fs) Init(dir string, upload bool, depth int) {
	fs.RootDir = dir
	fs.Upload = upload
	fs.depth = depth
	fs.createTree(dir, 1)
}

func (fs *Fs) Insert(file *FileEntry) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()
	fs.root.ReplaceOrInsert(file)
}

func (fs *Fs) Dir(path string) FileList {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()
	result := make([]FileEntry, 0)
	fs.root.AscendRange(&FileEntry{Path: path}, &FileEntry{Path: path + "/"}, func(i btree.Item) bool {
		result = append(result, *i.(*FileEntry))
		return true
	})
	return result
}

func (fs *Fs) File(fp string) (file *FileEntry, ok bool) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()
	file = &FileEntry{}
	fs.root.AscendGreaterOrEqual(file, func(i btree.Item) bool {
		f := i.(*FileEntry)
		if stpath.Join(f.Path, f.Name) == fp {
			file = i.(*FileEntry)
			ok = true
			return false
		}
		return true
	})
	return
}

func (fs *Fs) AllFiles() FileList {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()
	result := make([]FileEntry, 0)
	fs.root.Ascend(func(i btree.Item) bool {
		result = append(result, *i.(*FileEntry))
		return true
	})
	return result
}

func (fs *Fs) Search(keyword string) FileList {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()
	keyword = stpath.Clean(keyword)
	fn := stpath.Base(keyword)
	dir := stpath.Dir(keyword)
	result := make([]FileEntry, 0)
	if dir == "." {
		fs.root.Ascend(func(item btree.Item) bool {
			if strings.Contains(item.(*FileEntry).Name, keyword) {
				result = append(result, *item.(*FileEntry))
			}
			return true
		})
	} else {
		fs.root.AscendRange(&FileEntry{Path: dir}, &FileEntry{Path: dir + "/"}, func(item btree.Item) bool {
			if strings.Contains(item.(*FileEntry).Name, fn) {
				result = append(result, *item.(*FileEntry))
			}
			return true
		})
	}
	return result
}

func (fs *Fs) RealPath(file *FileEntry) string {
	return stpath.Join(fs.RootDir, file.Path, file.Name)
}

var FS = &Fs{
	root:  btree.New(2),
	mutex: &sync.Mutex{},
}
