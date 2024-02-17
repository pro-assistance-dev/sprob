package models

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type FileInfo struct {
	bun.BaseModel  `bun:"file_infos,alias:file_infos"`
	ID             uuid.NullUUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id" `
	OriginalName   string        `json:"originalName"`
	FileSystemPath string        `json:"fileSystemPath"`
}

type FileInfos []*FileInfo

func (item FileInfo) GetOriginalName() string {
	return item.OriginalName
}

func (item FileInfo) GetFullPath() string {
	return item.FileSystemPath
}

func (items FileInfos) GetPathsAndNames() (paths []string, names []string) {
	for _, item := range items {
		paths = append(paths, item.FileSystemPath)
		names = append(names, item.OriginalName)
	}
	return paths, names
}
