package template

type TaskTemplate interface {
	Template
	GetAllTemplates() []Template
}

func ParseTaskTemplate(raw interface{}) (TaskTemplate, error) {
	panic("unimplemented")
}
