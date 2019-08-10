package iexremote

import (
	"encoding/gob"
	"fmt"
	"net/http"

	"github.com/btmura/ponzi2/internal/stock/iex"
)

// Server processes requests.
type Server struct {
	port   int
	client *iex.Client
}

// NewServer returns a new Server.
func NewServer(port int, client *iex.Client) *Server {
	return &Server{port, client}
}

// Run runs the server. Should be called from main.
func (s *Server) Run() error {
	http.HandleFunc("/chart", s.chartHandler)
	http.HandleFunc("/quote", s.quoteHandler)
	return http.ListenAndServe(fmt.Sprintf(":%d", s.port), nil)
}

func (s *Server) quoteHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("quote")

	ctx := r.Context()

	iexReq := &iex.GetQuotesRequest{}

	dec := gob.NewDecoder(r.Body)
	if err := dec.Decode(iexReq); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	iexResp, err := s.client.GetQuotes(ctx, iexReq)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	enc := gob.NewEncoder(w)
	if err := enc.Encode(iexResp); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Server) chartHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("chart")

	ctx := r.Context()

	iexReq := &iex.GetChartsRequest{}

	dec := gob.NewDecoder(r.Body)
	if err := dec.Decode(iexReq); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	iexResp, err := s.client.GetCharts(ctx, iexReq)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	enc := gob.NewEncoder(w)
	if err := enc.Encode(iexResp); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
