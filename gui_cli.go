//go:build cli

package main

type cliApp struct{}

func (c cliApp) Run() error { usage(); return nil }

func createApp() cliApp {
	return cliApp{}
}
