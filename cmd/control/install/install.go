package install

type MenuEntry struct {
	Name, BootDir, Version, KernelArgs, Append string
}
type BootVars struct {
	BaseName, BootDir string
	Timeout           uint
	Fallback          int
	Entries           []MenuEntry
}
