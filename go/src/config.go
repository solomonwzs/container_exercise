package main

type Configuration struct {
	Name     string `toml:"name"`
	Hostname string `toml:"hostname"`
}
