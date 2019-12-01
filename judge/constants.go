package judge

const (
	// NoInteractor tells the judge to execute the problem solution without any interactor.
	NoInteractor = ""

	// DefaultInteractor tells the judge to execute the problem solution through the default
	// interactor, which discretely streams the input of each test case to the solution program and
	// discretely collects the output of each test case.
	DefaultInteractor = "default-interactor"

	// CustomInteractor tells the judge to execute the problem solution through a custom interactor
	// which is supposed to be inside the problem bundle.
	CustomInteractor = "custom-interactor"

	// The folder structure of a problem bundle.
	devNull                      = "/dev/null"
	folderPathBundle             = "./bundle"
	folderPathBundleInputs       = folderPathBundle + "/inputs"
	folderPathBundleOutputs      = folderPathBundle + "/outputs"
	folderPathSolutionOutputs    = "./outputs"
	filePathCompressedBundle     = "./bundle.tar.gz"
	filePathBundleInteractor     = folderPathBundle + "/interactor"
	filePathSolutionSingleOutput = folderPathSolutionOutputs + "/single.txt"
	filesPathPrefixInteractor    = "./interactor"
	filesPathPrefixSolution      = "./solution"
)
