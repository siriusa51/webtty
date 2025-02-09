package main

import (
	"flag"
	"fmt"
	"github.com/siriusa51/waitprocess/v2/ext/http_srv"
	"github.com/siriusa51/webtty/apis"
	"github.com/siriusa51/webtty/session"
	"log/slog"
	"os"
)

type Args struct {
	apis.RouterConfig
}

func ParseArgs() Args {
	args := Args{}
	flag.StringVar(&args.Host, "host", "localhost", "Host to listen on")
	flag.IntVar(&args.Port, "port", 8080, "Port to listen on")
	flag.StringVar(&args.PrefixPath, "prefix-path", "/", "Prefix path")
	flag.StringVar(&args.IndexFile, "index-file", "", "Index file, if not set, use the default index.html")
	flag.StringVar(&args.Workdir, "workdir", "", "Workdir for the command, default is current directory")
	flag.StringVar(&args.Command, "command", "", "Command to run")
	flag.Parse()

	if args.Command == "" {
		panic("command is required, please specify it with --command")
	}

	return args
}

func main() {
	args := ParseArgs()
	mgr := session.NewSessionManager()
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	router := apis.NewHandler(args.RouterConfig, log, mgr)

	if err := http_srv.RegisterHttpSrv(fmt.Sprintf("%s:%d", args.Host, args.Port), router).
		RegisterSignal(os.Kill, os.Interrupt).
		Run(); err != nil {
		log.Error("failed to start http server", "error", err)
	}
}
