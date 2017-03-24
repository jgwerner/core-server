package main

import (
	"flag"
	"os"
	"os/exec"
)

type Runner interface {
	Run() error
}

type RunGeneric struct {
	command string
	args    []string
}

func (rg *RunGeneric) Run() error {
	if rg.command == "" {
		rg.commandArgs()
	}
	err := os.Chdir(args.ResourceDir)
	if err != nil {
		go SetStatus(args, "Error")
		return err
	}
	cmd := exec.Command(rg.command, rg.args...)
	cmd.Stdout = out
	cmd.Stderr = out
	go SetStatus(args, "Running")
	return cmd.Run()
}

func (rg *RunGeneric) commandArgs() {
	fargs := flag.Args()
	rg.command = fargs[0]
	rg.args = fargs[1:]
}
