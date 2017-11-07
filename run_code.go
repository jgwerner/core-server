package main

import (
	"context"
	"fmt"
)

type RunCode struct{}

func (rp *RunCode) Run() error {
	RunKernelGateway(out, out, args.KernelName)
	GetKernel()
	_, _, err := Run(context.Background(), args.Script, fmt.Sprintf("%s()", args.Function))
	if err != nil {
		return err
	}
	return err
}
