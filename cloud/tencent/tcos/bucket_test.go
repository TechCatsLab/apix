/*
 * Revision History:
 *     Initial: 2018/05/25        Wang RiYu
 */

package tcos

import (
	"math/rand"
	"testing"
	"time"
)

var testBucketName = randNewStr(8)

func TestPutBucket(t *testing.T) {
	ac := &AuthorizationConfig{APPID, SID, SKEY}
	configs := []BucketConfig{{
		AuthorizationConfig: ac,
		Name:                testBucketName,
		Region:              "ap-shanghai",
	}, {
		AuthorizationConfig: ac,
		Name:                testBucketName,
		Region:              "ap-beijing",
	}, {
		AuthorizationConfig: ac,
	}, {
		AuthorizationConfig: ac,
		Region:              "ap-beijing",
	}, {
		AuthorizationConfig: ac,
		Name:                "∫≈˚Ωß∆øßå¥/ª•sdjdsfn黄金本身的 6787",
		Region:              "ap-shanghai",
	}, {
		Name:   testBucketName,
		Region: "ap-beijing",
	}}

	opt := &BucketPutOptions{
		XCosACL: "XXXXXXXXXXX",
	}
	_, err := PutBucket(configs[0], opt)
	if opErr, ok := err.(*OpError); (ok && opErr.Code != "InvalidArgument") || !ok {
		t.Error(err)
	}

	for i, config := range configs {
		_, err := PutBucket(config, nil)
		switch i {
		case 0:
			if err != nil {
				t.Error(err)
			}
		case 1:
			if err.Error() != "BucketAlreadyExists" {
				t.Error(err)
			}
		case 3:
			if opErr, ok := err.(*OpError); (ok && opErr.Code != "InvalidURI") || !ok {
				t.Error(err)
			}
		case 2, 4:
			if dnsErr, ok := err.(*OpError); (ok && dnsErr.Code != "no such host") || !ok {
				t.Error(err)
			}
		case 5:
			if err.Error() != "missing AuthorizationConfig" {
				t.Error(err)
			}
		default:
			t.Error(i, err)
		}
	}
}

func TestCreateBucketClient(t *testing.T) {
	var (
		authorizationConfig = &AuthorizationConfig{
			AppID:     APPID,
			SecretID:  SID,
			SecretKey: SKEY,
		}
		configs = []BucketConfig{{
			AuthorizationConfig: authorizationConfig,
			Name:                testBucketName,
			Region:              "ap-shanghai",
		}, {
			Name:   testBucketName,
			Region: "ap-shanghai",
		}, {
			AuthorizationConfig: &AuthorizationConfig{"", "abc", "123"},
			Name:                testBucketName,
			Region:              "ap-shanghai",
		}, {
			AuthorizationConfig: &AuthorizationConfig{"789", "abc", "123"},
			Name:                testBucketName,
			Region:              "ap-shanghai",
		}, {
			AuthorizationConfig: authorizationConfig,
			Name:                "",
			Region:              "ap-shanghai",
		}, {
			AuthorizationConfig: authorizationConfig,
		}, {
			AuthorizationConfig: authorizationConfig,
			Name:                testBucketName,
			Region:              "",
		}, {
			AuthorizationConfig: authorizationConfig,
			Name:                testBucketName,
			Region:              "somewhere",
		}}
	)

	for i, config := range configs {
		_, err := CreateBucketClient(config)
		switch i {
		case 0:
			if err != nil {
				t.Error(err)
			}
		case 1:
			if err.Error() != "missing AuthorizationConfig" {
				t.Error(err)
			}
		case 2:
			if err.Error() != "empty app ID" {
				t.Error(err)
			}
		case 3:
			if opErr, ok := err.(*OpError); ok && opErr.Code != "InvalidAccessKeyId" {
				t.Error(err)
			}
		case 4, 5:
			if err.Error() != "bucket name is needed but none exits" {
				t.Error(err)
			}
		case 6:
			if err.Error() != "bucket region is needed but none exits" {
				t.Error(err)
			}
		case 7:
			if opErr, ok := err.(*OpError); ok && opErr.Code != "no such host" {
				t.Error(err)
			}
		default:
			t.Error(i, err)
		}
	}
}

func TestBucketClient_Delete(t *testing.T) {
	var (
		authorizationConfig = &AuthorizationConfig{
			AppID:     APPID,
			SecretID:  SID,
			SecretKey: SKEY,
		}
		configs = []BucketConfig{{
			AuthorizationConfig: authorizationConfig,
			Name:                testBucketName,
			Region:              "ap-shanghai",
		}, {
			AuthorizationConfig: &AuthorizationConfig{"", "abc", "123"},
			Name:                testBucketName,
			Region:              "ap-shanghai",
		}, {
			AuthorizationConfig: &AuthorizationConfig{"789", "abc", "123"},
			Name:                testBucketName,
			Region:              "ap-shanghai",
		}, {
			AuthorizationConfig: authorizationConfig,
			Name:                "nothisbucket",
			Region:              "ap-shanghai",
		}}
	)

	for i, config := range configs {
		client, err := CreateBucketClient(config)

		switch i {
		case 0:
			if err != nil {
				t.Error(err)
			} else if client != nil {
				if err = client.Delete(); err != nil {
					t.Error(err)
				}
			}
		case 1:
			if err.Error() != "empty app ID" {
				t.Error(err)
			}
		 case 2:
		 	if opErr, ok := err.(*OpError); ok && opErr.Code != "InvalidAccessKeyId" {
		 		t.Error(err)
		 	}
		case 3:
			if opErr, ok := err.(*OpError); ok && opErr.Message != "NoSuchBucket" {
				t.Error(err)
			}
		default:
			t.Error(i, err)
		}
	}
}

func randNewStr(length int) string {
	runes := "abcdefghijklmnopqrstuvwxyz"
	data := make([]byte, length)

	rand.Seed(time.Now().UnixNano())

	for i := 0; i < length; i++ {
		idx := rand.Intn(len(runes))
		data[i] = byte(runes[idx])
	}

	return string(data)
}
