/*
 * Revision History:
 *     Initial: 2018/05/23        Wang RiYu
 */

package cos

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/mozillazg/go-cos"
	"github.com/mozillazg/go-cos/debug"
	"github.com/pkg/errors"
)

// Config for client
// "AppID" optional
// "SecretID" mandatory, get from https://console.cloud.tencent.com/cam/capi
// "SecretKey" mandatory, setting policy for permission in https://console.cloud.tencent.com/cam/policy
// "Region" optional
type Config map[string]string

// InitClient for get service
type InitClient struct {
	Config
	Client *cos.Client
}

// BucketClient for bucket operation
type BucketClient struct {
	Client *cos.Client
}

const isDebug = false // 是否输出请求日志

// Init for get client config
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

// GetService returns a service
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

// GetBucketsList for get all buckets of user
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

// GetBucketClient for get a bucket operation client
func (ic *InitClient) GetBucketClient(bucket cos.Bucket) *cos.Client {
	bucketURL, _ := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", bucket.Name, bucket.Region))
	baseURL := &cos.BaseURL{BucketURL: bucketURL}

	return cos.NewClient(baseURL, &http.Client{
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

// GetObjectsList for get all objects in bucket
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

// Download files to local
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

	if err = CheckPath(path); err != nil {
		log.Fatal(err)

		return err
	}
	//log.Println(resp.Request.URL, resp.ContentLength, len(data), "\n")
	if err = ioutil.WriteFile(path+"/"+filename, data, 0666); err != nil {
		log.Fatal("write file err: ", err)

		return err
	}

	return nil
}

// CheckPath to create directory if the path doesn't exit
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
