/*
 * Revision History:
 *     Initial: 2018/05/24        Wang RiYu
 */

package tcos

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"

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
func PutBucket(config BucketConfig, options *BucketPutOptions) (*BucketClient, error) {
	if config.AuthorizationConfig == nil {
		return nil, errors.New("missing AuthorizationConfig")
	}
	if err := CheckAuthorizationConfig(*config.AuthorizationConfig); err != nil {
		return nil, err
	}

	bucketURL, err := url.Parse(fmt.Sprintf("https://%s-%s.cos.%s.myqcloud.com", config.Name, config.AppID, config.Region))
	if err != nil {
		return nil, err
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
		return nil, errors.New("BucketAlreadyExists")
	}
	if err != nil {
		if opErr, ok := ErrConvert(err); ok {
			return nil, opErr
		}
		return nil, err
	}

	bc := &BucketClient{&config, c}
	if isLog {
		log.Printf("Put bucket \"%s\"\n", bc.Name)
	}

	return bc, nil
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
		return nil, err
	}
	err = HeadBucket(c.Client)
	if err != nil {
		return nil, err
	}

	return c, nil
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

	if isLog {
		log.Printf("Delete bucket \"%s\"\n", c.Name)
	}

	return nil
}

// HeadBucket tests bucket is available or not
// Status: 200 - ok, 403 - Forbidden, 404 - Not Found
func HeadBucket(client *cos.Client) error {
	resp, err := client.Bucket.Head(context.Background())
	if resp != nil {
		switch resp.StatusCode {
		case 200:
			return nil
		case 403:
			return &OpError{"403", "AccessDenied", err}
		case 404:
			return &OpError{"404", "NoSuchBucket", err}
		}
	}
	if err != nil {
		if opErr, ok := ErrConvert(err); ok {
			return opErr
		}
		return err
	}

	return nil
}
