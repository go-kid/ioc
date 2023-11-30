package creator

type ScanFile struct {
	Path string
}

func (s *ScanFile) Create() error {
	g := NewGoFile("scan", "init", "scan_"+s.Path, false)
	g.SetAttribute("Scan", s.Path)
	return g.Create()
}

type UsePluginFile struct {
	FileName     string
	Paths        []string //absolute path
	Plugins      []string //component(xxx)
	Imports      []string // _ "xxx/xxx"
	Packages     []string //relative path
	Dependencies []string //go get xxx
}

func (u *UsePluginFile) Create() error {
	sf := &ScanFile{Path: "plugin"}
	g := NewGoFile("use_plugin", "plugin", u.FileName, false)
	if u.Paths != nil {
		g.SetAttribute("Paths", u.Paths)
	}
	if u.Plugins != nil {
		g.SetAttribute("Plugin", u.Plugins)
	}
	if u.Imports != nil {
		g.SetAttribute("Imports", u.Imports)
	}
	if u.Packages != nil {
		g.SetAttribute("Packages", u.Packages)
	}
	if u.Dependencies != nil {
		for i := range u.Dependencies {
			cmd.Get(u.Dependencies[i])
		}
	}
	err := NewBatchCreator(sf, g).Create()
	if err != nil {
		return err
	}
	cmd.FmtAll("plugin")
	return nil
}
