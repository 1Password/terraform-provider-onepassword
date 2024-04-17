// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package check

import (
	"fmt"
	"log"
	"path/filepath"
)

const (
	CdktfIndexDirectory = `cdktf`

	LegacyIndexDirectory       = `website/docs`
	LegacyDataSourcesDirectory = `d`
	LegacyGuidesDirectory      = `guides`
	LegacyResourcesDirectory   = `r`
	LegacyFunctionsDirectory   = `functions`

	RegistryIndexDirectory       = `docs`
	RegistryDataSourcesDirectory = `data-sources`
	RegistryGuidesDirectory      = `guides`
	RegistryResourcesDirectory   = `resources`
	RegistryFunctionsDirectory   = `functions`

	// Terraform Registry Storage Limits
	// https://www.terraform.io/docs/registry/providers/docs.html#storage-limits
	RegistryMaximumNumberOfFiles = 2000
	RegistryMaximumSizeOfFile    = 500000 // 500KB

)

var ValidLegacyDirectories = []string{
	LegacyIndexDirectory,
	LegacyIndexDirectory + "/" + LegacyDataSourcesDirectory,
	LegacyIndexDirectory + "/" + LegacyGuidesDirectory,
	LegacyIndexDirectory + "/" + LegacyResourcesDirectory,
	LegacyIndexDirectory + "/" + LegacyFunctionsDirectory,
}

var ValidRegistryDirectories = []string{
	RegistryIndexDirectory,
	RegistryIndexDirectory + "/" + RegistryDataSourcesDirectory,
	RegistryIndexDirectory + "/" + RegistryGuidesDirectory,
	RegistryIndexDirectory + "/" + RegistryResourcesDirectory,
	RegistryIndexDirectory + "/" + RegistryFunctionsDirectory,
}

var ValidCdktfLanguages = []string{
	"csharp",
	"go",
	"java",
	"python",
	"typescript",
}

var ValidLegacySubdirectories = []string{
	LegacyIndexDirectory,
	LegacyDataSourcesDirectory,
	LegacyGuidesDirectory,
	LegacyResourcesDirectory,
}

var ValidRegistrySubdirectories = []string{
	RegistryIndexDirectory,
	RegistryDataSourcesDirectory,
	RegistryGuidesDirectory,
	RegistryResourcesDirectory,
}

func InvalidDirectoriesCheck(dirPath string) error {
	if IsValidRegistryDirectory(dirPath) {
		return nil
	}

	if IsValidLegacyDirectory(dirPath) {
		return nil
	}

	if IsValidCdktfDirectory(dirPath) {
		return nil
	}

	return fmt.Errorf("invalid Terraform Provider documentation directory found: %s", dirPath)

}

func MixedDirectoriesCheck(docFiles []string) error {
	var legacyDirectoryFound bool
	var registryDirectoryFound bool
	err := fmt.Errorf("mixed Terraform Provider documentation directory layouts found, must use only legacy or registry layout")

	for _, file := range docFiles {
		directory := filepath.Dir(file)
		log.Printf("[DEBUG] Found directory: %s", directory)

		// Allow docs/ with other files
		if IsValidRegistryDirectory(directory) && directory != RegistryIndexDirectory {
			registryDirectoryFound = true

			if legacyDirectoryFound {
				log.Printf("[DEBUG] Found mixed directories")
				return err
			}
		}

		if IsValidLegacyDirectory(directory) {
			legacyDirectoryFound = true

			if registryDirectoryFound {
				log.Printf("[DEBUG] Found mixed directories")
				return err
			}
		}
	}

	return nil
}

// NumberOfFilesCheck verifies that documentation is below the Terraform Registry storage limit.
// This check presumes that all provided directories are valid, e.g. that directory checking
// for invalid or mixed directory structures was previously completed.
func NumberOfFilesCheck(docFiles []string) error {
	var numberOfFiles int

	directoryCounts := make(map[string]int)
	for _, file := range docFiles {
		directory := filepath.Dir(file)

		// Ignore CDKTF files. The file limit is per-language and presumably there is one CDKTF file per source HCL file.
		if IsValidCdktfDirectory(directory) {
			continue
		}

		if directory == RegistryIndexDirectory || directory == filepath.FromSlash(LegacyIndexDirectory) {
			continue
		}

		directoryCounts[directory]++
	}

	for directory, count := range directoryCounts {

		log.Printf("[TRACE] Found %d documentation files in directory: %s", count, directory)
		numberOfFiles = numberOfFiles + count
	}

	log.Printf("[DEBUG] Found %d documentation files with limit of %d", numberOfFiles, RegistryMaximumNumberOfFiles)
	if numberOfFiles >= RegistryMaximumNumberOfFiles {
		return fmt.Errorf("exceeded maximum (%d) number of documentation files for Terraform Registry: %d", RegistryMaximumNumberOfFiles, numberOfFiles)
	}

	return nil
}

func IsValidLegacyDirectory(directory string) bool {
	for _, validLegacyDirectory := range ValidLegacyDirectories {
		if directory == filepath.FromSlash(validLegacyDirectory) {
			return true
		}
	}

	return false
}

func IsValidRegistryDirectory(directory string) bool {
	for _, validRegistryDirectory := range ValidRegistryDirectories {
		if directory == filepath.FromSlash(validRegistryDirectory) {
			return true
		}
	}

	return false
}

func IsValidCdktfDirectory(directory string) bool {
	if directory == filepath.FromSlash(fmt.Sprintf("%s/%s", LegacyIndexDirectory, CdktfIndexDirectory)) {
		return true
	}

	if directory == filepath.FromSlash(fmt.Sprintf("%s/%s", RegistryIndexDirectory, CdktfIndexDirectory)) {
		return true
	}

	for _, validCdktfLanguage := range ValidCdktfLanguages {

		if directory == filepath.FromSlash(fmt.Sprintf("%s/%s/%s", LegacyIndexDirectory, CdktfIndexDirectory, validCdktfLanguage)) {
			return true
		}

		if directory == filepath.FromSlash(fmt.Sprintf("%s/%s/%s", RegistryIndexDirectory, CdktfIndexDirectory, validCdktfLanguage)) {
			return true
		}

		for _, validLegacySubdirectory := range ValidLegacySubdirectories {
			if directory == filepath.FromSlash(fmt.Sprintf("%s/%s/%s/%s", LegacyIndexDirectory, CdktfIndexDirectory, validCdktfLanguage, validLegacySubdirectory)) {
				return true
			}
		}

		for _, validRegistrySubdirectory := range ValidRegistrySubdirectories {
			if directory == filepath.FromSlash(fmt.Sprintf("%s/%s/%s/%s", RegistryIndexDirectory, CdktfIndexDirectory, validCdktfLanguage, validRegistrySubdirectory)) {
				return true
			}
		}
	}

	return false
}
