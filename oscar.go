package main

import (
	"flag"
	"log"

	"github.com/dgraph-io/badger/v4"
)

var (
	dnsBind  = "0.0.0.0"
	dnsPort  = 5353
	httpBind = "0.0.0.0"
	httpPort = 12540

	zoneDB = "dns.zone"
)

func init() {
	flag.StringVar(&dnsBind, "dnsBind", dnsBind, "bind Oscar to a network interface")
	flag.IntVar(&dnsPort, "dnsPort", dnsPort, "Oscar will listen at this port")
	flag.StringVar(&zoneDB, "zone", zoneDB, "zone database")
	flag.StringVar(&httpBind, "httpBind", httpBind, "bind Oscar's HTTP API to a network interface")
	flag.IntVar(&httpPort, "httpPort", httpPort, "Oscar's HTTP API will listen at this port")

	flag.Parse()
}

func main() {
	db, err := badger.Open(badger.DefaultOptions(zoneDB))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	dnsServer := NewDnsServer(dnsBind, dnsPort, db)
	httpServer := NewHTTPServer(httpBind, httpPort, db)

	go httpServer.ListenAndServe()

	dnsServer.ListenAndServe()
}
