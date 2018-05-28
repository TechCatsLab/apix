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

// BucketConfig contains AuthorizationConfig, Name, Region.
// AuthorizationConfig is needed, for more information see service.go
// When creating new bucket, Name & Region is mandatory.
type BucketConfig struct {
	*AuthorizationConfig        // mandatory
	Name                 string // mandatory
	Region               string // mandatory, see: https://intl.cloud.tencent.com/document/product/436/6224
}

// BucketClient is used for bucket operation
type BucketClient struct {
	*BucketConfig
	Client *cos.Client
}

// Bucket contains Name, AppID, Region, CreateDate of created bucket
type Bucket = cos.Bucket

// BucketGetOptions contains request parameters of ListObjects()
// see options details in https://intl.cloud.tencent.com/document/product/436/7734#request-parameters
type BucketGetOptions = cos.BucketGetOptions

// BucketPutOptions in https://intl.cloud.tencent.com/document/product/436/7738#request-header
type BucketPutOptions = cos.BucketPutOptions

// PutBucket is used to create a new bucket.
// options: https://intl.cloud.tencent.com/document/product/436/7738#request-header
func PutBucket(config BucketConfig, options *BucketPutOptions) error {
	if config.AuthorizationConfig == nil {
		return errors.New("missing AuthorizationConfig")
	}
	if err := CheckAuthorizationConfig(*config.AuthorizationConfig); err != nil {
		return err
	}

	bucketURL, err := url.Parse(fmt.Sprintf("https://%s-%s.cos.%s.myqcloud.com", config.Name, config.AppID, config.Region))
	if err != nil {
		return err
	}

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
		options = &BucketPutOptions{
			XCosACL: "public-read",
		}
	}
	resp, err := c.Bucket.Put(context.Background(), options)
	if resp != nil && resp.StatusCode == 409 {
		return errors.New("BucketAlreadyExists")
	}
	if err != nil {
		if opErr, ok := ErrConvert(err); ok {
			return opErr
		}
		return err
	}

	return nil
}

// CreateBucketClient creates a bucket client, which can request bucket operations
func CreateBucketClient(config BucketConfig) (*BucketClient, error) {
	if config.AuthorizationConfig == nil {
		return nil, errors.New("missing AuthorizationConfig")
	}
	if err := CheckAuthorizationConfig(*config.AuthorizationConfig); err != nil {
		return nil, err
	}
	if config.Name == "" {
		return nil, errors.New("bucket name is needed but none exits")
	}
	if config.Region == "" {
		return nil, errors.New("bucket region is needed but none exits")
	}

	c := &BucketClient{
		BucketConfig: &config,
	}

	bucketURL, err := url.Parse(fmt.Sprintf("https://%s-%s.cos.%s.myqcloud.com", config.Name, config.AppID, config.Region))
	if err != nil {
		return nil, err
	}

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

	err = ConfirmAuthorization(c.Client)
	if err != nil {
		if opErr, ok := ErrConvert(err); ok {
			return nil, opErr
		}
		return nil, err
	}
	err = ConfirmBucket(c.Client)
	if err != nil {
		if opErr, ok := ErrConvert(err); ok {
			return nil, opErr
		}
		return nil, err
	}

	return c, nil
}

// ListObjects is used to get all objects in bucket
// see options details in https://intl.cloud.tencent.com/document/product/436/7734#request-parameters
func (c *BucketClient) ListObjects(opt *BucketGetOptions) ([]cos.Object, error) {
	bucketInfo, _, err := c.Client.Bucket.Get(context.Background(), opt)
	if err != nil {
		return nil, err
	}

	//for _, object := range bucketInfo.Contents {
	//	log.Fatalln("%#v\n", object)
	//}

	return bucketInfo.Contents, nil
}

// Delete the bucket. The bucket must be empty before deleting.
func (c *BucketClient) Delete() error {
	resp, err := c.Client.Bucket.Delete(context.Background())
	if resp != nil && resp.StatusCode == 409 {
		return errors.New("BucketNotEmpty")
	}
	if resp != nil && resp.StatusCode == 403 {
		return errors.New("AccessDenied")
	}
	if resp != nil && resp.StatusCode == 404 {
		return errors.New("NoSuchBucket")
	}
	if err != nil {
		return err
	}

	return nil
}

// ConfirmBucket is used to confirm bucket client is available or not
func ConfirmBucket(client *cos.Client) error {
	if client == nil {
		return errors.New("missing client of bucket")
	}

	opt := &cos.BucketGetOptions{}
	_, _, err := client.Bucket.Get(context.Background(), opt)
	if err != nil {
		opErr, _ := ErrConvert(err)

		return opErr
	}

	return nil
}

// Download files to local
func Download(client *cos.Client, objectKey, path, filename string) error {
	resp, err := client.Object.Get(context.Background(), objectKey, nil)
	if err != nil {
		log.Fatalln("get object err: ", err)

		return err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln("read resp body err: ", err)

		return err
	}
	defer resp.Body.Close()

	if err = checkPath(path); err != nil {
		log.Fatalln(err)

		return err
	}

	//log.Println(resp.Request.URL, resp.ContentLength, len(data), "\n")
	if err = ioutil.WriteFile(filepath.Join(path, filename), data, 0666); err != nil {
		log.Fatalln("write file err: ", err)

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
