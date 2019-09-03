package main

import (
	"log"
	"os"
	"strconv"

	"github.com/btmura/ponzi2/internal/server"
	"github.com/btmura/ponzi2/internal/stock/iex"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "1337"
	}

	intPort, err := strconv.Atoi(port)
	if err != nil {
		log.Fatal(err)
	}

	c := iex.NewClient(new(iex.NoOpChartCache), false)
	s := server.NewServer(intPort, c)
	log.Fatal(s.Run())
}
