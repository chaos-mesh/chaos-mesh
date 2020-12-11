package template

type SerialTemplate interface {
	Template
	GetSerialChildrenList() []Template
}

func ParseSerialTemplate(raw interface{}) (SerialTemplate, error) {
	panic("unimplemented")
}
