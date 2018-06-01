/*
 * Revision History:
 *     Initial: 2018/06/01        Wang RiYu
 */

package geoip2

import (
	"testing"
	"log"
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

	res, err := client.Lookup("118.25.101.120")
	if err != nil {
		t.Error(err)
	}
	log.Println(res)
}
