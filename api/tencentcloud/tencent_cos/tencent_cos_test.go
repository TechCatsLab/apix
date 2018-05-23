/*
 * MIT License
 *
 * Copyright (c) 2018 SmartestEE Co., Ltd..
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

/*
 * Revision History:
 *     Initial: 2018/05/23        Wang RiYu
 */

package tencent_cos

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
							if  err != nil {
								t.Error(err)
							}
							wg.Done()
						}(object.Key, "files/" + path, filename)
					}
				}
				wg.Wait()
			}
		}
	}
}
