/*
 * Revision History:
 *     Initial: 2018/05/23        Wang RiYu
 */

package tcos

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"errors"

	"github.com/mozillazg/go-cos"
	"github.com/mozillazg/go-cos/debug"
)

// AuthorizationConfig is used for Authorization.
// get values from https://console.cloud.tencent.com/cam/capi
// set policy for permission in https://console.cloud.tencent.com/cam/policy
type AuthorizationConfig struct {
	AppID     string // optional, AppID is used to creating new bucket
	SecretID  string // mandatory
	SecretKey string // mandatory
}

// AuthorizationClient is used for services.
type AuthorizationClient struct {
	*AuthorizationConfig
	Client *cos.Client
}

const isDebug = false // enable output http message log

// CreateAuthorizationClient creates a client for GetService().
func CreateAuthorizationClient(config AuthorizationConfig) (*AuthorizationClient, error) {
	if config.SecretID == "" {
		return nil, errors.New("empty secret ID")
	}
	if config.SecretKey == "" {
		return nil, errors.New("empty secret Key")
	}

	c := &AuthorizationClient{
		AuthorizationConfig: &config,
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

	err := ConfirmAuthorization(c.Client)
	if err != nil {
		return nil, errors.New("authorization failed")
	}

	return c, nil
}

// GetService returns a service
func (c *AuthorizationClient) GetService() (*cos.ServiceGetResult, error) {
	if c.Client == nil {
		return nil, errors.New("unauthorized client")
	}

	service, _, err := c.Client.Service.Get(context.Background())
	if err != nil {
		return nil, err
	}

	return service, nil
}

// ListBuckets is used to get all buckets of authorized user
func (c *AuthorizationClient) ListBuckets() ([]Bucket, error) {
	service, err := c.GetService()
	if err != nil {
		return nil, err
	}

	if len(service.Buckets) > 0 {
		for _, bucket := range service.Buckets {
			fmt.Printf("Buckets: %#v\n", bucket)
		}

		return service.Buckets, nil
	}

	return []Bucket{}, errors.New("no bucket exits")
}

// CreateBucketClient is used to get a bucket client by AuthorizationClient
func (c *AuthorizationClient) CreateBucketClient(bucket *Bucket) (*BucketClient, error) {
	bucketURL, err := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", bucket.Name, bucket.Region))
	if err != nil {
		return nil, err
	}
	baseURL := &cos.BaseURL{BucketURL: bucketURL}

	bc := &BucketClient{}
	bucketConfig := &BucketConfig{}
	bucketConfig.AuthorizationConfig = c.AuthorizationConfig

	bc.BucketConfig = bucketConfig
	bc.Client = cos.NewClient(baseURL, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  c.AuthorizationConfig.SecretID,
			SecretKey: c.AuthorizationConfig.SecretKey,
			Transport: &debug.DebugRequestTransport{
				RequestHeader:  isDebug,
				RequestBody:    isDebug,
				ResponseHeader: isDebug,
				ResponseBody:   false,
			},
		},
	})

	err = ConfirmBucket(bc.Client)
	if err != nil {
		return nil, errors.New("unavailable bucket client")
	}

	return bc, nil
}

// ConfirmAuthorization is used to confirm authorization info
func ConfirmAuthorization(client *cos.Client) error {
	_, resp, err := client.Service.Get(context.Background())
	if code := resp.StatusCode; !(200 <= code && code <= 299) && err != nil {
		return err
	}

	return nil
}
