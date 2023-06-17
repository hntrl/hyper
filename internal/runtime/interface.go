package runtime

// RuntimeNode represents anything that requires shared resources to function
type RuntimeNode interface {
	Attach(*Process) error
	Detach() error
}
