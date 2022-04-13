package main

import (
	"context"
	"fast/fileservice"
	"fast/server"
	"fast/util"
	"flag"
	"fmt"
	"net/http"
	"os"
)

var (
	port  int
	dir   string
	depth int
	write bool
)

func init() {
	flag.IntVar(&port, "p", 8081, "port to listen on")
	flag.IntVar(&depth, "depth", 0, "file depth, 0 for infinite")
	flag.BoolVar(&write, "w", false, "allow write file to server")

	flag.Usage = func() {
		fmt.Println("Usage: fast [options] [dir]")
		fmt.Printf("\nOptions:\n")
		flag.PrintDefaults()
	}
}

func main() {
	var err error
	flag.Parse()

	dir = flag.Arg(0)
	if dir == "" {
		dir, err = os.Getwd()
		if err != nil {
			fmt.Println("get current dir error:", err)
			return
		}
	}

	err = server.ImportTemplates()
	if err != nil {
		fmt.Println(err)
		return
	}

	fileservice.FS.Init(dir, write, depth)
	ctx := startServer(context.Background(), fmt.Sprintf(":%d", port), server.NewHandler())
	<-ctx.Done()
}

func startServer(ctx context.Context, addr string, handler http.Handler) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	srv := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			fmt.Println("server error:", err)
		}
		cancel()
	}()

	go func() {
		fmt.Printf("http server listening on [http://%v:%d] -> [%s]\n", util.GetOutboundIp(), port, dir)
		if write {
			fmt.Println("write file allowed")
		}
		fmt.Printf("Press any key to stop\n")
		var s string
		fmt.Scanln(&s)
		cancel()
	}()

	return ctx
}
