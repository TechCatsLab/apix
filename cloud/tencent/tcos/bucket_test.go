/*
 * Revision History:
 *     Initial: 2018/05/25        Wang RiYu
 */

package tcos

import (
	"path/filepath"
	"sync"
	"testing"
)

func TestBucket(t *testing.T) {
	var config = &AuthorizationConfig{
		SecretID:  SID,
		SecretKey: SKEY,
	}

	if authorizationClient, err := CreateAuthorizationClient(*config); err != nil {
		t.Error(err)
	} else {
		buckets, err := authorizationClient.ListBuckets()
		if err != nil {
			t.Error(err)
		}

		if len(buckets) > 0 {
			for i := range buckets {
				bucketClient, err := authorizationClient.CreateBucketClient(&buckets[i])
				if err != nil {
					t.Error(err)
				}

				opt := &BucketGetOptions{}
				objects, err := bucketClient.ListObjects(opt)
				if err != nil {
					t.Error(err)
				}

				wg := sync.WaitGroup{}
				for _, object := range objects {
					if object.Size > 0 {
						wg.Add(1)
						go func(objectKey string) {
							path, filename := filepath.Split(objectKey)
							if err = Download(bucketClient.Client, objectKey, "files/" + path, filename); err != nil {
								t.Error(err)
							}
							wg.Done()
						}(object.Key)
					}
				}
				wg.Wait()
			}
		}
	}
}

func TestPutBucket(t *testing.T) {
	ac := &AuthorizationConfig{
		SecretID: SID,
		SecretKey: SKEY,
	}
	config := &BucketConfig{}
	config.AuthorizationConfig = ac
	config.AppID = APPID
	config.Name = "test"
	config.Region = "ap-beijing"
	resp, err := PutBucket(*config, nil)
	if err != nil {
		t.Errorf("%v\n%#v\n", err, resp.Response)
	}
}
