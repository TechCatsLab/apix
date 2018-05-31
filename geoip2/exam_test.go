/*
 * Revision History:
 *     Initial: 2018/05/31        Wang RiYu
 */

package geoip2

import (
	"net"
	"testing"
	"net/http"
	"archive/tar"
	"compress/gzip"
	"os"
	"io"
	"path/filepath"
	"fmt"

	"github.com/oschwald/maxminddb-golang"
)

// License: https://dev.maxmind.com/geoip/geoip2/geolite2/
// The GeoLite2 databases are distributed under the Creative Commons Attribution-ShareAlike 4.0 International License.
// The attribution requirement may be met by including the following in all advertising and documentation mentioning features of or use of this database:
// This product includes GeoLite2 data created by MaxMind, available from <a href="http://www.maxmind.com">http://www.maxmind.com</a>.

// AutoUpdate: https://dev.maxmind.com/geoip/geoipupdate/

const test_ip = "118.25.101.120"
const city_db = "http://geolite.maxmind.com/download/geoip/database/GeoLite2-City.tar.gz"
const asn_db = "http://geolite.maxmind.com/download/geoip/database/GeoLite2-ASN.tar.gz"

func Test_Reader_Lookup(t *testing.T) {
	_, asn_db_name := filepath.Split(asn_db)
	if !pathExist(asn_db_name) {
		fmt.Printf("Downloading %s ...\n", asn_db_name)
		res, err := http.Get(asn_db)
		if err != nil {
			t.Fatal(err)
		}

		f, err := os.Create(asn_db_name)
		if err != nil {
			t.Fatal(err)
		}

		if _, err = io.Copy(f, res.Body); err != nil {
			os.Remove(asn_db_name)
			t.Fatal(err)
		}
		fmt.Printf("Download %s complete\n", asn_db_name)

		err = DeCompressAndMove(asn_db_name, "maxminddb/download/")
		if err != nil {
			t.Fatal(err)
		}
		os.RemoveAll("maxminddb/download/")
	}

	_, city_db_name := filepath.Split(city_db)
	if !pathExist(city_db_name) {
		fmt.Printf("Downloading %s ...\n", city_db_name)
		res, err := http.Get(city_db)
		if err != nil {
			t.Fatal(err)
		}

		f, err := os.Create(city_db_name)
		if err != nil {
			t.Fatal(err)
		}

		if _, err = io.Copy(f, res.Body); err != nil {
			os.Remove(city_db_name)
			t.Fatal(err)
		}
		fmt.Printf("Download %s complete\n", city_db_name)

		err = DeCompressAndMove(city_db_name, "maxminddb/download/")
		if err != nil {
			t.Fatal(err)
		}
		os.RemoveAll("maxminddb/download/")
	}

	db, err := maxminddb.Open("maxminddb/GeoLite2-City.mmdb")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	//err = db.Verify()
	//if err != nil {
	//	t.Error(err)
	//}

	ip := net.ParseIP(test_ip)

	var record struct {
		Continent struct {
			Code string `maxminddb:"code"`
			Names struct {
				ZhCN string `maxminddb:"zh-CN"`
				En   string `maxminddb:"en"`
			} `maxminddb:"names"`
		} `maxminddb:"continent"`
		Country struct {
			ISOCode string `maxminddb:"iso_code"`
			Names struct {
				ZhCN string `maxminddb:"zh-CN"`
				En   string `maxminddb:"en"`
			} `maxminddb:"names"`
		} `maxminddb:"country"`
		City struct {
			Names struct {
				ZhCN string `maxminddb:"zh-CN"`
				En   string `maxminddb:"en"`
			} `maxminddb:"names"`
		} `maxminddb:"city"`
		Subdivisions []struct {
			ISOCode string `maxminddb:"iso_code"`
			Names struct {
				ZhCN string `maxminddb:"zh-CN"`
				En   string `maxminddb:"en"`
			} `maxminddb:"names"`
		} `maxminddb:"subdivisions"`
		Location struct {
			TimeZone        string  `maxminddb:"time_zone"`
			Latitude        float32 `maxminddb:"latitude"`
			Longitude       float32 `maxminddb:"longitude"`
			Accuracy_radius int     `maxminddb:"accuracy_radius"`
		} `maxminddb:"location"`
		Registered_country struct {
			ISOCode string `maxminddb:"iso_code"`
			Names struct {
				ZhCN string `maxminddb:"zh-CN"`
				En   string `maxminddb:"en"`
			} `maxminddb:"names"`
		} `maxminddb:"registered_country"`
		Organization string
	}

	err = db.Lookup(ip, &record)
	if err != nil {
		t.Fatal(err)
	}

	asn_db, err := maxminddb.Open("maxminddb/GeoLite2-ASN.mmdb")
	if err != nil {
		t.Fatal(err)
	}
	defer asn_db.Close()

	var asn struct {
		Organization string `maxminddb:"autonomous_system_organization"`
	}
	err = asn_db.Lookup(ip, &asn)
	if err != nil {
		t.Fatal(err)
	}
	record.Organization = asn.Organization

	t.Log("IP:", ip)
	t.Logf("Continent: %+v\n", record.Continent)
	t.Logf("Country: %+v\n", record.Country)
	t.Logf("City: %+v\n", record.City)
	t.Logf("Organization: %+v\n", record.Organization)
	t.Logf("Location: %+v\n", record.Location)
}

func DeCompressAndMove(tarFile, dest string) error {
	fmt.Println(tarFile)
	srcFile, err := os.Open(tarFile)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	gr, err := gzip.NewReader(srcFile)
	if err != nil {
		return err
	}
	defer gr.Close()

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

		_, err = io.Copy(file, tr)
		if err != nil {
			return err
		}

		_, db_name := filepath.Split(hdr.Name)
		if db_name == "GeoLite2-City.mmdb" || db_name == "GeoLite2-ASN.mmdb" {
			err = os.Rename(filename, "maxminddb/" + db_name)
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
