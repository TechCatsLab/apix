/*
 * Revision History:
 *     Initial: 2018/05/28        Wang RiYu
 */

package tcos

import (
	"os"
	_ "io/ioutil"
	"testing"
	"fmt"
)

var (
	testObjectName      = randNewStr(8)
	authorizationConfig = &AuthorizationConfig{
		AppID:     APPID,
		SecretID:  SID,
		SecretKey: SKEY,
	}
	bucketConfig = &BucketConfig{
		AuthorizationConfig: authorizationConfig,
		Name:                testObjectName,
		Region:              "ap-shanghai",
	}
	filesname = []string{"object.go", "bucket.go", "service.go"}
)

func TestBucketClient_PutObject(t *testing.T) {
	client, err := PutBucket(*bucketConfig, nil)
	if err != nil {
		t.Fatal(err)
	}

	for _, filename := range filesname {
		f, err := os.Open(filename)
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
		_, err = client.PutObject(s.Name(), f, false, opt)
		if err != nil {
			t.Error(err)
		}
	}

	f, err := os.Open(filesname[0])
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
	_, err = client.PutObject(fmt.Sprintf("files/%s", s.Name()), f, false, opt)
	if err != nil {
		t.Error(err)
	}

	_, err = client.PutObject("newfolder/", nil, false, nil)
	if err != nil {
		t.Error(err)
	}

	_, err = client.PutObject("newfolder/", nil, false, nil)
	if err.Error() != "ObjectAlreadyExists(enable force if you want to overwrite)" {
		t.Error(err)
	}
}

func TestBucketClient_Copy(t *testing.T) {
	client, err := CreateBucketClient(*bucketConfig)
	if err != nil {
		t.Fatal(err)
	}

	// copy bucket.go to files/
	_, _, err = client.Copy(filesname[1], "files/" + filesname[1], true, nil)
	if err != nil {
		t.Error(err)
	}

	// overwrite no force
	_, _, err = client.Copy(filesname[0], "files/" + filesname[0], false, nil)
	if err.Error() != "ObjectAlreadyExists(enable force if you still want to copy)" {
		t.Error(err)
	}

	// overwrite object.go with force
	_, _, err = client.Copy(filesname[0], "files/" + filesname[0], true, nil)
	if err != nil {
		t.Error(err)
	}

	// move files/bucket.go to newfolder/
	_, _, err = client.Move("files/" + filesname[1], "newfolder/" + filesname[1], true, nil)
	if err != nil {
		t.Error(err)
	}

	// rename but filename conflicts with other files
	_, _, err = client.Rename(filesname[2], filesname[1], nil)
	if err.Error() != "this action conflicts with other files" {
		t.Error(err)
	}

	// rename newfolder/bucket.go
	_, _, err = client.Rename("newfolder/" + filesname[1], "rename.go", nil)
	if err != nil {
		t.Error(err)
	}
}

func TestBucketClient_ListObjects(t *testing.T) {
	client, err := CreateBucketClient(*bucketConfig)
	if err != nil {
		t.Fatal(err)
	}

	list, err := client.ListObjects(nil)
	for _, obj := range list {
		t.Logf("Object: %+v\n", obj)
	}
	if err != nil {
		t.Error(err)
	}
}

func TestBucketClient_GetObject(t *testing.T) {
	client, err := CreateBucketClient(*bucketConfig)
	if err != nil {
		t.Fatal(err)
	}

	// Get files
	for _, filename := range filesname {
		resp, err := client.GetObject(filename, nil)
		if resp.Status != "200 OK" {
			t.Error(err)
		}
	}

	// Get folder
	resp, err := client.GetObject("newfolder/", nil)
	if resp.Status != "200 OK" {
		t.Error(err)
	}

	// Get non-existent key
	resp, err = client.GetObject("NoSuchKey", nil)
	if opErr := checkErr(err); opErr.Code != "NoSuchKey" {
		t.Error(err)
	}

	// Get partial content
	opt := &ObjectGetOptions{
		Range: "bytes=0-70",
	}
	resp, err = client.GetObject(filesname[0], opt)
	if resp.Status != "206 Partial Content" {
		t.Error(err)
	} else {
		//bs, _ := ioutil.ReadAll(resp.Body)
		//resp.Body.Close()
		//t.Logf("%s\n", string(bs))
	}

	// InvalidRange
	opt = &ObjectGetOptions{
		Range: "XXX",
	}
	resp, err = client.GetObject(filesname[0], opt)
	if opErr := checkErr(err); opErr.Code != "InvalidRange" {
		t.Error(err)
	}

	// InvalidArgument of IfModifiedSince
	opt = &ObjectGetOptions{
		IfModifiedSince: "2018-05-29T05:55:46.000Z",
	}
	resp, err = client.GetObject(filesname[0], opt)
	if opErr := checkErr(err); opErr.Message != "The If-Modified-Since you specified is not valid" {
		t.Error(err)
	}
}

func TestBucketClient_HeadObject(t *testing.T) {
	client, err := CreateBucketClient(*bucketConfig)
	if err != nil {
		t.Fatal(err)
	}

	for _, filename := range filesname {
		_, err = client.HeadObject(filename, nil)
		if err != nil {
			t.Error(err)
		}

		_, err = client.HeadObject(filename, &ObjectHeadOptions{IfModifiedSince: "0"})
		if opErr := checkErr(err); opErr.Message != "NotModified" {
			t.Error(err)
		}

		//err = DownloadToLocal(client, filename, "files", filename)
		//if err != nil {
		//	t.Error(err)
		//}
	}

	_, err = client.HeadObject("files/" + filesname[0], nil)
	if err != nil {
		t.Error(err)
	}

	_, err = client.HeadObject("files/" + filesname[0], &ObjectHeadOptions{IfModifiedSince: "0"})
	if opErr := checkErr(err); opErr.Message != "NotModified" {
		t.Error(err)
	}

	_, err = client.HeadObject("newfolder/", nil)
	if err != nil {
		t.Error(err)
	}

	_, err = client.HeadObject("newfolder/", &ObjectHeadOptions{IfModifiedSince: "0"})
	if opErr := checkErr(err); opErr.Message != "NotModified" {
		t.Error(err)
	}

	_, err = client.HeadObject("nothisobject", nil)
	if opErr := checkErr(err); opErr.Message != "NoSuchObject" {
		t.Error(err)
	}
}

func TestBucketClient_DeleteObject(t *testing.T) {
	client, err := CreateBucketClient(*bucketConfig)
	if err != nil {
		t.Fatal(err)
	}

	err = client.Delete()
	if opErr := checkErr(err); opErr.Message != "BucketNotEmpty" {
		t.Error(err)
	}

	for _, filename := range filesname {
		err = client.DeleteObject(filename)
		if err != nil {
			t.Error(err)
		}
	}

	err = client.DeleteObject("files/" + filesname[0])
	if err != nil {
		t.Error(err)
	}

	err = client.DeleteObject("files/")
	if err != nil {
		t.Error(err)
	}

	err = client.DeleteObject("newfolder/rename.go")
	if err != nil {
		t.Error(err)
	}

	err = client.DeleteObject("newfolder/")
	if err != nil {
		t.Error(err)
	}
}
