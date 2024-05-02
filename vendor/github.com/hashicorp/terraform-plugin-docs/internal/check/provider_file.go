// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package check

import (
	"fmt"
	"log"
	"os"
)

type ProviderFileOptions struct {
	*FileOptions

	FrontMatter     *FrontMatterOptions
	ValidExtensions []string
}

type ProviderFileCheck struct {
	Options *ProviderFileOptions
}

func NewProviderFileCheck(opts *ProviderFileOptions) *ProviderFileCheck {
	check := &ProviderFileCheck{
		Options: opts,
	}

	if check.Options == nil {
		check.Options = &ProviderFileOptions{}
	}

	if check.Options.FileOptions == nil {
		check.Options.FileOptions = &FileOptions{}
	}

	if check.Options.FrontMatter == nil {
		check.Options.FrontMatter = &FrontMatterOptions{}
	}

	return check
}

func (check *ProviderFileCheck) Run(path string) error {
	fullpath := check.Options.FullPath(path)

	log.Printf("[DEBUG] Checking file: %s", fullpath)

	if err := FileExtensionCheck(path, check.Options.ValidExtensions); err != nil {
		return fmt.Errorf("%s: error checking file extension: %w", path, err)
	}

	if err := FileSizeCheck(fullpath); err != nil {
		return fmt.Errorf("%s: error checking file size: %w", path, err)
	}

	content, err := os.ReadFile(fullpath)

	if err != nil {
		return fmt.Errorf("%s: error reading file: %w", path, err)
	}

	if err := NewFrontMatterCheck(check.Options.FrontMatter).Run(content); err != nil {
		return fmt.Errorf("%s: error checking file frontmatter: %w", path, err)
	}

	return nil
}
