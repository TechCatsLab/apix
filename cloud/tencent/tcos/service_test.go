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
	SID   = "AKIDmW6azJ7hOyUpz26kP7h6PYWJwW2NeZYO"
	SKEY  = "KEY-XXXXX"
)

func TestCreateAuthorizationClient(t *testing.T) {
	var configs = []AuthorizationConfig{{
		AppID:     APPID,
		SecretID:  SID,
		SecretKey: SKEY,
	}, {
		AppID:     APPID,
		SecretID:  SID,
		SecretKey: "",
	}, {
		AppID:     APPID,
		SecretKey: SKEY,
	}, {}, {
		AppID:     APPID,
		SecretID:  SID,
		SecretKey: "unvalidkey",
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
		case 2:
			if err.Error() != "empty secret ID" {
				t.Error(err)
			}
		case 3:
			if err.Error() != "empty app ID" {
				t.Error(err)
			}
		case 4:
			if opErr := checkErr(err); opErr.Code != "SignatureDoesNotMatch" {
				t.Error(err)
			}
		default:
			t.Error(i, err)
		}
	}
}

func TestAuthorizationClient_ListBuckets(t *testing.T) {
	var configs = []AuthorizationConfig{{
		AppID:     APPID,
		SecretID:  SID,
		SecretKey: SKEY,
	}}

	if authorizationClient, err := CreateAuthorizationClient(configs[0]); err != nil {
		t.Fatal(err)
	} else if authorizationClient != nil {
		buckets, err := authorizationClient.ListBuckets()
		for _, bucket := range buckets {
			t.Logf("Bucket: %+v\n", bucket)
		}
		if err != nil {
			t.Error(err)
		}
		if buckets == nil {
			t.Error("unexpected nil of buckets")
		}
	}
}

func TestAuthorizationClient_CreateBucketClient(t *testing.T) {
	var config = &AuthorizationConfig{
		AppID:     APPID,
		SecretID:  SID,
		SecretKey: SKEY,
	}

	if authorizationClient, err := CreateAuthorizationClient(*config); err != nil {
		t.Fatal(err)
	} else if authorizationClient != nil {
		buckets, err := authorizationClient.ListBuckets()
		if err != nil {
			t.Error(err)
		}

		for i := range buckets {
			_, err := authorizationClient.CreateBucketClient(&buckets[i])
			if err != nil {
				t.Error(err)
			}
		}
	}

	client, err := CreateBucketClient(*bucketConfig)
	if err != nil {
		t.Fatal(err)
	}
	err = client.Delete()
	if err != nil {
		t.Error(err)
	}
}
