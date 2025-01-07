package main

import (
	"fmt"
	"github.com/algorandfoundation/nodekit/cmd"
	"github.com/algorandfoundation/nodekit/ui/style"
	"github.com/charmbracelet/x/ansi"
	"github.com/spf13/cobra/doc"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func copyFile(src, dst string, move bool) error {
	// Open the source file
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Create the destination file
	destinationFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	// Copy the contents from source to destination
	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return err
	}

	// Optionally, sync to ensure writes are flushed to disk
	err = destinationFile.Sync()
	if err != nil {
		return err
	}
	if !move {
		return nil
	}

	return os.Remove(src)
}
func appendString(filePath, content string) error {
	// Open the file in append mode, create it if it doesn't exist
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write the content to the file
	_, err = file.WriteString(content)
	if err != nil {
		return err
	}

	return nil
}

func generateMarkdown() error {
	filePrepender := func(filename string) string {
		return ""
	}
	linkHandler := func(name string) string {
		base := strings.TrimSuffix(name, path.Ext(name))
		if base == cmd.Name {
			return "/README.md"
		}
		return "/man/" + strings.ToLower(base) + ".md"
	}
	return doc.GenMarkdownTreeCustom(cmd.RootCmd, "./man", filePrepender, linkHandler)
}

// replaceBetweenStrings replaces everything between startString and endString with replacementText in the content of the file
func replaceBetweenStrings(filePath, startString, endString, replacementText string) error {
	// Read the file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Convert content to string
	text := string(content)

	// Find the start and end boundaries
	startIndex := strings.Index(text, startString)
	endIndex := strings.Index(text, endString)
	if startIndex == -1 || endIndex == -1 || startIndex >= endIndex {
		return fmt.Errorf("could not find valid boundaries between '%s' and '%s'", startString, endString)
	}

	// Build the new content
	// Preserve everything before and after the boundaries, and insert the replacement text in-between
	newContent := text[:startIndex] + replacementText + text[endIndex+len(endString):]

	// Write the modified content back to the file
	err = os.WriteFile(filePath, []byte(newContent), 0644)
	if err != nil {
		return err
	}

	return nil
}
func updateBanner(filePath string) error {
	textBanner := ansi.Strip(style.BANNER)
	textSplit := strings.Split(textBanner, "\n")
	return replaceBetweenStrings(filePath, textSplit[1], textSplit[len(textSplit)-2], "<img alt=\"Terminal Render\" src=\"/assets/nodekit.png\" width=\"65%\">")
}
func updateBanners(dirPath string, starlight bool) error {
	// Open the directory
	dir, err := os.Open(dirPath)
	if err != nil {
		return err
	}
	defer dir.Close()

	// Read directory entries
	files, err := dir.Readdir(-1) // `-1` reads all entries in the directory
	if err != nil {
		return err
	}

	// Iterate over all files and directories
	for _, file := range files {
		if file.IsDir() {
			fmt.Printf("Skipping directory: %s\n", file.Name())
			continue
		}
		if strings.HasSuffix(file.Name(), ".md") && strings.HasPrefix(file.Name(), cmd.Name) {
			if starlight {
				err = updateStarlightHeadings(dirPath + file.Name())
				if err != nil {
					return err
				}
			} else {
				err = updateBanner(dirPath + file.Name())
				if err != nil {
					return err
				}
			}

		}
	}

	return nil
}

func updateStarlightHeadings(filePath string) error {
	textBanner := ansi.Strip(style.BANNER)
	textSplit := strings.Split(textBanner, "\n")
	return replaceBetweenStrings(filePath, "## nodekit", textSplit[len(textSplit)-2], "## Synopsis")
}

const fmTemplate = `---
title: "%s"
slug: "%s"
---
`

func generateStarlightMarkdown() error {
	filePrepender := func(filename string) string {
		name := filepath.Base(filename)
		base := strings.TrimSuffix(name, path.Ext(name))

		return fmt.Sprintf(fmTemplate, strings.Replace(base, "_", " ", -1), fmt.Sprintf("reference/%s", strings.Replace(base, "_", "/", -1)))
	}
	linkHandler := func(name string) string {
		base := strings.TrimSuffix(name, path.Ext(name))
		return "/reference/" + strings.Replace(base, "_", "/", -1)
	}
	return doc.GenMarkdownTreeCustom(cmd.RootCmd, "./docs/src/content/docs/reference", filePrepender, linkHandler)
}

func main() {
	err := generateMarkdown()
	if err != nil {
		panic(err)
	}

	rootCmdDocPath := fmt.Sprintf("./man/%s.md", cmd.Name)

	err = updateBanners("./man/", false)
	if err != nil {
		panic(err)
	}
	// Add Footer
	footerDocPath := "./assets/footer.md"
	footerBytes, err := os.ReadFile(footerDocPath)
	if err != nil {
		panic(err)
	}
	err = appendString(rootCmdDocPath, "\n"+string(footerBytes))
	if err != nil {
		panic(err)
	}
	err = copyFile(fmt.Sprintf("./man/%s.md", cmd.Name), "./README.md", true)
	if err != nil {
		panic(err)
	}

	err = generateStarlightMarkdown()
	if err != nil {
		panic(err)
	}

	err = updateBanners("./docs/src/content/docs/reference/", true)
}
