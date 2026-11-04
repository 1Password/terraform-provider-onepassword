package provider

import (
	"encoding/base64"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/1Password/terraform-provider-onepassword/v3/internal/onepassword/model"
)

func loadedFile(name string, content []byte) *model.ItemFile {
	f := &model.ItemFile{ID: "fileid", Name: name}
	f.SetContent(content)
	return f
}

func TestToStateDocument(t *testing.T) {
	binary := []byte{0x00, 0xff, 0xfe}

	tests := map[string]struct {
		file              *model.ItemFile
		prior             *OnePasswordItemResourceDocumentModel
		wantContent       types.String
		wantContentBase64 types.String
	}{
		"no document": {},
		"import of a text file uses content": {
			file:              loadedFile("a.txt", []byte("hello")),
			wantContent:       types.StringValue("hello"),
			wantContentBase64: types.StringNull(),
		},
		"import of a binary file uses content_base64": {
			file:              loadedFile("a.bin", binary),
			wantContent:       types.StringNull(),
			wantContentBase64: types.StringValue(base64.StdEncoding.EncodeToString(binary)),
		},
		"keeps content_base64 for a text file when config uses it": {
			file: loadedFile("a.txt", []byte("hello")),
			prior: &OnePasswordItemResourceDocumentModel{
				Content:       types.StringNull(),
				ContentBase64: types.StringValue("aGVsbG8="),
			},
			wantContent:       types.StringNull(),
			wantContentBase64: types.StringValue("aGVsbG8="),
		},
		"reports invalid UTF-8 as drift when config uses content": {
			file: loadedFile("a.txt", binary),
			prior: &OnePasswordItemResourceDocumentModel{
				Content:       types.StringValue("text"),
				ContentBase64: types.StringNull(),
			},
			wantContent:       types.StringValue("\x00\uFFFD"),
			wantContentBase64: types.StringNull(),
		},
		"keeps prior content when the file wasn't downloaded": {
			file: &model.ItemFile{ID: "fileid", Name: "a.txt"},
			prior: &OnePasswordItemResourceDocumentModel{
				Content:       types.StringValue("from state"),
				ContentBase64: types.StringNull(),
			},
			wantContent:       types.StringValue("from state"),
			wantContentBase64: types.StringNull(),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := toStateDocument(tc.file, tc.prior)
			if tc.file == nil {
				if got != nil {
					t.Fatalf("expected nil, got %+v", got)
				}
				return
			}
			if got.ID.ValueString() != tc.file.ID || got.Name.ValueString() != tc.file.Name {
				t.Errorf("unexpected id/name: %s/%s", got.ID, got.Name)
			}
			if !got.Content.Equal(tc.wantContent) {
				t.Errorf("content: want %s, got %s", tc.wantContent, got.Content)
			}
			if !got.ContentBase64.Equal(tc.wantContentBase64) {
				t.Errorf("content_base64: want %s, got %s", tc.wantContentBase64, got.ContentBase64)
			}
		})
	}
}

func TestToModelDocument(t *testing.T) {
	file, err := toModelDocument(&OnePasswordItemResourceDocumentModel{
		ID:            types.StringValue("fileid"),
		Name:          types.StringValue("a.bin"),
		Content:       types.StringNull(),
		ContentBase64: types.StringValue("AP/+"),
	})
	if err != nil {
		t.Fatal(err)
	}
	content, err := file.Content()
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "\x00\xff\xfe" || file.Name != "a.bin" || file.ID != "fileid" {
		t.Errorf("unexpected file %+v with content %q", file, content)
	}

	if _, err := toModelDocument(&OnePasswordItemResourceDocumentModel{
		Name:          types.StringValue("a.bin"),
		ContentBase64: types.StringValue("not base64!"),
	}); err == nil {
		t.Error("expected an error for invalid base64")
	}

	if file, err := toModelDocument(nil); file != nil || err != nil {
		t.Errorf("expected nil file and error, got %+v, %v", file, err)
	}
}

func TestDocumentChanged(t *testing.T) {
	doc := func(name, content string) *OnePasswordItemResourceDocumentModel {
		return &OnePasswordItemResourceDocumentModel{
			ID:            types.StringUnknown(),
			Name:          types.StringValue(name),
			Content:       types.StringValue(content),
			ContentBase64: types.StringNull(),
		}
	}

	if documentChanged(doc("a", "x"), doc("a", "x")) {
		t.Error("same name and content should not count as a change")
	}
	if !documentChanged(doc("a", "x"), doc("b", "x")) {
		t.Error("a new name should count as a change")
	}
	if !documentChanged(doc("a", "x"), doc("a", "y")) {
		t.Error("new content should count as a change")
	}
	if !documentChanged(doc("a", "x"), nil) || documentChanged(nil, nil) {
		t.Error("unexpected result for nil documents")
	}
}
