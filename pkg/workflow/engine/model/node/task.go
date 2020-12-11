package node

type TaskNode interface {
	Node

	FetchAvailableChildren() ([]string, error)
}


func ParseTaskNode(raw interface{}) (TaskNode, error) {
	panic("unimplemented")
}
