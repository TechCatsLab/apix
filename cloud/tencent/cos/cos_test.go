/*
 * Revision History:
 *     Initial: 2018/05/23        Wang RiYu
 */

package cos

import (
	"path/filepath"
	"sync"
	"testing"
)

func TestService(t *testing.T) {
	var (
		config = Config{
			"AppID":     "",
			"SecretID":  "AKID9E9WmU9UyaAqtmwHhmFg4Thltv6VhvP9",
			"SecretKey": "ZzLae2Hd5MtqMuYK31srWVT0T5iGuush",
			"Region":    "ap-beijing",
		}
		initClient = new(InitClient)
	)

	if err := initClient.Init(config); err != nil {
		t.Error(err)
	} else {
		buckets, err := initClient.GetBucketsList()
		if err != nil {
			t.Error(err)
		}

		if len(buckets) > 0 {
			for _, bucket := range buckets {
				objects, err := initClient.GetObjectsList(bucket)
				client := initClient.GetBucketClient(bucket)
				if err != nil {
					t.Error(err)
				}

				wg := sync.WaitGroup{}
				for _, object := range objects {
					if object.Size > 0 {
						wg.Add(1)
						path, filename := filepath.Split(object.Key)
						CheckPath("files/" + path)
						go func(objectKey, path, filename string) {
							err = Download(client, objectKey, path, filename)
							if err != nil {
								t.Error(err)
							}
							wg.Done()
						}(object.Key, "files/"+path, filename)
					}
				}
				wg.Wait()
			}
		}
	}
}
