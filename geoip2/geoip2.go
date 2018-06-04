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
	"fmt"
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

// CityDB download link
const CityDB = "http://geolite.maxmind.com/download/geoip/database/GeoLite2-City.tar.gz"

// AsnDB download link
const AsnDB = "http://geolite.maxmind.com/download/geoip/database/GeoLite2-ASN.tar.gz"

type (
	// Client ...
	Client struct {
		DBLocationDir string        // DBLocationDir is used to store mmdb files
		MaxConnect    int           // max synchronously lookup
		Timeout       time.Duration // lookup duration
		mux           sync.RWMutex  // maintain db
		AsnDB         *maxminddb.Reader
		CityDB        *maxminddb.Reader
	}
	// Result of lookup
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
	// DBMeta metadata
	DBMeta struct {
		Version      string    `json:"version"`
		IPVersion    string    `json:"ipVersion"`
		DatabaseType string    `json:"type"`
		BuildEpoch   time.Time `json:"buildEpoch"`
	}
)

// DefaultClient ...
var DefaultClient = &Client{
	DBLocationDir: "maxminddb",
	Timeout:       time.Second * 15,
	MaxConnect:    0x64,
}

// Init and verify the database, if none exists, download it
func (c *Client) Init() error {
	var (
		asnDBLocation  = filepath.Join(c.DBLocationDir, "GeoLite2-ASN.mmdb")
		cityDBLocation = filepath.Join(c.DBLocationDir, "GeoLite2-City.mmdb")
	)

	defer func() {
		err := os.RemoveAll(strings.Join([]string{c.DBLocationDir, "download/"}, "/"))
		if err != nil {
			log.Println(err)
		}
	}()

	g := errgroup.Group{}
	g.Go(func() error {
		if !pathExist(asnDBLocation) {
			path, err := c.downloadMMDB(AsnDB)
			if err != nil {
				return err
			}
			err = os.Rename(path, asnDBLocation)
			if err != nil {
				return err
			}
		}

		asnDB, err := maxminddb.Open(asnDBLocation)
		if err != nil {
			os.Remove(asnDBLocation)
			return err
		}

		err = asnDB.Verify()
		if err != nil {
			os.Remove(asnDBLocation)
			return err
		}

		c.AsnDB = asnDB

		return nil
	})

	g.Go(func() error {
		if !pathExist(cityDBLocation) {
			path, err := c.downloadMMDB(CityDB)
			if err != nil {
				return err
			}
			err = os.Rename(path, cityDBLocation)
			if err != nil {
				return err
			}
		}

		cityDB, err := maxminddb.Open(cityDBLocation)
		if err != nil {
			os.Remove(cityDBLocation)
			return err
		}

		err = cityDB.Verify()
		if err != nil {
			os.Remove(cityDBLocation)
			return err
		}

		c.CityDB = cityDB

		return nil
	})

	if err := g.Wait(); err != nil {
		log.Println("initialization faild with err:", err)
		return err
	}

	log.Println("initialization complete")
	return nil
}

// UpdateDB when remote database update, used with cron
func (c *Client) UpdateDB() {
	var (
		asnDBLocation  = filepath.Join(c.DBLocationDir, "GeoLite2-ASN.mmdb")
		cityDBLocation = filepath.Join(c.DBLocationDir, "GeoLite2-City.mmdb")
	)

	log.Println("Update database at", time.Now().UTC().String())

	defer func() {
		err := os.RemoveAll(strings.Join([]string{c.DBLocationDir, "download/"}, "/"))
		if err != nil {
			log.Println(err)
		}
	}()

	g := errgroup.Group{}
	g.Go(func() error {
		path, err := c.downloadMMDB(AsnDB)
		if err != nil {
			return err
		}

		db, err := maxminddb.Open(path)
		if err != nil {
			return err
		}

		err = db.Verify()
		if err != nil {
			return err
		}

		err = os.Rename(path, asnDBLocation)
		if err != nil {
			return err
		}

		return nil
	})

	g.Go(func() error {
		path, err := c.downloadMMDB(CityDB)
		if err != nil {
			return err
		}

		db, err := maxminddb.Open(path)
		if err != nil {
			return err
		}

		err = db.Verify()
		if err != nil {
			return err
		}

		err = os.Rename(path, cityDBLocation)
		if err != nil {
			return err
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		log.Println("Update database failed with err:", err)
		c.UpdateDB()
	} else {
		asnDB, err := maxminddb.Open(asnDBLocation)
		if err != nil {
			log.Println(err)
			c.UpdateDB()
		}
		cityDB, err := maxminddb.Open(cityDBLocation)
		if err != nil {
			log.Println(err)
			c.UpdateDB()
		}
		c.mux.Lock()
		c.AsnDB = asnDB
		c.CityDB = cityDB
		c.mux.Unlock()

		log.Println("Update database complete")
	}
}

// DBMeta return database metadata
func (c *Client) DBMeta() ([]DBMeta, error) {
	c.mux.RLock()
	defer c.mux.RUnlock()

	if c.AsnDB == nil || c.CityDB == nil {
		return nil, errors.New("no database")
	}

	var asnMeta = &DBMeta{}
	asnMeta.Version = fmt.Sprintf("%d.%d", c.AsnDB.Metadata.BinaryFormatMajorVersion, c.AsnDB.Metadata.BinaryFormatMinorVersion)
	asnMeta.IPVersion = fmt.Sprintf("%d", c.AsnDB.Metadata.IPVersion)
	asnMeta.DatabaseType = c.AsnDB.Metadata.DatabaseType
	asnMeta.BuildEpoch = time.Unix(int64(c.AsnDB.Metadata.BuildEpoch), 0).UTC()

	var cityMeta = &DBMeta{}
	cityMeta.Version = fmt.Sprintf("%d.%d", c.CityDB.Metadata.BinaryFormatMajorVersion, c.CityDB.Metadata.BinaryFormatMinorVersion)
	cityMeta.IPVersion = fmt.Sprintf("%d", c.CityDB.Metadata.IPVersion)
	cityMeta.DatabaseType = c.CityDB.Metadata.DatabaseType
	cityMeta.BuildEpoch = time.Unix(int64(c.CityDB.Metadata.BuildEpoch), 0).UTC()

	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()

	return []DBMeta{*asnMeta, *cityMeta}, nil
}

// Lookup for ip geo information
func (c *Client) Lookup(ipStr string) (*Result, error) {
	c.MaxConnect--
	c.mux.RLock()
	defer func() {
		c.MaxConnect++
		c.mux.RUnlock()
	}()

	if c.MaxConnect < 0 {
		return nil, errors.New("no more lookup operation for now, wait a minute")
	}

	var (
		ip     = net.ParseIP(ipStr)
		result = &Result{}
	)

	if ip == nil {
		return nil, errors.New("unvalid ip address")
	}

	if _, err := c.DBMeta(); err != nil {
		return nil, err
	}

	eg := &errgroup.Group{}

	eg.Go(func() error {
		err := c.AsnDB.Lookup(ip, &result)
		if err != nil {
			return err
		}

		err = c.CityDB.Lookup(ip, &result)
		if err != nil {
			return err
		}

		return nil
	})

	select {
	case <-time.After(c.Timeout):
		return result, errors.New("lookup timeout")
	case err := <-wait(eg):
		if err != nil {
			return result, err
		}
		if result.Subdivisions == nil {
			return result, errors.New("not found")
		}
		return result, nil
	}
}

// Close database
func (c *Client) Close() error {
	if c.AsnDB == nil || c.CityDB == nil {
		return errors.New("no database")
	}

	err := c.AsnDB.Close()
	if err != nil {
		return err
	}
	err = c.CityDB.Close()
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) downloadMMDB(dbLink string) (string, error) {
	_, dbFile := filepath.Split(dbLink)
	defer os.Remove(dbFile)

	log.Printf("Downloading %s ...\n", dbFile)
	res, err := http.Get(dbLink)
	if err != nil {
		return "", err
	}

	f, err := os.Create(dbFile)
	if err != nil {
		return "", err
	}

	if _, err = io.Copy(f, res.Body); err != nil {
		return "", err
	}
	log.Printf("Download %s complete\n", dbFile)

	path, err := deCompress(dbFile, strings.Join([]string{c.DBLocationDir, "download/"}, "/"))
	if err != nil {
		return "", err
	}

	return path, nil
}

func deCompress(tarFile, dest string) (string, error) {
	srcFile, err := os.Open(tarFile)
	defer srcFile.Close()
	if err != nil {
		return "", err
	}

	gr, err := gzip.NewReader(srcFile)
	defer gr.Close()
	if err != nil {
		return "", err
	}

	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return "", err
			}
		}

		filename := dest + hdr.Name
		file, err := createFile(filename)
		if err != nil {
			return "", err
		}

		if file != nil {
			_, err = io.Copy(file, tr)
			if err != nil {
				return "", err
			}
		}

		_, dbName := filepath.Split(hdr.Name)
		if dbName == "GeoLite2-City.mmdb" || dbName == "GeoLite2-ASN.mmdb" {
			return filename, nil
		}
	}

	return "", errors.New("not found db file")
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

func wait(eg *errgroup.Group) chan error {
	ch := make(chan error, 1)

	go func() {
		ch <- eg.Wait()
	}()

	return ch
}
