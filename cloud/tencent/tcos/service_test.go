/*
 * Revision History:
 *     Initial: 2018/05/23        Wang RiYu
 */

package tcos

import (
	"testing"
)

const (
	APPID = "1255567157"
	SID = "AKIDUsALNDBVwIUVFMD8mfKtKeqdefWFuu1y"
	SKEY = "KeyXXXXXXXXXXXXXX"
)

func TestCreateAuthorizationClient(t *testing.T) {
	var configs = []AuthorizationConfig{{
		AppID: "",
		SecretID:  SID,
		SecretKey: SKEY,
	}, {
		SecretID:  SID,
		SecretKey: "",
	}, {
		AppID: "sadasd",
		SecretKey: SKEY,
	}, {}, {
		SecretID:  SID,
		SecretKey: "unvaildkey",
	}}

	for i, config := range configs {
		_, err := CreateAuthorizationClient(config)

		switch i {
		case 0:
			if err != nil {
				t.Error(err)
			}
		case 1:
			if err.Error() != "empty secret Key" {
				t.Error(err)
			}
		case 2, 3:
			if err.Error() != "empty secret ID" {
				t.Error(err)
			}
		case 4:
			if err.Error() != "authorization failed" {
				t.Error(err)
			}
		}
	}
}

func TestListBuckets(t *testing.T) {
	var configs = []AuthorizationConfig{{
		SecretID:  SID,
		SecretKey: SKEY,
	}}

	if authorizationClient, err := CreateAuthorizationClient(configs[0]); err != nil {
		t.Error(err)
	} else {
		buckets, err := authorizationClient.ListBuckets()
		if err != nil {
			t.Error(err)
		}
		if buckets == nil {
			t.Error("unexpected nil of buckets")
		}
	}
}
