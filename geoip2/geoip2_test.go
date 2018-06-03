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

			t.Logf("%d: %+v\n", i, res)
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

			t.Logf("%d: %+v\n", i, res)
		}(i)
	}
	wg.Wait()
}

func TestClient_DBMeta(t *testing.T) {
	var client = DefaultClient

	meta, err := client.DBMeta()
	if err != nil {
		t.Fatal(err)
	}

	for _, db := range meta {
		t.Logf("Version: %s\n", db.Version)
		t.Logf("IPVersion: %s\n", db.IPVersion)
		t.Logf("DatabaseType: %s\n", db.DatabaseType)
		t.Logf("BuildEpoch: %s\n\n", db.BuildEpoch)
	}
}

func TestClient_UpdateDB(t *testing.T) {
	cron := cron.New()

	cron.Start()
	defer cron.Stop()

	var client = DefaultClient
	// every 5 minutes execute the func
	cron.AddFunc("0 */5 * * * *", func() {
		log.Println(time.Now().UTC().String())
		client.UpdateDB()
	})

	// at 4:30 on the first Wednesday of each month execute the func
	// the first Tuesday is the "GeoLite2-City.mmdb" update time
	cron.AddFunc("0 30 4 * * 3", func() {
		log.Println(time.Now().UTC().String())
		client.UpdateDB()
	})

	select {}
}
