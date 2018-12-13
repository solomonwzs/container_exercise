package main

import "cnet"

type Configuration struct {
	Name     string        `toml:"name"`
	Hostname string        `toml:"hostname"`
	Env      []string      `toml:"env"`
	Mount    []CMount      `toml:"mount"`
	BaseSys  CBaseSystem   `toml:"base_system"`
	Network  cnet.CNetwork `toml:"network"`
}

type CBaseSystem struct {
	Dir       string `toml:"dir"`
	System    string `toml:"system"`
	Workspace string `toml:"workspace"`
}

type CMount struct {
	Source string `toml:"src"`
	Target string `toml:"target"`
}
