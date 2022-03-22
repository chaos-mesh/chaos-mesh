package disk

type CommonConfig struct {
	Path string

	Size    string
	Percent string
}

type FillConfig struct {
	CommonConfig
	FillByFAllocate bool
}

type ReadConfig struct {
	CommonConfig
}
