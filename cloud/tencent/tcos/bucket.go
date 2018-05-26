/*
 * Revision History:
 *     Initial: 2018/05/24        Wang RiYu
 */

package tcos

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/mozillazg/go-cos"
	"github.com/mozillazg/go-cos/debug"
)

// BucketConfig is used for BucketClient
type BucketConfig struct {
	*AuthorizationConfig // AppID is mandatory when creating new bucket
	Name   string // mandatory, bucket name
	Region string // mandatory, see: https://intl.cloud.tencent.com/document/product/436/6224
}

// BucketClient is used for bucket operation
type BucketClient struct {
	*BucketConfig
	Client *cos.Client
}

// Bucket contains Name, AppID, Region, CreateDate
type Bucket = cos.Bucket

// BucketGetOptions contains request parameters of ListObjects()
// see options details in https://intl.cloud.tencent.com/document/product/436/7734#request-parameters
type BucketGetOptions = cos.BucketGetOptions

// CreateBucketClient creates a bucket client
func CreateBucketClient(config BucketConfig) (*BucketClient, error) {
	if config.AuthorizationConfig.SecretID == "" {
		return nil, errors.New("empty secret ID")
	}
	if config.AuthorizationConfig.SecretKey == "" {
		return nil, errors.New("empty secret Key")
	}

	c := &BucketClient{
		BucketConfig: &config,
	}

	bucketURL, _ := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", config.Name, config.Region))
	baseURL := &cos.BaseURL{BucketURL: bucketURL}
	c.Client = cos.NewClient(baseURL, &http.Client{
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

	err := ConfirmAuthorization(c.Client)
	if err != nil {
		return nil, errors.New("authorization failed")
	}
	err = ConfirmBucket(c.Client)
	if err != nil {
		return nil, errors.New("unavailable bucket client")
	}

	return c, nil
}

// PutBucket is used to create a new bucket.
// options: https://intl.cloud.tencent.com/document/product/436/7738#request-header
func PutBucket(config BucketConfig, options *cos.BucketPutOptions) (*cos.Response, error) {
	bucketURL, _ := url.Parse(fmt.Sprintf("https://%s-%s.cos.%s.myqcloud.com", config.Name, config.AppID, config.Region))
	baseURL := &cos.BaseURL{
		BucketURL: bucketURL,
	}
	c := cos.NewClient(baseURL, &http.Client{
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

	if options == nil {
		options = &cos.BucketPutOptions{
			XCosACL: "public-read",
		}
	}
	resp, err := c.Bucket.Put(context.Background(), options)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

// ListObjects is used to get all objects in bucket
func (c *BucketClient) ListObjects(opt *BucketGetOptions) ([]cos.Object, error) {
	bucketInfo, _, err := c.Client.Bucket.Get(context.Background(), opt)
	if err != nil {
		return nil, err
	}

	//for _, object := range bucketInfo.Contents {
	//	fmt.Printf("%#v\n", object)
	//}

	return bucketInfo.Contents, nil
}

// ConfirmBucket is used to confirm bucket client is available
func ConfirmBucket(client *cos.Client) error {
	opt := &cos.BucketGetOptions{}
	_, resp, err := client.Bucket.Get(context.Background(), opt)
	if code := resp.StatusCode; !(200 <= code && code <= 299) && err != nil {
		return err
	}

	return nil
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

	if err = checkPath(path); err != nil {
		log.Fatal(err)

		return err
	}

	//log.Println(resp.Request.URL, resp.ContentLength, len(data), "\n")
	if err = ioutil.WriteFile(filepath.Join(path, filename), data, 0666); err != nil {
		log.Fatal("write file err: ", err)

		return err
	}

	return nil
}

// CheckPath is used to create directory if the path doesn't exit
func checkPath(path string) error {
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
