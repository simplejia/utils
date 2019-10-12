package utils

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

var GRACEFUL_ENV = "GRACEFUL=true"

type Server struct {
	server   *http.Server
	listener net.Listener

	isGraceful   bool
	shutdownTime time.Duration
	startTime    time.Duration
	signalChan   chan os.Signal
}

func NewServer(addr string, handler http.Handler, shutdownTime, startTime, timeout time.Duration) *Server {
	isGraceful := false

	GRACEFUL_ENV = filepath.Base(os.Args[0]) + "_" + GRACEFUL_ENV

	for _, v := range os.Environ() {
		if v == GRACEFUL_ENV {
			isGraceful = true
			break
		}
	}

	return &Server{
		server: &http.Server{
			Addr:         addr,
			Handler:      handler,
			ReadTimeout:  timeout,
			WriteTimeout: timeout,
		},

		isGraceful:   isGraceful,
		signalChan:   make(chan os.Signal),
		shutdownTime: shutdownTime,
		startTime:    startTime,
	}
}

func (srv *Server) ListenAndServe() (err error) {
	var ln net.Listener

	if srv.isGraceful {
		file := os.NewFile(3, "")
		ln, err = net.FileListener(file)
		if err != nil {
			return
		}
	} else {
		ln, err = net.Listen("tcp", srv.server.Addr)
		if err != nil {
			return
		}
	}

	srv.listener = ln

	go func() {
		if err := srv.server.Serve(srv.listener); err != http.ErrServerClosed {
			log.Printf("http Serve error: %v\n", err)
		}
	}()

	srv.handleSignal()

	return
}

func (srv *Server) handleSignal() {
	signal.Notify(srv.signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	for s := range srv.signalChan {
		switch s {
		case syscall.SIGINT, syscall.SIGTERM:
			ctx, cancel := context.WithTimeout(context.Background(), srv.shutdownTime)
			srv.server.Shutdown(ctx)
			cancel()
		case syscall.SIGHUP:
			err := srv.fork()
			if err != nil {
				log.Printf("start new process failed, please retry: %v\n", err)
				continue
			}

			time.Sleep(srv.startTime)

			ctx, cancel := context.WithTimeout(context.Background(), srv.shutdownTime)
			srv.server.Shutdown(ctx)
			cancel()
		}
		break
	}
}

func (srv *Server) fork() (err error) {
	log.Println("grace restart...")

	var env []string
	for _, v := range os.Environ() {
		if v != GRACEFUL_ENV {
			env = append(env, v)
		}
	}
	env = append(env, GRACEFUL_ENV)

	cmd := exec.Command(os.Args[0], os.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = env
	file, _ := srv.listener.(*net.TCPListener).File()
	cmd.ExtraFiles = []*os.File{file}

	err = cmd.Start()
	if err != nil {
		return
	}

	return
}

func ListenAndServe(addr string, handler http.Handler) error {
	shutdownTime := time.Millisecond * 800
	startTime := time.Millisecond * 200
	timeout := time.Minute * 10
	return NewServer(addr, handler, shutdownTime, startTime, timeout).ListenAndServe()
}

func ListenAndServeWithTimeout(addr string, handler http.Handler, shutdownTime, startTime, timeout time.Duration) error {
	return NewServer(addr, handler, shutdownTime, startTime, timeout).ListenAndServe()
}
