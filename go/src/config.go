package main

type Configuration struct {
	Name     string      `toml:"name"`
	Hostname string      `toml:"hostname"`
	Env      []string    `toml:"env"`
	BaseSys  CBaseSystem `toml:"base_system"`
	Networks []CNetwork  `toml:"network"`
}

type CBaseSystem struct {
	Dir       string `toml:"dir"`
	System    string `toml:"system"`
	Workspace string `toml:"workspace"`
}

type CNetwork struct {
	Name string `toml:"name"`
	Type string `toml:"type"`
	Mask string `toml:"mask"`
	IP   string `toml:"ip"`
}
