package factories

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
)

type (
	// FileTreeNode represents a node of the OS file tree.
	FileTreeNode interface {
		// GetName returns the name of the node.
		GetName() string

		// IsFolder returns true if the node is a Folder.
		IsFolder() bool

		// Write writes the whole subtree rooted at the node prepending the path with the given relative
		// path.
		Write(relativePath string) error

		// Clone clones the subtree rooted at the node.
		Clone() FileTreeNode

		// SortChildren sorts the children of the node and returns the node.
		SortChildren() FileTreeNode

		// Remove removes the path obtained by prepending the node name with the given relative path.
		Remove(relativePath string) error
	}

	// Folder represents an OS folder.
	Folder struct {
		// Name is the folder name.
		Name string

		// Children are the files and fod
		Children []FileTreeNode
	}

	// File represents an OS file.
	File struct {
		// Name is the file name.
		Name string

		// Content is the file content.
		Content string
	}
)

// GetName returns Name.
func (f *Folder) GetName() string {
	return f.Name
}

// IsFolder returns true.
func (f *Folder) IsFolder() bool {
	return true
}

// Write writes the folder.
func (f *Folder) Write(relativePath string) error {
	curPath := filepath.Clean(relativePath) + "/" + f.Name
	if err := os.Mkdir(curPath, os.ModeDir|0755); err != nil {
		return fmt.Errorf("error creating test folder: %w", err)
	}
	for _, child := range f.Children {
		if err := child.Write(curPath); err != nil {
			return fmt.Errorf("error writing test folder children: %w", err)
		}
	}
	return nil
}

// Clone clones the subtree rooted at the folder.
func (f *Folder) Clone() FileTreeNode {
	if len(f.Children) == 0 {
		return &Folder{Name: f.Name}
	}
	children := make([]FileTreeNode, len(f.Children))
	for i, child := range f.Children {
		children[i] = child.Clone()
	}
	return &Folder{
		Name:     f.Name,
		Children: children,
	}
}

// SortChildren sorts the children of the folder and returns the folder.
func (f *Folder) SortChildren() FileTreeNode {
	if len(f.Children) > 0 {
		sort.Slice(f.Children, func(i, j int) bool {
			return f.Children[i].GetName() < f.Children[j].GetName()
		})
	}
	return f
}

// Remove removes the folder.
func (f *Folder) Remove(relativePath string) error {
	return os.RemoveAll(filepath.Join(filepath.Clean(relativePath), f.Name))
}

// GetName returns Name.
func (f *File) GetName() string {
	return f.Name
}

// IsFolder returns false.
func (f *File) IsFolder() bool {
	return false
}

// Write writes the file.
func (f *File) Write(relativePath string) error {
	curPath := filepath.Clean(relativePath) + "/" + f.Name
	if err := ioutil.WriteFile(curPath, []byte(f.Content), os.ModePerm); err != nil {
		return fmt.Errorf("error writing test file: %w", err)
	}
	return nil
}

// Clone clones the file.
func (f *File) Clone() FileTreeNode {
	return &File{
		Name:    f.Name,
		Content: f.Content,
	}
}

// SortChildren only returns the file.
func (f *File) SortChildren() FileTreeNode {
	return f
}

// Remove removes the file.
func (f *File) Remove(relativePath string) error {
	return os.RemoveAll(filepath.Join(filepath.Clean(relativePath), f.Name))
}

// ReadFileTree reads a file tree rooted at the given relative path.
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
		return folder.SortChildren(), nil
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

// SortFileTree sorts a whole file subtree and returns it.
func SortFileTree(subroot FileTreeNode) FileTreeNode {
	subroot.SortChildren()
	if subroot.IsFolder() {
		folder, _ := subroot.(*Folder)
		for _, node := range folder.Children {
			SortFileTree(node)
		}
	}
	return subroot
}
