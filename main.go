package main

import (
	"flag"
	"io"
	"log"
	"os"
)

var (
	out    io.Writer = os.Stderr
	logger           = log.New(out, "", log.Lshortfile)
	args             = &Args{}
)

func main() {
	args.ResourceDir = "/resources"
	flag.StringVar(&args.ApiKey, "key", "", "Api Key")
	flag.StringVar(&args.Namespace, "ns", "", "Namespace")
	flag.StringVar(&args.ProjectID, "projectID", "", "Project id")
	flag.StringVar(&args.ServerID, "serverID", "", "Server id")
	flag.StringVar(&args.KernelName, "kernel", "python", "Kernel name")
	flag.StringVar(&args.ServerType, "type", "", "Server type")
	flag.StringVar(&args.ApiRoot, "root", "", "Api Root domain")
	flag.StringVar(&args.Script, "script", "", "Script to run")
	flag.StringVar(&args.Function, "function", "", "Function to run")
	flag.Parse()
	if args.KernelName == "" {
		args.KernelName = os.Getenv("KERNEL_NAME")
	}
	if args.ServerType == "" {
		args.ServerType = os.Getenv("SERVER_TYPE")
	}
	SetKernelName(args.KernelName)
	err := os.Chdir(args.ResourceDir)
	if err != nil {
		go SetStatus(args, "Error")
		logger.Fatal(err)
	}
	err = StartScript()
	if err != nil {
		go SetStatus(args, "Error")
		logger.Fatalf("[StartScript]: %s", err)
	}
	err = CreateSSHTunnels(args)
	if err != nil {
		logger.Printf("[SSH tunnel]: %s", err)
	}
	go SetStatus(args, "Running")
	logger.Fatal(getRunner(args.ServerType).Run())
}
