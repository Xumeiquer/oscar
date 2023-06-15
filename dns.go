package main

import (
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/dgraph-io/badger/v4"
	"github.com/miekg/dns"
)

type dnsServer struct {
	bind     string
	port     int
	zoneDB   *badger.DB
	srv      *dns.Server
	handlers dnsHandlers
}

type dnsHandlers struct {
	zoneDB *badger.DB
}

func NewDnsServer(bind string, port int, zoneDB *badger.DB) *dnsServer {
	handler := &dnsServer{
		zoneDB: zoneDB,
		bind:   bind,
		port:   port,
	}

	handler.srv = &dns.Server{
		Addr: fmt.Sprintf("%s:%d", bind, port),
		Net:  "udp",
	}

	handler.srv.Handler = &dnsHandlers{
		zoneDB: zoneDB,
	}

	return handler
}

func (ds *dnsServer) ListenAndServe() {
	log.Println("Starting DNS server")
	defer log.Println("Stopping DNS server")
	log.Printf("Listening at %s:%d\n", ds.bind, ds.port)
	err := ds.srv.ListenAndServe()
	if err != nil {
		log.Fatalf("Failed to set udp listener %s\n", err.Error())
	}
}

func (dh *dnsHandlers) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	msg := dns.Msg{}
	msg.SetReply(r)

	switch r.Question[0].Qtype {
	case dns.TypeA:
		msg.Authoritative = true
		domain := msg.Question[0].Name

		address, ok := dh.lookup(domain, "A")
		if ok {
			msg.Answer = append(msg.Answer, &dns.A{
				Hdr: dns.RR_Header{Name: domain, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
				A:   net.ParseIP(address),
			})
		}
	}
	w.WriteMsg(&msg)
}

func (dh *dnsHandlers) lookup(domain string, queryType string) (string, bool) {
	var buff []byte
	err := dh.zoneDB.View(func(txn *badger.Txn) error {
		var query string
		if strings.HasSuffix(domain, ".") {
			query = fmt.Sprintf("%s|%s", domain, queryType)
		} else {
			query = fmt.Sprintf("%s.|%s", domain, queryType)
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
		return "", false
	}

	response := string(buff)

	return strings.Split(response, "|")[0], true
}
