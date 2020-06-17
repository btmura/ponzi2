package server

import (
	"encoding/gob"
	"fmt"
	"net/http"

	"gocloud.dev/server"

	"github.com/btmura/ponzi2/internal/errors"
	"github.com/btmura/ponzi2/internal/logger"
	"github.com/btmura/ponzi2/internal/stock/iex"
)

// Server processes requests.
type Server struct {
	client *iex.Client
}

// New returns a new Server.
func New(client *iex.Client) *Server {
	return &Server{client}
}

// Run runs the server. Should be called from main.
func (s *Server) Run(port int) error {
	srv := server.New(http.DefaultServeMux, nil)
	http.HandleFunc("/chart", s.chartHandler)
	http.HandleFunc("/quote", s.quoteHandler)

	addr := fmt.Sprintf(":%d", port)
	logger.Infof("listening on %s", addr)
	return srv.ListenAndServe(addr)
}

func (s *Server) quoteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	iexReq := &iex.GetQuotesRequest{}
	dec := gob.NewDecoder(r.Body)
	if err := dec.Decode(iexReq); err != nil {
		logAndWriteError(w, http.StatusBadRequest, errors.Errorf("decoding quote request failed: %v", err))
		return
	}

	iexResp, err := s.client.GetQuotes(ctx, iexReq)
	if err != nil {
		logAndWriteError(w, http.StatusBadRequest, errors.Errorf("getting quotes failed: %v", err))
		return
	}

	enc := gob.NewEncoder(w)
	if err := enc.Encode(iexResp); err != nil {
		logAndWriteError(w, http.StatusInternalServerError, errors.Errorf("encoding quote response failed: %v", err))
		return
	}
}

func (s *Server) chartHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	iexReq := &iex.GetChartsRequest{}
	dec := gob.NewDecoder(r.Body)
	if err := dec.Decode(iexReq); err != nil {
		logAndWriteError(w, http.StatusBadRequest, errors.Errorf("decoding chart request failed: %v", err))
		return
	}

	iexResp, err := s.client.GetCharts(ctx, iexReq)
	if err != nil {
		logAndWriteError(w, http.StatusBadRequest, errors.Errorf("getting charts failed: %v", err))
		return
	}

	enc := gob.NewEncoder(w)
	if err := enc.Encode(iexResp); err != nil {
		logAndWriteError(w, http.StatusInternalServerError, errors.Errorf("encoding chart response failed: %v", err))
		return
	}
}

func logAndWriteError(w http.ResponseWriter, statusCode int, err error) {
	logger.Error(err)
	w.WriteHeader(statusCode)
	fmt.Fprint(w, err)
}
