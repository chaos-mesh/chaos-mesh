package disk

type CommonConfig struct {
	Path string

	Size    string
	Percent string

	SpaceLock string
}

type FillConfig struct {
	CommonConfig
	FillByFAllocate bool
}

type RuntimeConfig struct {
	ProcessNum    uint8
	LoopExecution bool
}

type PayloadAction int

const (
	Read PayloadAction = iota
	Write
)

type PayloadConfig struct {
	Action PayloadAction
	CommonConfig
	RuntimeConfig
}
