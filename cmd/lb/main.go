package main

import (
	"context"
	"errors"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/gorilla/handlers"
	"github.com/protomem/simplelb/pkg/backendpool"
	"github.com/protomem/simplelb/pkg/httpserver"
	"github.com/protomem/simplelb/pkg/loadbalancer"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

var (
	ServerPort int
	Addrs      string
)

func init() {
	flag.IntVar(&ServerPort, "port", 4000, "server port")
	flag.StringVar(&Addrs, "addrs", "", "addres server list")
}

func main() {

	logrus.Info("Server Configure ....")
	defer logrus.Info("Server Stop.")

	flag.Parse()
	if Addrs == "" {
		logrus.Fatal(errors.New("address list is empty"))
	}

	addrList := strings.Split(Addrs, ",")

	logrus.Info("Server Init ...")
	mainCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	// Backend Pool Init
	bp, err := backendpool.New(addrList...)
	if err != nil {
		logrus.Fatal(err)
	}

	//Load Balancer Init
	var lb http.Handler

	lbCore := loadbalancer.New(bp)
	lbRecovery := handlers.RecoveryHandler()(lbCore)
	lbLogger := handlers.LoggingHandler(os.Stdout, lbRecovery)

	lb = lbLogger

	// Server Init
	server := httpserver.New(mainCtx, lb, ServerPort)

	g, gCtx := errgroup.WithContext(mainCtx)

	g.Go(func() error {
		return server.Run()
	})

	g.Go(func() error {
		<-gCtx.Done()
		logrus.Info("Server Shutting Down ...")
		return server.ShutDown(context.Background())
	})

	logrus.Infof("Server Start on port: | %s | ...", strconv.Itoa(ServerPort))
	if err := g.Wait(); err != nil {
		logrus.Error(err)
	}

}
