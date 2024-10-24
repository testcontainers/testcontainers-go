package workfile

type ProjectDirectories struct {
	Examples []string
	Modules  []string
}

func newProjectDirectories(examples []string, modules []string) *ProjectDirectories {
	return &ProjectDirectories{
		Examples: examples,
		Modules:  modules,
	}
}
