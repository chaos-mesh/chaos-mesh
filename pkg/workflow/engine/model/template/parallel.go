package template

type ParallelTemplate interface {
	Template
	GetParallelChildrenList() []Template
}

func ParseParallelTemplate(raw interface{}) (ParallelTemplate, error) {
	panic("unimplemented")
}
