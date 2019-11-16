package factory

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type (
	// FileTreeNode represents a node of the OS file tree.
	FileTreeNode interface {
		// IsFolder returns true if the node is a Folder.
		IsFolder() bool

		// Write writes the whole subtree rooted at the node.
		Write(relativePath string) error
	}

	// File represents an OS file.
	File struct {
		// Name is the file name.
		Name string

		// Content is the file content.
		Content string
	}

	// Folder represents an OS folder.
	Folder struct {
		// Name is the folder name.
		Name string

		// Children are the files and fod
		Children []FileTreeNode
	}
)

// ReadFileTree reads a
func ReadFileTree(relativePath string, readFileContents bool) (FileTreeNode, error) {
	relativePath = filepath.Clean(relativePath)

	info, err := os.Stat(relativePath)
	if err != nil {
		return nil, fmt.Errorf("error Stat()ing relative path to read file tree: %w", err)
	}

	// folder
	if info.IsDir() {
		folder := &Folder{
			Name: info.Name(),
		}
		children, err := ioutil.ReadDir(relativePath)
		if err != nil {
			return nil, fmt.Errorf("error ReadDir()ing relative path to read folder node: %w", err)
		}
		for _, childInfo := range children {
			child, err := ReadFileTree(filepath.Join(relativePath, childInfo.Name()), readFileContents)
			if err != nil {
				return nil, fmt.Errorf("error ReadFileTree()ing child of folder node: %w", err)
			}
			folder.Children = append(folder.Children, child)
		}
		return folder, nil
	}

	// file
	if !readFileContents {
		return &File{Name: info.Name()}, nil
	}
	bytes, err := ioutil.ReadFile(relativePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file node: %w", err)
	}
	return &File{
		Name:    info.Name(),
		Content: string(bytes),
	}, nil
}

// IsFolder returns true.
func (f *Folder) IsFolder() bool {
	return true
}

// Write writes a folder tree.
func (f *Folder) Write(relativePath string) error {
	curPath := filepath.Clean(filepath.Join(relativePath, f.Name))
	if err := os.Mkdir(curPath, os.ModeDir); err != nil {
		return fmt.Errorf("error creating test folder: %w", err)
	}
	for _, child := range f.Children {
		if err := child.Write(curPath); err != nil {
			return fmt.Errorf("error writing test folder children: %w", err)
		}
	}
	return nil
}

// IsFolder returns false.
func (f *File) IsFolder() bool {
	return false
}

// Write writes a file.
func (f *File) Write(relativePath string) error {
	curPath := filepath.Clean(filepath.Join(relativePath, f.Name))
	if err := ioutil.WriteFile(curPath, []byte(f.Content), os.ModePerm); err != nil {
		return fmt.Errorf("error writing test file: %w", err)
	}
	return nil
}
