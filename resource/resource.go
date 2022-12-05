package resource

type Resource interface {
	Attach() (Resource, error)
	Detach() error
}
