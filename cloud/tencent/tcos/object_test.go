/*
 * Revision History:
 *     Initial: 2018/05/28        Wang RiYu
 */

package tcos

import (
	"testing"
	"os"
)

var testObjectName = randNewStr(8)

func TestBucketClient_PutObject(t *testing.T) {
	ac := &AuthorizationConfig{APPID, SID, SKEY}
	config := &BucketConfig{
		AuthorizationConfig: ac,
		Name:                testObjectName,
		Region:              "ap-shanghai",
	}

	client, err := PutBucket(*config, nil)
	if err != nil {
		t.Error(err)
	}

	f, err := os.Open("object.go")
	if err != nil {
		t.Error(err)
	}
	s, err := f.Stat()
	if err != nil {
		t.Error(err)
	}
	opt := &ObjectPutOptions{
		ObjectPutHeaderOptions: &ObjectPutHeaderOptions{
			ContentLength: int(s.Size()),
		},
	}
	err = client.PutObject(s.Name(), f, opt)
	if err != nil {
		t.Error(err)
	}
}

func TestBucketClient_ListObjects(t *testing.T) {
	var (
		authorizationConfig = &AuthorizationConfig{
			AppID:     APPID,
			SecretID:  SID,
			SecretKey: SKEY,
		}
		config = &BucketConfig{
			AuthorizationConfig: authorizationConfig,
			Name:                testObjectName,
			Region:              "ap-shanghai",
		}
	)

	client, err := CreateBucketClient(*config)
	if err != nil {
		t.Error(err)
	}

	_, err = client.ListObjects(nil)
	if err != nil {
		t.Error(err)
	}
}
