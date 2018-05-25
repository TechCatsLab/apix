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
type Config struct {
	AppID     string // optional
	SecretID  string // mandatory
	SecretKey string // mandatory
	Region    string // optional
}

// AuthorizationClient for get service
type AuthorizationClient struct {
	*Config
	Client *cos.Client
}

// BucketClient for bucket operation
type BucketClient struct {
	Client *cos.Client
}

const isDebug = false // 是否输出请求日志

// CreateAuthorizationClient creates a client for authorization.
func CreateAuthorizationClient(config Config) (*AuthorizationClient, error) {
	if config.SecretID == "" {
		return nil, errors.New("not valid secretid")
	}
	if config.SecretKey == "" {
		return nil, errors.New("not valid secretkey")
	}

	c := &AuthorizationClient{
		Config: &config,
	}

	c.Client = cos.NewClient(nil, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  config.SecretID,
			SecretKey: config.SecretKey,
			Transport: &debug.DebugRequestTransport{
				RequestHeader:  isDebug,
				RequestBody:    isDebug,
				ResponseHeader: isDebug,
				ResponseBody:   false,
			},
		},
	})

	return c, nil
}

// GetService returns a service
func (c *AuthorizationClient) ListService() (*cos.ServiceGetResult, error) {
	if c.Client == nil {
		return nil, errors.New("client with not Init")
	}

	service, _, err := c.Client.Service.Get(context.Background())
	if err != nil {
		log.Fatal(err)

		return nil, err
	}

	return service, nil
}

// ListBuckets for get all buckets of user
func (c *AuthorizationClient) ListBuckets() ([]cos.Bucket, error) {
	service, err := c.ListService()
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

// BucketClient for get a bucket operation client
func (c *AuthorizationClient) BucketClient(bucket *cos.Bucket) *cos.Client {
	bucketURL, _ := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", bucket.Name, bucket.Region))
	baseURL := &cos.BaseURL{BucketURL: bucketURL}

	return cos.NewClient(baseURL, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  c.Config.SecretID,
			SecretKey: c.Config.SecretKey,
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
func (c *AuthorizationClient) GetObjectsList(bucket *cos.Bucket) ([]cos.Object, error) {
	client := c.BucketClient(bucket)

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
