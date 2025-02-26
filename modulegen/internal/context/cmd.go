package context

// TestcontainersModuleVar is a struct that contains the name, title and image of a testcontainers module.
// It's used to hold the input values from the command line.
type TestcontainersModuleVar struct {
	// Name is the name of the module.
	Name string

	// NameTitle is the title of the module.
	NameTitle string

	// Image is the image of the module.
	Image string
}
