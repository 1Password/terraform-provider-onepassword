package model

import (
	"fmt"
)

type ItemFile struct {
	ID           string
	Name         string
	Size         int
	SectionID    string
	SectionLabel string
	content      []byte
	ContentPath  string
}

func (f *ItemFile) Content() ([]byte, error) {
	if f.content == nil {
		return nil, fmt.Errorf("file content not loaded")
	}
	return f.content, nil
}

func (f *ItemFile) SetContent(content []byte) {
	f.content = content
}
