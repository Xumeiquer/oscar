package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/dgraph-io/badger/v4"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/unrolled/render"
)

type httpServer struct {
	bind   string
	port   int
	r      *chi.Mux
	zoneDB *badger.DB
	render *render.Render
}

func NewHTTPServer(bind string, port int, zoneDB *badger.DB) *httpServer {
	s := &httpServer{
		r:      chi.NewRouter(),
		bind:   bind,
		port:   port,
		zoneDB: zoneDB,
		render: new(render.Render),
	}

	s.r.Use(middleware.Logger)

	s.r.Get("/read/{domain}/{type}", s.Read)
	s.r.Post("/create/{domain}/{type}/{value}/{ttl}", s.Create)
	s.r.Put("/update/{domain}/{type}/{value}/{ttl}", s.Update)
	s.r.Delete("/delete/{domain}/{type}", s.Delete)

	return s
}

func (hs *httpServer) ListenAndServe() {
	log.Println("Starting HTTP server")
	log.Printf("Listening at %s:%d\n", hs.bind, hs.port)
	defer log.Println("Stopping HTTP server")
	http.ListenAndServe(fmt.Sprintf("%s:%d", hs.bind, hs.port), hs.r)
}

func (hs *httpServer) Read(w http.ResponseWriter, req *http.Request) {
	domainParam := chi.URLParam(req, "domain")
	typeParam := chi.URLParam(req, "type")

	var buff []byte

	err := hs.zoneDB.View(func(txn *badger.Txn) error {
		var query string
		if strings.HasSuffix(domainParam, ".") {
			query = fmt.Sprintf("%s|%s", domainParam, typeParam)
		} else {
			query = fmt.Sprintf("%s.|%s", domainParam, typeParam)
		}

		item, err := txn.Get([]byte(query))
		if err != nil {
			return err
		}

		buff, err = item.ValueCopy(nil)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		hs.render.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": "resource does not exist",
		})
		return
	}

	response := string(buff)

	hs.render.JSON(w, http.StatusOK, map[string]interface{}{
		"name":  domainParam,
		"type":  typeParam,
		"value": strings.Split(response, "|")[0],
		"ttl":   strings.Split(response, "|")[1],
	})
}

func (hs *httpServer) Create(w http.ResponseWriter, req *http.Request) {
	domainParam := chi.URLParam(req, "domain")
	typeParam := chi.URLParam(req, "type")
	valueParam := chi.URLParam(req, "value")
	ttlParam := chi.URLParam(req, "ttl")

	err := hs.zoneDB.Update(func(txn *badger.Txn) error {
		var key string
		if strings.HasSuffix(domainParam, ".") {
			key = fmt.Sprintf("%s.|%s", domainParam, typeParam)
		} else {
			key = fmt.Sprintf("%s|%s", domainParam, typeParam)
		}

		e := badger.NewEntry([]byte(key), []byte(fmt.Sprintf("%s|%s", valueParam, ttlParam)))
		err := txn.SetEntry(e)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		hs.render.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": "unable to save it",
		})
		return
	}

	hs.render.JSON(w, http.StatusOK, map[string]interface{}{
		"name":  domainParam,
		"type":  typeParam,
		"value": valueParam,
		"ttl":   ttlParam,
	})
}

func (hs *httpServer) Update(w http.ResponseWriter, req *http.Request) {
	domainParam := chi.URLParam(req, "domain")
	typeParam := chi.URLParam(req, "type")
	valueParam := chi.URLParam(req, "value")
	ttlParam := chi.URLParam(req, "ttl")

	err := hs.zoneDB.Update(func(txn *badger.Txn) error {
		var key string
		if strings.HasSuffix(domainParam, ".") {
			key = fmt.Sprintf("%s.|%s", domainParam, typeParam)
		} else {
			key = fmt.Sprintf("%s|%s", domainParam, typeParam)
		}

		err := txn.Set([]byte(key), []byte(fmt.Sprintf("%s|%s", valueParam, ttlParam)))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		hs.render.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": "unable to update it",
		})
		return
	}

	hs.render.JSON(w, http.StatusOK, map[string]interface{}{
		"name":  domainParam,
		"type":  typeParam,
		"value": valueParam,
		"ttl":   ttlParam,
	})
}

func (hs *httpServer) Delete(w http.ResponseWriter, req *http.Request) {
	domainParam := chi.URLParam(req, "domain")
	typeParam := chi.URLParam(req, "type")

	err := hs.zoneDB.Update(func(txn *badger.Txn) error {
		err := txn.Delete([]byte(fmt.Sprintf("%s|%s", domainParam, typeParam)))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		hs.render.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": "unable to delete it",
		})
		return
	}

	hs.render.JSON(w, http.StatusOK, map[string]interface{}{
		"deleted": true,
	})
}
