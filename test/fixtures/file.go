package fixtures

import "github.com/matheuscscp/fd8-judge/test/factories"

var singleFile = &factories.File{
	Name:    "SingleFile.txt",
	Content: "singleFileContent",
}

var singleFile2 = &factories.File{
	Name:    "SingleFile2.txt",
	Content: "singleFile2Content",
}

var singleFile3 = &factories.File{
	Name:    "SingleFile3.txt",
	Content: "singleFile3Content",
}

var emptyFolder = &factories.Folder{
	Name: "EmptyFolder",
}

var emptyFolder2 = &factories.Folder{
	Name: "EmptyFolder2",
}

var emptyFolder3 = &factories.Folder{
	Name: "EmptyFolder3",
}

var testFolderOneFile = &factories.Folder{
	Name:     "TestFolderOneFile",
	Children: []factories.FileTreeNode{singleFile},
}

var testFolderThreeFiles = &factories.Folder{
	Name: "TestFolderThreeFiles",
	Children: []factories.FileTreeNode{
		singleFile,
		singleFile2,
		singleFile3,
	},
}

var testFolderOneFolder = &factories.Folder{
	Name:     "TestFolderOneFolder",
	Children: []factories.FileTreeNode{emptyFolder},
}

var testFolderThreeFolders = &factories.Folder{
	Name: "TestFolderThreeFolders",
	Children: []factories.FileTreeNode{
		emptyFolder,
		emptyFolder2,
		emptyFolder3,
	},
}

var testFolder = &factories.Folder{
	Name: "TestFolder",
	Children: []factories.FileTreeNode{
		singleFile,
		emptyFolder,
		testFolderOneFile,
		testFolderThreeFiles,
		testFolderOneFolder,
		testFolderThreeFolders,
	},
}

var middleFolder = &factories.Folder{
	Name: "MiddleFolder",
	Children: []factories.FileTreeNode{
		testFolder,
	},
}

var testDummyRootFolder = &factories.Folder{
	Name: "TestDummyRootFolder",
	Children: []factories.FileTreeNode{
		middleFolder,
	},
}

func SingleFile() factories.FileTreeNode {
	return factories.SortFileTree(singleFile.Clone())
}

func SingleFile2() factories.FileTreeNode {
	return factories.SortFileTree(singleFile2.Clone())
}

func SingleFile3() factories.FileTreeNode {
	return factories.SortFileTree(singleFile3.Clone())
}

func EmptyFolder() factories.FileTreeNode {
	return factories.SortFileTree(emptyFolder.Clone())
}

func EmptyFolder2() factories.FileTreeNode {
	return factories.SortFileTree(emptyFolder2.Clone())
}

func EmptyFolder3() factories.FileTreeNode {
	return factories.SortFileTree(emptyFolder3.Clone())
}

func TestFolderOneFile() factories.FileTreeNode {
	return factories.SortFileTree(testFolderOneFile.Clone())
}

func TestFolderThreeFiles() factories.FileTreeNode {
	return factories.SortFileTree(testFolderThreeFiles.Clone())
}

func TestFolderOneFolder() factories.FileTreeNode {
	return factories.SortFileTree(testFolderOneFolder.Clone())
}

func TestFolderThreeFolders() factories.FileTreeNode {
	return factories.SortFileTree(testFolderThreeFolders.Clone())
}

func TestFolder() factories.FileTreeNode {
	return factories.SortFileTree(testFolder.Clone())
}

func MiddleFolder() factories.FileTreeNode {
	return factories.SortFileTree(middleFolder.Clone())
}

func TestDummyRootFolder() factories.FileTreeNode {
	return factories.SortFileTree(testDummyRootFolder.Clone())
}
