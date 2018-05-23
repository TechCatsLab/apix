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
	"net/http"
	"context"
	"io/ioutil"
	"os"
	"net/url"
	"log"
	"fmt"

	"github.com/mozillazg/go-cos"
	"github.com/mozillazg/go-cos/debug"
	"github.com/pkg/errors"
)

/*
* 配置
* "AppID" 可选
* "SecretID" 必须有，到 https://console.cloud.tencent.com/cam/capi 获取
* "SecretKey" 必须有，可以设置关联策略限制权限: https://console.cloud.tencent.com/cam/policy
* "Region" 可选
*/
type Config map[string]string

type InitClient struct {
	Config
	Client *cos.Client
}

type BucketClient struct {
	Client *cos.Client
}

const isDebug = false // 是否输出请求日志

/*
* 初始化配置
*/
func (ic *InitClient) Init(config Config) error {
	if id, ok := config["SecretID"]; !ok || id == "" {
		return errors.New("not valid secretid")
	}
	if k, ok := config["SecretKey"]; !ok || k == "" {
		return errors.New("not valid secretkey")
	}

	ic.Config = config
	ic.Client = cos.NewClient(nil, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  config["SecretID"],
			SecretKey: config["SecretKey"],
			Transport: &debug.DebugRequestTransport{
				RequestHeader:  isDebug,
				RequestBody:    isDebug,
				ResponseHeader: isDebug,
				ResponseBody:   false,
			},
		},
	})

	return nil
}

/*
* Get Service
*/
func (ic *InitClient) GetService() (*cos.ServiceGetResult, error) {
	if ic.Client == nil {
		return nil, errors.New("client with not Init")
	}

	service, _, err := ic.Client.Service.Get(context.Background())
	if err != nil {
		log.Fatal(err)

		return nil, err
	}

	return service, nil
}

/*
* 获取用户所有桶信息
*/
func (ic *InitClient) GetBucketsList() ([]cos.Bucket, error) {
	service, err := ic.GetService()
	if err != nil {
		return nil, err
	}

	if len(service.Buckets) > 0 {
		for _, bucket := range service.Buckets {
			fmt.Printf("Buckets: %#v\n", bucket)
		}

		return service.Buckets, nil
	}

	return []cos.Bucket{}, errors.New("no bucket exits")
}

/*
* 获取桶的 Client
*/
func (ic *InitClient) GetBucketClient(bucket cos.Bucket) *cos.Client {
	bucketUrl, _ := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", bucket.Name, bucket.Region))
	baseUrl := &cos.BaseURL{BucketURL: bucketUrl}

	return cos.NewClient(baseUrl, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  ic.Config["SecretID"],
			SecretKey: ic.Config["SecretKey"],
			Transport: &debug.DebugRequestTransport{
				RequestHeader:  isDebug,
				RequestBody:    isDebug,
				ResponseHeader: isDebug,
				ResponseBody:   false,
			},
		},
	})
}

/*
* 获取桶中对象列表
*/
func (ic *InitClient) GetObjectsList(bucket cos.Bucket) ([]cos.Object, error) {
	client := ic.GetBucketClient(bucket)

	opt := &cos.BucketGetOptions{}

	bucketInfo, _, err := client.Bucket.Get(context.Background(), opt)
	if err != nil {
		log.Fatal(err)

		return nil, err
	}

	for _, o := range bucketInfo.Contents {
		fmt.Printf("%#v\n", o)
	}

	return bucketInfo.Contents, nil
}

/*
* 下载文件到本地
*/
func Download(client *cos.Client, objectKey, path, filename string) error {
	resp, err := client.Object.Get(context.Background(), objectKey, nil)
	if err != nil {
		log.Fatal("get object err: ", err)

		return err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("read resp body err: ", err)

		return err
	}
	defer resp.Body.Close()

	//log.Println(resp.Request.URL, resp.ContentLength, len(data), "\n")
	if err = ioutil.WriteFile(path + "/" + filename, data, 0666); err != nil {
		log.Fatal("write file err: ", err)

		return err
	}

	return nil
}

/*
* 检查目录是否存在
*/
func CheckPath(path string) error {
	_, err := os.Stat(path)

	if err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(path, 0777); err != nil {
				return err
			}
		}
	}

	return nil
}