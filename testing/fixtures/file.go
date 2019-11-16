package fixtures

import "github.com/matheuscscp/fd8-judge/testing/factory"

var singleFile = &factory.File{
	Name:    "SingleFile.txt",
	Content: "singleFileContent",
}

var singleFile2 = &factory.File{
	Name:    "SingleFile2.txt",
	Content: "singleFile2Content",
}

var singleFile3 = &factory.File{
	Name:    "SingleFile3.txt",
	Content: "singleFile3Content",
}

var emptyFolder = &factory.Folder{
	Name: "EmptyFolder",
}

var emptyFolder2 = &factory.Folder{
	Name: "EmptyFolder2",
}

var emptyFolder3 = &factory.Folder{
	Name: "EmptyFolder3",
}

var testFolderOneFile = &factory.Folder{
	Name:     "TestFolderOneFile",
	Children: []factory.FileTreeNode{singleFile},
}

var testFolderThreeFiles = &factory.Folder{
	Name: "TestFolderThreeFiles",
	Children: []factory.FileTreeNode{
		singleFile,
		singleFile2,
		singleFile3,
	},
}

var testFolderOneFolder = &factory.Folder{
	Name:     "TestFolderOneFolder",
	Children: []factory.FileTreeNode{emptyFolder},
}

var testFolderThreeFolders = &factory.Folder{
	Name: "TestFolderThreeFolders",
	Children: []factory.FileTreeNode{
		emptyFolder,
		emptyFolder2,
		emptyFolder3,
	},
}

var testFolder = &factory.Folder{
	Name: "TestFolder",
	Children: []factory.FileTreeNode{
		singleFile,
		emptyFolder,
		testFolderOneFile,
		testFolderThreeFiles,
		testFolderOneFolder,
		testFolderThreeFolders,
	},
}

var middleFolder = &factory.Folder{
	Name: "MiddleFolder",
	Children: []factory.FileTreeNode{
		testFolder,
	},
}

var testDummyRootFolder = &factory.Folder{
	Name: "TestDummyRootFolder",
	Children: []factory.FileTreeNode{
		middleFolder,
	},
}

func SingleFile() factory.FileTreeNode {
	return singleFile.Clone().SortChildren()
}

func SingleFile2() factory.FileTreeNode {
	return singleFile2.Clone().SortChildren()
}

func SingleFile3() factory.FileTreeNode {
	return singleFile3.Clone().SortChildren()
}

func EmptyFolder() factory.FileTreeNode {
	return emptyFolder.Clone().SortChildren()
}

func EmptyFolder2() factory.FileTreeNode {
	return emptyFolder2.Clone().SortChildren()
}

func EmptyFolder3() factory.FileTreeNode {
	return emptyFolder3.Clone().SortChildren()
}

func TestFolderOneFile() factory.FileTreeNode {
	return testFolderOneFile.Clone().SortChildren()
}

func TestFolderThreeFiles() factory.FileTreeNode {
	return testFolderThreeFiles.Clone().SortChildren()
}

func TestFolderOneFolder() factory.FileTreeNode {
	return testFolderOneFolder.Clone().SortChildren()
}

func TestFolderThreeFolders() factory.FileTreeNode {
	return testFolderThreeFolders.Clone().SortChildren()
}

func TestFolder() factory.FileTreeNode {
	return testFolder.Clone().SortChildren()
}

func MiddleFolder() factory.FileTreeNode {
	return middleFolder.Clone().SortChildren()
}

func TestDummyRootFolder() factory.FileTreeNode {
	return testDummyRootFolder.Clone().SortChildren()
}
