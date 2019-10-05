package main

import (
	"context"
	"log"
	"os"
	"strconv"

	"gocloud.dev/docstore"
	_ "gocloud.dev/docstore/memdocstore"

	"github.com/btmura/ponzi2/internal/server"
	"github.com/btmura/ponzi2/internal/stock/iex"
)

func main() {
	ctx := context.Background()

	port := os.Getenv("PORT")
	if port == "" {
		port = "1337"
	}

	intPort, err := strconv.Atoi(port)
	if err != nil {
		log.Fatal(err)
	}

	coll, err := docstore.OpenCollection(ctx, "mem://collection/Key")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := coll.Close(); err != nil {
			log.Print(err)
		}
	}()

	d := server.NewDocChartCache(coll)
	c := iex.NewClient(d, false)
	s := server.New(c)
	log.Fatal(s.Run(intPort))
}
