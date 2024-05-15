// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/hashicorp/cli"
	tfjson "github.com/hashicorp/terraform-json"

	"github.com/hashicorp/terraform-plugin-docs/internal/check"
)

const (
	FileExtensionHtmlMarkdown = `.html.markdown`
	FileExtensionHtmlMd       = `.html.md`
	FileExtensionMarkdown     = `.markdown`
	FileExtensionMd           = `.md`

	DocumentationGlobPattern    = `{docs/index.md,docs/{,cdktf/}{data-sources,guides,resources,functions}/**/*,website/docs/**/*}`
	DocumentationDirGlobPattern = `{docs/{,cdktf/}{data-sources,guides,resources,functions}{,/*},website/docs/**/*}`
)

var ValidLegacyFileExtensions = []string{
	FileExtensionHtmlMarkdown,
	FileExtensionHtmlMd,
	FileExtensionMarkdown,
	FileExtensionMd,
}

var ValidRegistryFileExtensions = []string{
	FileExtensionMd,
}

var LegacyFrontMatterOptions = &check.FrontMatterOptions{
	NoSidebarCurrent:   true,
	RequireDescription: true,
	RequireLayout:      true,
	RequirePageTitle:   true,
}

var LegacyIndexFrontMatterOptions = &check.FrontMatterOptions{
	NoSidebarCurrent:   true,
	NoSubcategory:      true,
	RequireDescription: true,
	RequireLayout:      true,
	RequirePageTitle:   true,
}

var LegacyGuideFrontMatterOptions = &check.FrontMatterOptions{
	NoSidebarCurrent:   true,
	RequireDescription: true,
	RequireLayout:      true,
	RequirePageTitle:   true,
}

var RegistryFrontMatterOptions = &check.FrontMatterOptions{
	NoLayout:         true,
	NoSidebarCurrent: true,
}

var RegistryIndexFrontMatterOptions = &check.FrontMatterOptions{
	NoLayout:         true,
	NoSidebarCurrent: true,
	NoSubcategory:    true,
}

var RegistryGuideFrontMatterOptions = &check.FrontMatterOptions{
	NoLayout:         true,
	NoSidebarCurrent: true,
	RequirePageTitle: true,
}

type validator struct {
	providerName        string
	providerDir         string
	providersSchemaPath string

	tfVersion      string
	providerSchema *tfjson.ProviderSchema

	logger *Logger
}

func Validate(ui cli.Ui, providerDir, providerName, providersSchemaPath, tfversion string) error {
	// Ensure provider directory is resolved absolute path
	if providerDir == "" {
		wd, err := os.Getwd()

		if err != nil {
			return fmt.Errorf("error getting working directory: %w", err)
		}

		providerDir = wd
	} else {
		absProviderDir, err := filepath.Abs(providerDir)

		if err != nil {
			return fmt.Errorf("error getting absolute path with provider directory %q: %w", providerDir, err)
		}

		providerDir = absProviderDir
	}

	// Verify provider directory
	providerDirFileInfo, err := os.Stat(providerDir)

	if err != nil {
		return fmt.Errorf("error getting information for provider directory %q: %w", providerDir, err)
	}

	if !providerDirFileInfo.IsDir() {
		return fmt.Errorf("expected %q to be a directory", providerDir)
	}

	v := &validator{
		providerName:        providerName,
		providerDir:         providerDir,
		providersSchemaPath: providersSchemaPath,
		tfVersion:           tfversion,

		logger: NewLogger(ui),
	}

	ctx := context.Background()

	return v.validate(ctx)
}

func (v *validator) validate(ctx context.Context) error {
	var result error

	var err error

	if v.providersSchemaPath == "" {
		v.logger.infof("exporting schema from Terraform")
		v.providerSchema, err = TerraformProviderSchemaFromTerraform(ctx, v.providerName, v.providerDir, v.tfVersion, v.logger)
		if err != nil {
			return fmt.Errorf("error exporting provider schema from Terraform: %w", err)
		}
	} else {
		v.logger.infof("exporting schema from JSON file")
		v.providerSchema, err = TerraformProviderSchemaFromFile(v.providerName, v.providersSchemaPath, v.logger)
		if err != nil {
			return fmt.Errorf("error exporting provider schema from JSON file: %w", err)
		}
	}

	providerFs := os.DirFS(v.providerDir)

	files, globErr := doublestar.Glob(providerFs, DocumentationGlobPattern)
	if globErr != nil {
		return fmt.Errorf("error finding documentation files: %w", err)
	}

	log.Printf("[DEBUG] Found documentation files %v", files)

	v.logger.infof("running mixed directories check")
	err = check.MixedDirectoriesCheck(files)
	result = errors.Join(result, err)

	v.logger.infof("running number of files check")
	err = check.NumberOfFilesCheck(files)
	result = errors.Join(result, err)

	if dirExists(filepath.Join(v.providerDir, "docs")) {
		v.logger.infof("detected static docs directory, running checks")
		err = v.validateStaticDocs(filepath.Join(v.providerDir, "docs"))
		result = errors.Join(result, err)

	}
	if dirExists(filepath.Join(v.providerDir, filepath.Join("website", "docs"))) {
		v.logger.infof("detected legacy website directory, running checks")
		err = v.validateLegacyWebsite(filepath.Join(v.providerDir, "website/docs"))
		result = errors.Join(result, err)
	}

	return result
}

func (v *validator) validateStaticDocs(dir string) error {

	var result error

	options := &check.ProviderFileOptions{
		FrontMatter:     RegistryFrontMatterOptions,
		ValidExtensions: ValidRegistryFileExtensions,
	}

	var files []string

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error walking directory %q: %w", dir, err)
		}

		rel, err := filepath.Rel(v.providerDir, path)
		if err != nil {
			return err
		}
		if d.IsDir() {
			match, err := doublestar.PathMatch(filepath.FromSlash(DocumentationDirGlobPattern), rel)
			if err != nil {
				return err
			}
			if !match {
				return nil // skip valid non-documentation directories
			}

			v.logger.infof("running invalid directories check on %s", rel)
			result = errors.Join(result, check.InvalidDirectoriesCheck(rel))
			return nil
		}
		match, err := doublestar.PathMatch(filepath.FromSlash(DocumentationGlobPattern), rel)
		if err != nil {
			return err
		}
		if !match {
			return nil // skip valid non-documentation files
		}

		// Configure FrontMatterOptions based on file type
		if d.Name() == "index.md" {
			options.FrontMatter = RegistryIndexFrontMatterOptions
		} else if _, relErr := filepath.Rel(rel, "guides"); relErr != nil {
			options.FrontMatter = RegistryGuideFrontMatterOptions
		} else {
			options.FrontMatter = RegistryFrontMatterOptions
		}
		v.logger.infof("running file checks on %s", rel)
		result = errors.Join(result, check.NewProviderFileCheck(options).Run(path))

		files = append(files, path)
		return nil
	})
	if err != nil {
		return fmt.Errorf("error walking directory %q: %w", dir, err)
	}

	mismatchOpt := &check.FileMismatchOptions{
		ProviderShortName: providerShortName(v.providerName),
		Schema:            v.providerSchema,
	}

	if dirExists(filepath.Join(dir, "data-sources")) {
		dataSourceFiles, _ := os.ReadDir(filepath.Join(dir, "data-sources"))
		mismatchOpt.DatasourceEntries = dataSourceFiles
	}
	if dirExists(filepath.Join(dir, "resources")) {
		resourceFiles, _ := os.ReadDir(filepath.Join(dir, "resources"))
		mismatchOpt.ResourceEntries = resourceFiles
	}
	if dirExists(filepath.Join(dir, "functions")) {
		functionFiles, _ := os.ReadDir(filepath.Join(dir, "functions"))
		mismatchOpt.FunctionEntries = functionFiles
	}

	v.logger.infof("running file mismatch check")
	if err := check.NewFileMismatchCheck(mismatchOpt).Run(); err != nil {
		result = errors.Join(result, err)
	}

	return result
}

func (v *validator) validateLegacyWebsite(dir string) error {

	var result error

	options := &check.ProviderFileOptions{
		FrontMatter:     LegacyFrontMatterOptions,
		ValidExtensions: ValidLegacyFileExtensions,
	}

	var files []string
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error walking directory %q: %w", dir, err)
		}

		rel, err := filepath.Rel(v.providerDir, path)
		if err != nil {
			return err
		}
		if d.IsDir() {
			match, err := doublestar.PathMatch(filepath.FromSlash(DocumentationDirGlobPattern), rel)
			if err != nil {
				return err
			}
			if !match {
				return nil // skip valid non-documentation directories
			}

			v.logger.infof("running invalid directories check on %s", rel)
			result = errors.Join(result, check.InvalidDirectoriesCheck(rel))
			return nil
		}

		match, err := doublestar.PathMatch(filepath.FromSlash(DocumentationGlobPattern), rel)
		if err != nil {
			return err
		}
		if !match {
			return nil // skip non-documentation files
		}

		// Configure FrontMatterOptions based on file type
		if d.Name() == "index.md" {
			options.FrontMatter = LegacyIndexFrontMatterOptions
		} else if _, relErr := filepath.Rel(rel, "guides"); relErr != nil {
			options.FrontMatter = LegacyGuideFrontMatterOptions
		} else {
			options.FrontMatter = LegacyFrontMatterOptions
		}
		v.logger.infof("running file checks on %s", rel)
		result = errors.Join(result, check.NewProviderFileCheck(options).Run(path))

		files = append(files, path)
		return nil
	})
	if err != nil {
		return fmt.Errorf("error walking directory %q: %w", dir, err)
	}

	mismatchOpt := &check.FileMismatchOptions{
		ProviderShortName: providerShortName(v.providerName),
		Schema:            v.providerSchema,
	}

	if dirExists(filepath.Join(dir, "d")) {
		dataSourceFiles, _ := os.ReadDir(filepath.Join(dir, "d"))
		mismatchOpt.DatasourceEntries = dataSourceFiles
	}
	if dirExists(filepath.Join(dir, "r")) {
		resourceFiles, _ := os.ReadDir(filepath.Join(dir, "r"))
		mismatchOpt.ResourceEntries = resourceFiles
	}
	if dirExists(filepath.Join(dir, "functions")) {
		functionFiles, _ := os.ReadDir(filepath.Join(dir, "functions"))
		mismatchOpt.FunctionEntries = functionFiles
	}

	v.logger.infof("running file mismatch check")
	if err := check.NewFileMismatchCheck(mismatchOpt).Run(); err != nil {
		result = errors.Join(result, err)
	}

	return result
}

func dirExists(name string) bool {
	if file, err := os.Stat(name); err != nil {
		return false
	} else if !file.IsDir() {
		return false
	}

	return true
}
