/*
 * Revision History:
 *     Initial: 2018/05/23        Wang RiYu
 */

package tcos

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/mozillazg/go-cos"
	"github.com/mozillazg/go-cos/debug"
)

const isDebug = false // enable output http message log
const isLog = true // enable operation log

// AuthorizationConfig contains AppID, SecretID, SecretKey.
// AppID can be used to create a new bucket. SecretID & SecretKey are used for AuthorizationTransport.
// Get values from https://console.cloud.tencent.com/cam/capi
// Set policy for permission in https://console.cloud.tencent.com/cam/policy
type AuthorizationConfig struct {
	AppID     string // mandatory
	SecretID  string // mandatory
	SecretKey string // mandatory
}

// AuthorizationClient is used for services.
type AuthorizationClient struct {
	*AuthorizationConfig
	Client *cos.Client
}

// CreateAuthorizationClient creates a service client.
// Client contains BaseURL, common service, *ServiceService, *BucketService, *ObjectService
func CreateAuthorizationClient(config AuthorizationConfig) (*AuthorizationClient, error) {
	if err := CheckAuthorizationConfig(config); err != nil {
		return nil, err
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
		return nil, err
	}

	return c, nil
}

// Service returns a service, which contains Owner and Buckets.
func (c *AuthorizationClient) Service() (*cos.ServiceGetResult, error) {
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
	service, err := c.Service()
	if err != nil {
		return nil, err
	}

	if len(service.Buckets) > 0 {
		//for _, bucket := range service.Buckets {
		//	log.Printf("Buckets: %#v\n", bucket)
		//}

		return service.Buckets, nil
	}

	return []Bucket{}, errors.New("no bucket exists")
}

// CreateBucketClient is used to get a bucket client by AuthorizationClient
func (c *AuthorizationClient) CreateBucketClient(bucket *Bucket) (*BucketClient, error) {
	// e.g. BucketURL: https://<bucketname-appid>.cos.<region>.myqcloud.com
	// ListBuckets() return bucket like this: cos.Bucket{Name:"test-1255567152", AppID:"", Region:"ap-shanghai", CreateDate:""}
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

	err = HeadBucket(bc.Client)
	if err != nil {
		return nil, err
	}

	return bc, nil
}

// ConfirmAuthorization is used to confirm authorization info
func ConfirmAuthorization(client *cos.Client) error {
	if client == nil {
		return errors.New("missing client of authorization")
	}

	_, resp, err := client.Service.Get(context.Background())
	if resp != nil {
		if code := resp.StatusCode; 200 <= code && code <= 299 {
			return nil
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

// CheckAuthorizationConfig ...
func CheckAuthorizationConfig(config AuthorizationConfig) error {
	if config.AppID == "" {
		return errors.New("empty app ID")
	}
	if config.SecretID == "" {
		return errors.New("empty secret ID")
	}
	if config.SecretKey == "" {
		return errors.New("empty secret Key")
	}

	return nil
}

// OpError contains code & message, which describes the Err
type OpError struct {
	Code    string
	Message string
	Err     error
}

// Error return string format
func (err *OpError) Error() string {
	if err.Code != "" {
		return fmt.Sprintf("%s(%s): %s", err.Code, err.Message, err.Err)
	}

	return fmt.Sprintf("%s: %s", err.Message, err.Err)
}

// ErrConvert to OpError
func ErrConvert(err error) (*OpError, bool) {
	switch v := err.(type) {
	case *url.Error:
		if dnsErr, ok := v.Err.(*net.OpError).Err.(*net.DNSError); ok {
			// e.g. 'no such host' DNSError
			return &OpError{dnsErr.Err, dnsErr.Name, err}, true
		}
		return &OpError{v.Op, v.URL, v.Err}, true
	case *cos.ErrorResponse:
		return &OpError{v.Code, v.Message, err}, true
	}

	return nil, false
}
