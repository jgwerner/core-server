package main

import (
	"context"
	"fmt"
)

type RunCode struct{}

func (rp *RunCode) Run() error {
	RunKernelGateway(out, out, args.KernelName)
	GetKernel()
	stats := NewStats()
	_, err := Run(context.Background(), stats, args.Script, fmt.Sprintf("%s()", args.Function))
	if err != nil {
		SetStatus(args, "Error")
	}
	stats.Send(args)
	return err
}
