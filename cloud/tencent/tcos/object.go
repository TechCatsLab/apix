/*
 * Revision History:
 *     Initial: 2018/05/28        Wang RiYu
 */

package tcos

import (
	"context"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"fmt"
	"regexp"
	"errors"

	"github.com/mozillazg/go-cos"
)

// Object contains Key, ETag, Size, PartNumber, LastModified, StorageClass, Owner of created object
// e.g. Object{Key:"text.txt", ETag:"\"847f4281d3e9ad10844ef37da835cfc0\"", Size:4545, PartNumber:0, LastModified:"2018-05-22T14:39:23.000Z", StorageClass:"STANDARD", Owner:(*cos.Owner)(0xc42018bb00)}
type Object = cos.Object

// ObjectCopyResult ...
type ObjectCopyResult = cos.ObjectCopyResult

// ObjectPutOptions contains ACLHeaderOptions and ObjectPutHeaderOptions
// details see https://intl.cloud.tencent.com/document/product/436/7749#request-header
type ObjectPutOptions = cos.ObjectPutOptions

// ACLHeaderOptions see Permission-related headers in https://intl.cloud.tencent.com/document/product/436/7749#non-common-header
type ACLHeaderOptions = cos.ACLHeaderOptions

// ObjectPutHeaderOptions see Recommended Header in https://intl.cloud.tencent.com/document/product/436/7749#non-common-header
type ObjectPutHeaderOptions = cos.ObjectPutHeaderOptions

// ObjectGetOptions is used for GetObject()
// details see https://intl.cloud.tencent.com/document/product/436/7753
type ObjectGetOptions = cos.ObjectGetOptions

// ObjectHeadOptions specified "IfModifiedSince" Header
type ObjectHeadOptions = cos.ObjectHeadOptions

// ObjectCopyOptions contains ObjectCopyHeaderOptions and ACLHeaderOptions
type ObjectCopyOptions = cos.ObjectCopyOptions

// ObjectCopyHeaderOptions see https://cloud.tencent.com/document/product/436/10881#.E9.9D.9E.E5.85.AC.E5.85.B1.E5.A4.B4.E9.83.A8
type ObjectCopyHeaderOptions = cos.ObjectCopyHeaderOptions

// GetObject ...
func (c *BucketClient) GetObject(objectKey string, opt *ObjectGetOptions) (*http.Response, error) {
	if objectKey == "" {
		return nil, errors.New("empty objectKey")
	}

	resp, err := c.Client.Object.Get(context.Background(), objectKey, opt)
	if err != nil {
		if opErr, ok := ErrConvert(err); ok {
			return resp.Response, opErr
		}
		if resp != nil {
			return resp.Response, err
		}
		return nil, err
	}

	return resp.Response, nil
}

// PutObject to bucket. This action requires WRITE permission for the Bucket.
// enable force will overwrite object if ObjectAlreadyExists
func (c *BucketClient) PutObject(objectKey string, reader io.Reader, force bool, opt *ObjectPutOptions) (*http.Response, error) {
	if objectKey == "" {
		return nil, errors.New("empty objectKey")
	}
	// TODO: test object unvalid key
	if reg, err := regexp.MatchString(`[\^|\&|\/|\||\s]`, objectKey); reg == true && err != nil {
		return nil, errors.New("objectKey cannot contain any ^&/| or whitespace")
	}

	_, err := c.GetObject(objectKey, nil)
	if err == nil && !force {
		return nil, errors.New("ObjectAlreadyExists(enable force if you want to overwrite)")
	}

	if dir, _ := filepath.Split(objectKey); dir != "" {
		if resp, err := c.Client.Object.Put(context.Background(), dir, nil, nil); err != nil {
			if resp != nil {
				return resp.Response, err
			}
			return nil, err
		}
	}

	resp, err := c.Client.Object.Put(context.Background(), objectKey, reader, opt)
	if err != nil {
		if resp != nil {
			return resp.Response, err
		}
		return nil, err
	}

	if isLog {
		log.Printf("Put object \"%s\" in bucket \"%s\"\n", objectKey, c.Name)
	}

	return resp.Response, nil
}

// Copy sourceKey to destKey.
// Enable force will overwrite file if destKey exists.
// This action can be used to copy, move, rename and reset object
func (c *BucketClient) Copy(sourceKey, destKey string, force bool, opt *ObjectCopyOptions) (*ObjectCopyResult, *http.Response, error) {
	if sourceKey == "" || destKey == "" {
		return nil, nil, errors.New("empty key")
	}

	_, err := c.GetObject(sourceKey, nil)
	if err != nil {
		return nil, nil, err
	}

	_, sourceName := filepath.Split(sourceKey)
	_, destName := filepath.Split(destKey)
	_, err = c.GetObject(destKey, nil)
	if err == nil && sourceName == destName && !force {
		return nil, nil, errors.New("ObjectAlreadyExists(enable force if you still want to copy)")
	}

	// NOTICE: sourceURL is Host/path/file, no Scheme, e.g. <sourcebucket-appid>.cos.<region>.myqcloud/sourcekey
	sourceURL := fmt.Sprintf("%s/%s", c.Client.BaseURL.BucketURL.Host, sourceKey)
	res, resp, err := c.Client.Object.Copy(context.Background(), destKey, sourceURL, opt)
	if err != nil {
		if resp != nil {
			return res, resp.Response, err
		}
		return res, nil, err
	}

	if isLog {
		log.Printf("Copy object \"%s\" to \"%s\" in bucket \"%s\"\n", sourceKey, destKey, c.Name)
	}

	return res, resp.Response, nil
}

// Copy ...
func Copy(client *BucketClient, sourceKey, destKey string, force bool, opt *ObjectCopyOptions) (*ObjectCopyResult, *http.Response, error) {
	if sourceKey == "" || destKey == "" {
		return nil, nil, errors.New("empty key")
	}

	_, err := client.GetObject(sourceKey, nil)
	if err != nil {
		return nil, nil, err
	}

	_, sourceName := filepath.Split(sourceKey)
	_, destName := filepath.Split(destKey)
	_, err = client.GetObject(destKey, nil)
	if err == nil && sourceName == destName && !force {
		return nil, nil, errors.New("ObjectAlreadyExists(enable force if you still want to copy)")
	}

	// NOTICE: sourceURL is Host/path/file, no Scheme, e.g. <sourcebucket-appid>.cos.<region>.myqcloud/sourcekey
	sourceURL := fmt.Sprintf("%s/%s", client.Client.BaseURL.BucketURL.Host, sourceKey)
	res, resp, err := client.Client.Object.Copy(context.Background(), destKey, sourceURL, opt)
	if err != nil {
		if resp != nil {
			return res, resp.Response, err
		}
		return res, nil, err
	}

	return res, resp.Response, nil
}

// Move object
func (c *BucketClient) Move(sourceKey, destKey string, force bool, opt *ObjectCopyOptions) (*ObjectCopyResult, *http.Response, error) {
	res, resp, err := Copy(c, sourceKey, destKey, force, opt)
	if err != nil {
		return res, resp, err
	}

	_, err = c.Client.Object.Delete(context.Background(), sourceKey)
	if err != nil {
		return res, resp, &OpError{"Move with err", "delete sourceKey failed", err}
	}

	if isLog {
		log.Printf("Move object \"%s\" to \"%s\" in bucket \"%s\"\n", sourceKey, destKey, c.Name)
	}

	return res, resp, nil
}

// Rename object
// TODO: Rename directory
func (c *BucketClient) Rename(sourceKey, fileName string, opt *ObjectCopyOptions) (*ObjectCopyResult, *http.Response, error) {
	// TODO: test object unvalid name
	if reg, err := regexp.MatchString(`[\\|\^|\&|\/|\||\s]`, fileName); reg == true && err != nil {
		return nil, nil, errors.New("filename cannot contain any \\^&/| or whitespace")
	}

	path, _ := filepath.Split(sourceKey)
	destKey := filepath.Join(path, fileName)
	_, err := c.GetObject(destKey, nil)
	if err == nil {
		return nil, nil, errors.New("this action conflicts with other files")
	}

	res, resp, err := Copy(c, sourceKey, destKey, false, opt)
	if err != nil {
		return res, resp, err
	}

	_, err = c.Client.Object.Delete(context.Background(), sourceKey)
	if err != nil {
		return res, resp, &OpError{"Rename with err", "delete sourceKey failed", err}
	}

	if isLog {
		log.Printf("Rename object \"%s\" to \"%s\" in bucket \"%s\"\n", sourceKey, destKey, c.Name)
	}

	return res, resp, nil
}

// ListObjects is used to get all objects in bucket. This action requires READ permission for the Bucket.
// see options details in https://intl.cloud.tencent.com/document/product/436/7734#request-parameters
func (c *BucketClient) ListObjects(opt *BucketGetOptions) ([]Object, error) {
	if opt == nil {
		opt = &BucketGetOptions{}
	}

	bucketInfo, _, err := c.Client.Bucket.Get(context.Background(), opt)
	if err != nil {
		return nil, err
	}

	return bucketInfo.Contents, nil
}

// DeleteObject is used to delete one file (Object) in Bucket. This action requires WRITE permission for the Bucket.
// if the named object doesn't exist, delete operation is ok and return 204 - No Content
func (c *BucketClient) DeleteObject(objectKey string) error {
	if objectKey == "" {
		return errors.New("empty objectKey")
	}

	_, err := c.Client.Object.Delete(context.Background(), objectKey)
	if err != nil {
		return err
	}

	if isLog {
		log.Printf("Delete object \"%s\" in bucket \"%s\"\n", objectKey, c.Name)
	}

	return nil
}

// ObjectDownloadURL ...
func (c *BucketClient) ObjectDownloadURL(objectKey string) (string, error) {
	if objectKey == "" {
		return "", errors.New("empty objectKey")
	}

	_, err := c.GetObject(objectKey, nil)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s", c.Client.BaseURL.BucketURL, objectKey), nil
}

// ObjectStaticURL return static url, which can be embedded in website
// NOTICE: This action need enable static website in basicconfig of bucket
func (c *BucketClient) ObjectStaticURL(objectKey string) (string, error) {
	if objectKey == "" {
		return "", errors.New("empty objectKey")
	}

	_, err := c.GetObject(objectKey, nil)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("https://%s-%s.cos-website.%s.myqcloud.com/%s", c.BucketConfig.Name, c.BucketConfig.AppID, c.BucketConfig.Region, objectKey), nil
}

// HeadObject requests object meta info.
// if opt specified "IfModifiedSince" Header and object is not modified, response 304
func (c *BucketClient) HeadObject(objectKey string, opt *ObjectHeadOptions) (*http.Response, error) {
	if objectKey == "" {
		return nil, errors.New("empty objectKey")
	}

	resp, err := c.Client.Object.Head(context.Background(), objectKey, opt)
	if resp != nil {
		switch resp.StatusCode {
		case 404:
			return resp.Response, &OpError{"404", "NoSuchObject", err}
		case 304:
			return resp.Response, &OpError{"304", "NotModified", err}
		}
	}
	if err != nil {
		if resp != nil {
			return resp.Response, err
		}
		return nil, err
	}

	return resp.Response, nil
}

// DownloadObject ...
func (c *BucketClient) DownloadObject(objectKey string, writer io.Writer) error {
	if objectKey == "" {
		return errors.New("empty objectKey")
	}

	resp, err := c.GetObject(objectKey, nil)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, resp.Body)

	return err
}

// DownloadToLocal object to local
func DownloadToLocal(client *BucketClient, objectKey, localPath, filename string) error {
	if objectKey == "" {
		return errors.New("empty objectKey")
	}

	resp, err := client.GetObject(objectKey, nil)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		log.Fatalln("read resp body err: ", err)

		return err
	}

	if err = checkPath(localPath); err != nil {
		log.Fatalln(err)

		return err
	}

	//log.Println(resp.Request.URL, resp.ContentLength, len(data), "\n")
	if err = ioutil.WriteFile(filepath.Join(localPath, filename), data, 0666); err != nil {
		log.Fatalln("write file err: ", err)

		return err
	}

	return nil
}

// CheckPath is used to create directory if the path doesn't exist
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
