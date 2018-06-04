/*
 * Revision History:
 *     Initial: 2018/06/01        Wang RiYu
 */

package geoip2

import (
	"log"
	"sync"
	"testing"
	"time"

	"github.com/robfig/cron"
)

func TestClient_Init(t *testing.T) {
	var client = DefaultClient

	err := client.Init()
	if err != nil {
		t.Fatal(err)
	}
}

func TestClient_Lookup(t *testing.T) {
	var client = DefaultClient

	wg := sync.WaitGroup{}
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			res, err := client.Lookup("2001:4860:4860::8888") // ipv6
			if err != nil {
				t.Fatal(err)
			}

			t.Logf("ipv6-%d: %+v\n", i, *res)
		}(i)
	}

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			res, err := client.Lookup("106.117.249.107") // ipv4
			if err != nil {
				t.Fatal(err)
			}

			t.Logf("ipv4-%d: %+v\n", i, *res)
		}(i)
	}
	wg.Wait()

	_, err := client.Lookup("127.0.0.1")
	if err.Error() != "not found" {
		t.Error(err)
	}

	_, err = client.Lookup("1234.5")
	if err.Error() != "unvalid ip address" {
		t.Error(err)
	}

	c := &Client{
		DBLocationDir: "maxminddb",
		Timeout:       time.Second * 5,
		MaxConnect:    0x0,
	}
	c.AsnDB = client.AsnDB
	c.CityDB = client.CityDB

	_, err = c.Lookup("106.117.249.107")
	if err.Error() != "no more lookup operation for now, wait a minute" {
		t.Error(err)
	}

	c.Timeout = 0
	c.MaxConnect = 0x5

	_, err = c.Lookup("106.117.249.107")
	if err.Error() != "lookup timeout" {
		t.Error(err)
	}
}

func TestClient_DBMeta(t *testing.T) {
	var client = DefaultClient

	meta, err := client.DBMeta()
	if err != nil {
		t.Fatal(err)
	}

	for _, db := range meta {
		log.Printf("Version: %s\n", db.Version)
		log.Printf("IPVersion: %s\n", db.IPVersion)
		log.Printf("DatabaseType: %s\n", db.DatabaseType)
		log.Printf("BuildEpoch: %s\n\n", db.BuildEpoch)
	}
}

func TestClient_UpdateDB(t *testing.T) {
	cron := cron.New()

	cron.Start()
	defer cron.Stop()

	var client = DefaultClient
	// every 2 minutes execute the func
	cron.AddFunc("0 */2 * * * *", func() {
		client.UpdateDB()
	})

	// at 4:30 on the first Wednesday of each month execute the func
	// the first Tuesday is the "GeoLite2-City.mmdb" update time
	cron.AddFunc("0 30 4 * * 3", func() {
		client.UpdateDB()
	})

	select {}
}
