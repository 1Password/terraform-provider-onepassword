package model

import (
	"fmt"
)

func (f *ItemFile) Content() ([]byte, error) {
	if f.content == nil {
		return nil, fmt.Errorf("file content not loaded")
	}
	return f.content, nil
}

func (f *ItemFile) SetContent(content []byte) {
	f.content = content
}
