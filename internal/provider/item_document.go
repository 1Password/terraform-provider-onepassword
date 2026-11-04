package provider

import (
	"encoding/base64"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/1Password/terraform-provider-onepassword/v3/internal/onepassword/model"
)

// documentContent returns the bytes configured in a document block, decoding
// content_base64 when it is the attribute in use.
func documentContent(doc *OnePasswordItemResourceDocumentModel) ([]byte, error) {
	if !doc.ContentBase64.IsNull() && !doc.ContentBase64.IsUnknown() {
		content, err := base64.StdEncoding.DecodeString(doc.ContentBase64.ValueString())
		if err != nil {
			return nil, fmt.Errorf("document content_base64 is not valid base64: %w", err)
		}
		return content, nil
	}
	return []byte(doc.Content.ValueString()), nil
}

// toModelDocument converts the document block into a model file with its content loaded.
func toModelDocument(doc *OnePasswordItemResourceDocumentModel) (*model.ItemFile, error) {
	if doc == nil {
		return nil, nil
	}

	content, err := documentContent(doc)
	if err != nil {
		return nil, err
	}

	file := &model.ItemFile{
		ID:   doc.ID.ValueString(),
		Name: doc.Name.ValueString(),
	}
	file.SetContent(content)
	return file, nil
}

// documentChanged reports whether the configured file name or content differ
// between two document blocks, meaning the file has to be uploaded again.
func documentChanged(plan, state *OnePasswordItemResourceDocumentModel) bool {
	if plan == nil || state == nil {
		return plan != state
	}
	return !plan.Name.Equal(state.Name) ||
		!plan.Content.Equal(state.Content) ||
		!plan.ContentBase64.Equal(state.ContentBase64)
}

// toStateDocument converts a document file into its Terraform representation.
// The content is stored in whichever attribute the prior state or config used.
// Without a prior value (for example after an import), text files use content
// and binary files use content_base64. If the file content wasn't loaded, the
// prior content is kept.
func toStateDocument(doc *model.ItemFile, prior *OnePasswordItemResourceDocumentModel) *OnePasswordItemResourceDocumentModel {
	if doc == nil {
		return nil
	}

	result := &OnePasswordItemResourceDocumentModel{
		ID:            types.StringValue(doc.ID),
		Name:          types.StringValue(doc.Name),
		Content:       types.StringNull(),
		ContentBase64: types.StringNull(),
	}

	content, err := doc.Content()
	if err != nil {
		if prior != nil {
			result.Content = prior.Content
			result.ContentBase64 = prior.ContentBase64
		}
		return result
	}

	useBase64 := !utf8.Valid(content)
	if prior != nil {
		useBase64 = !prior.ContentBase64.IsNull()
	}

	if useBase64 {
		result.ContentBase64 = types.StringValue(base64.StdEncoding.EncodeToString(content))
	} else {
		// Terraform strings must be valid UTF-8. Content that isn't can only show up
		// here if the file was changed outside Terraform, so it is reported as drift.
		result.Content = types.StringValue(strings.ToValidUTF8(string(content), "\uFFFD"))
	}

	return result
}

// carryDocumentContent sets the content of the document returned by the API to the
// content Terraform just wrote, so that state can be saved without downloading
// the file again.
func carryDocumentContent(item *model.Item, doc *OnePasswordItemResourceDocumentModel) error {
	if item == nil || item.Document == nil || doc == nil {
		return nil
	}
	content, err := documentContent(doc)
	if err != nil {
		return err
	}
	item.Document.SetContent(content)
	return nil
}
