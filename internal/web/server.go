package web

import (
	_ "embed"
	"encoding/json"
	"net/http"

	"github.com/Cycloctane/dnspass/internal/records"
)

//go:embed index.html
var indexHTML []byte

type recordjson struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
	TTL   uint32 `json:"ttl"`
}

func NewServer(username, password string) (*http.Server, error) {
	router := http.NewServeMux()
	router.HandleFunc("/api/records", recordsHandler)
	router.HandleFunc("/index.html", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write(indexHTML)
	})
	router.Handle("/", http.RedirectHandler("/index.html", http.StatusMovedPermanently))
	if username != "" || password != "" {
		return &http.Server{Handler: NewBasicAuth(router, username, password)}, nil
	} else {
		return &http.Server{Handler: router}, nil
	}
}

func recordsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		items := records.List()
		out := make([]recordjson, len(items))
		for i, item := range items {
			out[i] = recordjson{Name: item.Name, Type: string(item.Type), Value: item.Value, TTL: item.TTL}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(out)
	case http.MethodPut:
		var items []recordjson
		if err := json.NewDecoder(r.Body).Decode(&items); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		list := make([]records.Record, 0, len(items))
		for _, item := range items {
			list = append(list, records.Record{
				Name:  item.Name,
				Type:  records.DNSType(item.Type),
				Value: item.Value,
				TTL:   item.TTL,
			})
		}
		records.SetAll(list)
		w.WriteHeader(http.StatusNoContent)
	case http.MethodPost:
		var item recordjson
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if !records.Add(records.Record{Name: item.Name, Type: records.DNSType(item.Type), Value: item.Value, TTL: item.TTL}) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	case http.MethodDelete:
		var item recordjson
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if !records.Delete(item.Name, records.DNSType(item.Type), item.Value) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
