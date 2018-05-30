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
	if opErr := checkErr(err); opErr.Code != "InvalidArgument" {
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
			if opErr := checkErr(err); opErr.Code != "InvalidURI" {
				t.Error(err)
			}
		case 2, 4:
			if opErr := checkErr(err); opErr.Code != "no such host" {
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
			if opErr := checkErr(err); opErr.Code != "InvalidAccessKeyId" {
				t.Error(err)
			}
		case 4, 5:
			if err.Error() != "bucket name is needed but none exists" {
				t.Error(err)
			}
		case 6:
			if err.Error() != "bucket region is needed but none exists" {
				t.Error(err)
			}
		case 7:
			if opErr := checkErr(err); opErr.Code != "no such host" {
				t.Error(err)
			}
		default:
			t.Error(i, err)
		}
	}
}

func TestBucketClient_PutCORS(t *testing.T) {
	var config = &BucketConfig{
		AuthorizationConfig: &AuthorizationConfig{
			AppID:     APPID,
			SecretID:  SID,
			SecretKey: SKEY,
		},
		Name:   testBucketName,
		Region: "ap-shanghai",
	}

	client, err := CreateBucketClient(*config)
	if err != nil {
		t.Fatal(err)
	}

	err = client.PutCORS(nil)
	if err.Error() != "400 InvalidArgument" {
		t.Error(err)
	}

	opt := &BucketPutCORSOptions{
		Rules: []BucketCORSRule{
			{
				AllowedOrigins: []string{"http://www.qq.com"},
				AllowedMethods: []string{"PUT", "GET"},
				AllowedHeaders: []string{"x-cos-meta-test", "x-cos-xx"},
				MaxAgeSeconds:  500,
				ExposeHeaders:  []string{"x-cos-meta-test1"},
			},
			{
				ID:             "1234",
				AllowedOrigins: []string{"http://www.google.com", "twitter.com"},
				AllowedMethods: []string{"PUT", "GET"},
				MaxAgeSeconds:  500,
			},
		},
	}
	err = client.PutCORS(opt)
	if err != nil {
		t.Error(err)
	}
}

func TestBucketClient_GetCORS(t *testing.T) {
	var config = &BucketConfig{
		AuthorizationConfig: &AuthorizationConfig{
			AppID:     APPID,
			SecretID:  SID,
			SecretKey: SKEY,
		},
		Name:   testBucketName,
		Region: "ap-shanghai",
	}

	client, err := CreateBucketClient(*config)
	if err != nil {
		t.Fatal(err)
	}

	rules, err := client.GetCORS()
	if err != nil {
		t.Error(err)
	}
	for _, r := range rules {
		t.Logf("CORS: %+v\n", r)
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
				t.Fatal(err)
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
			if opErr := checkErr(err); opErr.Code != "InvalidAccessKeyId" {
				t.Error(err)
			}
		case 3:
			if opErr := checkErr(err); opErr.Message != "NoSuchBucket" {
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

func checkErr(err error) *OpError {
	switch v := err.(type) {
	case *OpError:
		return v
	default:
		return &OpError{Err: err}
	}
}
