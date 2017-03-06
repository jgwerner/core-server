package main

import "context"

type RunCode struct{}

func (rp *RunCode) Run() error {
	stats := NewStats()
	_, err := Run(context.Background(), stats, args.Code)
	if err != nil {
		SetStatus(args, "Error")
	}
	stats.Send(args)
	return err
}
