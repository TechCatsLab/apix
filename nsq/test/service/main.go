package main

import (
	"log"

	"github.com/TechCatsLab/apix/nsq"
)

func main() {
	lookupd, err := nsq.StartNsqlookupd(nil)
	if err != nil {
		log.Printf("%v", err)
	}

	nsqdOpts := nsq.NewNsqdOptions()
	nsqdOpts.NSQLookupdTCPAddresses = []string{lookupd.RealTCPAddr().String()}

	_, err = nsq.StartNsqd(nsqdOpts)
	if err != nil {
		log.Printf("%v", err)
	}

	nsqAdminOpts := nsq.NewNsqadminOptions()
	nsqAdminOpts.NSQLookupdHTTPAddresses = []string{lookupd.RealHTTPAddr().String()}

	_, err = nsq.StartNsqAdmin(nsqAdminOpts)
	if err != nil {
		log.Printf("%v", err)
	}

	select {}
}
