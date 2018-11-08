package main

type Configuration struct {
	Name     string      `toml:"name"`
	Hostname string      `toml:"hostname"`
	Env      []string    `toml:"env"`
	Mount    []CMount    `toml:"mount"`
	BaseSys  CBaseSystem `toml:"base_system"`
	Network  CNetwork    `toml:"network"`
}

type CBaseSystem struct {
	Dir       string `toml:"dir"`
	System    string `toml:"system"`
	Workspace string `toml:"workspace"`
}

type CNetwork struct {
	Interfaces []CNetworkInterface `toml:"interface"`
	Routes     []CNetworkRoute     `toml:"route"`
}

type CNetworkInterface struct {
	HostInterface string `toml:"host_interface"`
	IP            string `toml:"ip"`
	Mask          string `toml:"mask"`
	Mode          string `toml:"mode"`
	Name          string `toml:"name"`
	Type          string `toml:"type"`
}

type CNetworkRoute struct {
	Dest    string `toml:"dest"`
	Gateway string `toml:"gateway"`
	Mask    string `toml:"mask"`
}

type CMount struct {
	Source string `toml:"src"`
	Target string `toml:"target"`
}
