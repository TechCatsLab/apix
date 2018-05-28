/*
 * Revision History:
 *     Initial: 2018/05/28        Wang RiYu
 */

package tcos

import (
	"os"
	"testing"
)

var (
	testObjectName      = randNewStr(8)
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
	filename = "object.go"
)

func TestBucketClient_PutObject(t *testing.T) {
	client, err := PutBucket(*config, nil)
	if err != nil {
		t.Fatal(err)
	}

	f, err := os.Open(filename)
	defer f.Close()
	if err != nil {
		t.Fatal(err)
	}
	s, err := f.Stat()
	if err != nil {
		t.Fatal(err)
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
	client, err := CreateBucketClient(*config)
	if err != nil {
		t.Fatal(err)
	}

	list, err := client.ListObjects(nil)
	for _, obj := range list {
		t.Logf("Object: %#v\n", obj)
	}
	if err != nil {
		t.Error(err)
	}
}

func TestBucketClient_HeadObject(t *testing.T) {
	client, err := CreateBucketClient(*config)
	if err != nil {
		t.Fatal(err)
	}

	err = client.HeadObject(filename, nil)
	if err != nil {
		t.Error(err)
	}

	err = client.HeadObject(filename, &ObjectHeadOptions{IfModifiedSince: "0"})
	if opErr, ok := err.(*OpError); ok && opErr.Message != "NotModified" {
		t.Error(opErr)
	}

	err = client.HeadObject("nothisobject", nil)
	if opErr, ok := err.(*OpError); ok && opErr.Message != "NoSuchObject" {
		t.Error(opErr)
	}

	//err = DownloadToLocal(client, filename, "files", filename)
	//if err != nil {
	//	t.Error(err)
	//}
}

func TestBucketClient_DeleteObject(t *testing.T) {
	client, err := CreateBucketClient(*config)
	if err != nil {
		t.Fatal(err)
	}

	err = client.DeleteObject(filename)
	if err != nil {
		t.Fatal(err)
	}

	err = client.Delete()
	if err != nil {
		t.Error(err)
	}
}
