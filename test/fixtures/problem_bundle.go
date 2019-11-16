package fixtures

import "github.com/matheuscscp/fd8-judge/test/factories"

var problemBundleHelloWorld = &factories.Folder{
	Name: "bundle",
	Children: []factories.FileTreeNode{
		&factories.Folder{
			Name: "outputs",
			Children: []factories.FileTreeNode{
				&factories.File{
					Name:    "single.txt",
					Content: "hello, world!\n",
				},
			},
		},
	},
}

var problemBundleHelloPerson = &factories.Folder{
	Name: "bundle",
	Children: []factories.FileTreeNode{
		&factories.Folder{
			Name: "inputs",
			Children: []factories.FileTreeNode{
				&factories.File{
					Name:    "mario.txt",
					Content: "mario\n",
				},
				&factories.File{
					Name:    "luigi.txt",
					Content: "luigi\n",
				},
			},
		},
		&factories.Folder{
			Name: "outputs",
			Children: []factories.FileTreeNode{
				&factories.File{
					Name:    "mario.txt",
					Content: "hello, mario!\n",
				},
				&factories.File{
					Name:    "luigi.txt",
					Content: "hello, luigi!\n",
				},
			},
		},
	},
}

var problemBundleHelloDefaultInteractor = &factories.Folder{
	Name: "bundle",
	Children: []factories.FileTreeNode{
		&factories.Folder{
			Name: "inputs",
			Children: []factories.FileTreeNode{
				&factories.File{
					Name:    "single.txt",
					Content: "1\nmario\n1\nluigi\n",
				},
			},
		},
		&factories.Folder{
			Name: "outputs",
			Children: []factories.FileTreeNode{
				&factories.File{
					Name:    "single.txt",
					Content: "hello, mario!\nhello, luigi!\n",
				},
			},
		},
	},
}

var problemBundleHelloCustomInteractor = &factories.Folder{
	Name: "bundle",
	Children: []factories.FileTreeNode{
		&factories.Folder{
			Name: "inputs",
			Children: []factories.FileTreeNode{
				&factories.File{
					Name:    "single.txt",
					Content: "mario\nluigi\n",
				},
			},
		},
		&factories.Folder{
			Name: "outputs",
			Children: []factories.FileTreeNode{
				&factories.File{
					Name:    "single.txt",
					Content: "hello, mario!\nhello, luigi!\n",
				},
			},
		},
		&factories.File{
			Name: "interactor",
			Content: `#include <iostream>
#include <fstream>

using namespace std;

int main() {
	string inputPath, outputPath;
	getline(cin, inputPath);
	getline(cin, outputPath);

	fstream inputFile(inputPath.c_str());
	ofstream outputFile(outputPath.c_str());

	string in, out;
	while (inputFile >> in) {
		cout << in << endl << flush;
		getline(cin, out);
		outputFile << out << endl;
	}

	return 0;
}
`,
		},
	},
}

func ProblemBundleHelloWorld() factories.FileTreeNode {
	return factories.SortFileTree(problemBundleHelloWorld.Clone())
}

func ProblemBundleHelloPerson() factories.FileTreeNode {
	return factories.SortFileTree(problemBundleHelloPerson.Clone())
}

func ProblemBundleHelloDefaultInteractor() factories.FileTreeNode {
	return factories.SortFileTree(problemBundleHelloDefaultInteractor.Clone())
}

func ProblemBundleHelloCustomInteractor() factories.FileTreeNode {
	return factories.SortFileTree(problemBundleHelloCustomInteractor.Clone())
}
