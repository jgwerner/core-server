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
	rg.commandArgs()
	err := os.Chdir(args.ResourceDir)
	if err != nil {
		return err
	}
	cmd := exec.Command(rg.command, rg.args...)
	cmd.Stdout = out
	cmd.Stderr = out
	return cmd.Run()
}

func (rg *RunGeneric) commandArgs() {
	fargs := flag.Args()
	if len(fargs) > 0 {
		rg.command = fargs[0]
	}
	if len(fargs) > 1 {
		rg.args = fargs[1:]
	}
}
