/*
 * This product includes GeoLite2 data created by MaxMind, available from http://www.maxmind.com.
 *
 * Revision History:
 *     Initial: 2018/06/01        Wang RiYu
 */

package geoip2

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/oschwald/maxminddb-golang"
	"golang.org/x/sync/errgroup"
)

// CityDB ...
const CityDB = "http://geolite.maxmind.com/download/geoip/database/GeoLite2-City.tar.gz"

// AsnDB ...
const AsnDB = "http://geolite.maxmind.com/download/geoip/database/GeoLite2-ASN.tar.gz"

// DBLocationDir is used to store mmdb files
const DBLocationDir = "maxminddb"

type (
	// Client ...
	Client struct {
		UserAgent  string
		mux        sync.Mutex
		MaxConnect uint
		Timeout    time.Duration
	}
	// Result ...
	Result struct {
		Continent struct {
			Code  string `maxminddb:"code" json:"code"`
			Names struct {
				ZhCN string `maxminddb:"zh-CN" json:"zh-cn"`
				En   string `maxminddb:"en" json:"en"`
			} `maxminddb:"names" json:"names"`
		} `maxminddb:"continent" json:"continent"`
		Country struct {
			ISOCode string `maxminddb:"iso_code" json:"iso_code"`
			Names   struct {
				ZhCN string `maxminddb:"zh-CN" json:"zh-cn"`
				En   string `maxminddb:"en" json:"en"`
			} `maxminddb:"names" json:"names"`
		} `maxminddb:"country" json:"country"`
		City struct {
			Names struct {
				ZhCN string `maxminddb:"zh-CN" json:"zh-cn"`
				En   string `maxminddb:"en" json:"en"`
			} `maxminddb:"names" json:"names"`
		} `maxminddb:"city" json:"city"`
		Subdivisions []struct {
			ISOCode string `maxminddb:"iso_code" json:"iso_code"`
			Names   struct {
				ZhCN string `maxminddb:"zh-CN" json:"zh-cn"`
				En   string `maxminddb:"en" json:"en"`
			} `maxminddb:"names" json:"names"`
		} `maxminddb:"subdivisions" json:"subdivisions"`
		Location struct {
			TimeZone       string  `maxminddb:"time_zone" json:"time_zone"`
			Latitude       float32 `maxminddb:"latitude" json:"latitude"`
			Longitude      float32 `maxminddb:"longitude" json:"longitude"`
			AccuracyRadius int     `maxminddb:"accuracy_radius" json:"accuracy_radius"`
		} `maxminddb:"location" json:"location"`
		RegisteredCountry struct {
			ISOCode string `maxminddb:"iso_code" json:"iso_code"`
			Names   struct {
				ZhCN string `maxminddb:"zh-CN" json:"zh-cn"`
				En   string `maxminddb:"en" json:"en"`
			} `maxminddb:"names" json:"names"`
		} `maxminddb:"registered_country" json:"registered_country"`
		Organization string `maxminddb:"autonomous_system_organization" json:"organization"`
	}
)

// DefaultClient ...
var DefaultClient = &Client{
	Timeout:    time.Second * 15,
	MaxConnect: 0x64,
}

// Init the database
func (c *Client) Init() error {
	var (
		asnDBLocation  = filepath.Join(DBLocationDir, "GeoLite2-ASN.mmdb")
		cityDBLocation = filepath.Join(DBLocationDir, "GeoLite2-City.mmdb")
	)

	g := errgroup.Group{}

	g.Go(func() error {
		if !pathExist(asnDBLocation) {
			err := downloadMMDB(AsnDB)
			if err != nil {
				return err
			}
		}

		asnDB, err := maxminddb.Open(asnDBLocation)
		if err != nil {
			os.Remove(asnDBLocation)
			return err
		}
		defer asnDB.Close()

		err = asnDB.Verify()
		if err != nil {
			os.Remove(asnDBLocation)
			return err
		}

		return nil
	})

	g.Go(func() error {
		if !pathExist(cityDBLocation) {
			err := downloadMMDB(CityDB)
			if err != nil {
				return err
			}
		}

		cityDB, err := maxminddb.Open(cityDBLocation)
		if err != nil {
			os.Remove(cityDBLocation)
			return err
		}
		defer cityDB.Close()

		err = cityDB.Verify()
		if err != nil {
			os.Remove(cityDBLocation)
			return err
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}

//func (c *Client) updateDB(){
//	c.mux.Lock()
//	defer c.mux.Unlock()
//}

// Lookup for ip geo information
func (c *Client) Lookup(ipStr string) (*Result, error) {
	if c.MaxConnect <= 0 {
		return nil, errors.New("no more lookup operation")
	}
	c.MaxConnect--

	var (
		ip     = net.ParseIP(ipStr)
		result = &Result{}
	)

	if ip == nil {
		return nil, errors.New("unvalid ip address")
	}

	asnDB, err := maxminddb.Open(filepath.Join(DBLocationDir, "GeoLite2-ASN.mmdb"))
	if err != nil {
		return result, err
	}
	defer asnDB.Close()

	err = asnDB.Lookup(ip, &result)
	if err != nil {
		return result, err
	}

	cityDB, err := maxminddb.Open(filepath.Join(DBLocationDir, "GeoLite2-City.mmdb"))
	if err != nil {
		return result, err
	}
	defer cityDB.Close()

	err = cityDB.Lookup(ip, &result)
	if err != nil {
		return result, err
	}

	c.MaxConnect++

	return result, nil
}

func downloadMMDB(dbLink string) error {
	_, dbFile := filepath.Split(dbLink)

	log.Printf("Downloading %s ...\n", dbFile)
	res, err := http.Get(dbLink)
	if err != nil {
		return err
	}

	f, err := os.Create(dbFile)
	if err != nil {
		return err
	}

	if _, err = io.Copy(f, res.Body); err != nil {
		os.Remove(dbFile)
		return err
	}
	log.Printf("Download %s complete\n", dbFile)

	err = deCompressAndMove(dbFile, strings.Join([]string{DBLocationDir, "download/"}, "/"))
	if err != nil {
		return err
	}
	os.RemoveAll(strings.Join([]string{DBLocationDir, "download/"}, "/"))
	os.Remove(dbFile)

	return nil
}

func deCompressAndMove(tarFile, dest string) error {
	srcFile, err := os.Open(tarFile)
	defer srcFile.Close()
	if err != nil {
		return err
	}

	gr, err := gzip.NewReader(srcFile)
	defer gr.Close()
	if err != nil {
		return err
	}

	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}

		filename := dest + hdr.Name
		file, err := createFile(filename)
		if err != nil {
			return err
		}

		if file != nil {
			_, err = io.Copy(file, tr)
			if err != nil {
				return err
			}
		}

		_, dbName := filepath.Split(hdr.Name)
		if dbName == "GeoLite2-City.mmdb" || dbName == "GeoLite2-ASN.mmdb" {
			err = os.Rename(filename, filepath.Join(DBLocationDir, dbName))
		}
	}

	return nil
}

func pathExist(path string) bool {
	_, err := os.Stat(path)

	if err != nil && os.IsNotExist(err) {
		return false
	}

	return true
}

func createFile(name string) (*os.File, error) {
	dir, file := filepath.Split(name)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return nil, err
	}

	if file != "" {
		return os.Create(name)
	}

	return nil, nil
}
