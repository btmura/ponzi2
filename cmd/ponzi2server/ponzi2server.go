package main

import (
	"context"
	"flag"
	"os"
	"strconv"

	"gocloud.dev/docstore"
	_ "gocloud.dev/docstore/memdocstore"
	_ "gocloud.dev/docstore/mongodocstore"

	"github.com/btmura/ponzi2/internal/logger"
	"github.com/btmura/ponzi2/internal/server"
	"github.com/btmura/ponzi2/internal/stock/iex"
)

var (
	port        = flag.Int("port", 1337, "Port on which the server will listen for HTTP requests.")
	docstoreURL = flag.String("docstore_url", "mem://collection/Key", "URL of the Docstore backing the IEX chart cache.")
)

func main() {
	flag.Parse()

	ctx := context.Background()

	if envPort := os.Getenv("PORT"); envPort != "" {
		intPort, err := strconv.Atoi(envPort)
		if err != nil {
			logger.Fatalf("converting int port failed: %v", err)
		}
		*port = intPort
		logger.Infof("overriding port flag with env value: %d", *port)
	}

	if envDocstoreURL := os.Getenv("DOCSTORE_URL"); envDocstoreURL != "" {
		*docstoreURL = envDocstoreURL
		logger.Infof("overriding docstore_url flag with env value: %s", *docstoreURL)
	}

	if *port == 0 {
		logger.Fatal("port must not be non-zero")
	}

	if *docstoreURL == "" {
		logger.Fatal("docstore URL must not be empty")
	}

	coll, err := docstore.OpenCollection(ctx, *docstoreURL)
	if err != nil {
		logger.Fatalf("opening collection failed: %v", err)
	}
	defer func() {
		if err := coll.Close(); err != nil {
			logger.Infof("closing collection failed: %v", err)
		}
	}()

	d := server.NewDocChartCache(coll)
	c := iex.NewClient(d, false)
	s := server.New(c)
	logger.Fatal(s.Run(*port))
}
