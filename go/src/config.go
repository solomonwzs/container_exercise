package main

type Configuration struct {
	Name     string     `toml:"name"`
	Hostname string     `toml:"hostname"`
	Env      []string   `toml:"env"`
	BaseSys  BaseSystem `toml:"base_system"`
}

type BaseSystem struct {
	Dir       string `toml:"dir"`
	System    string `toml:"system"`
	Workspace string `toml:"workspace"`
}
