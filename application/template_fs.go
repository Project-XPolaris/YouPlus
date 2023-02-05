package application

import "os"

type FileInfoTemplate struct {
	FileName    string `json:"name"`
	FileSize    int64  `json:"size"`
	FileMode    uint32 `json:"mode"`
	FileModTime string `json:"modTime"`
	FileIsDir   bool   `json:"isDir"`
}

func NewFileInfoTemplate(file os.FileInfo) *FileInfoTemplate {
	return &FileInfoTemplate{
		FileName:    file.Name(),
		FileSize:    file.Size(),
		FileMode:    uint32(file.Mode()),
		FileModTime: file.ModTime().Format("2006-01-02 15:04:05"),
		FileIsDir:   file.IsDir(),
	}
}

type FileTemplate struct {
	Info FileInfoTemplate `json:"info"`
}

func NewFileTemplate(info os.FileInfo) *FileTemplate {
	return &FileTemplate{
		Info: *NewFileInfoTemplate(info),
	}
}
