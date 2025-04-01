package main

import (
	"os"

	"github.com/cert-manager/cert-manager/pkg/acme/webhook/cmd"
	"github.com/imroc/cert-manager-webhook-dnspod/dnspod"
)

var (
	GroupName = os.Getenv("GROUP_NAME")
	LogLevel  = os.Getenv("LOG_LEVEL")
)

func main() {
	if GroupName == "" {
		panic("GROUP_NAME must be specified")
	}
	solver := dnspod.NewSolver()
	if LogLevel != "" {
		solver.SetLogLevel(LogLevel)
	}
	cmd.RunWebhookServer(GroupName, solver)
}
