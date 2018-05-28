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

	"github.com/mozillazg/go-cos"
	"github.com/pkg/errors"
)

// Object contains Key, ETag, Size, PartNumber, LastModified, StorageClass, Owner of created object
// e.g. Object{Key:"text.txt", ETag:"\"847f4281d3e9ad10844ef37da835cfc0\"", Size:4545, PartNumber:0, LastModified:"2018-05-22T14:39:23.000Z", StorageClass:"STANDARD", Owner:(*cos.Owner)(0xc42018bb00)}
type Object = cos.Object

// ObjectPutOptions contains ACLHeaderOptions and ObjectPutHeaderOptions
// details see https://intl.cloud.tencent.com/document/product/436/7749#request-header
type ObjectPutOptions = cos.ObjectPutOptions

// ACLHeaderOptions ...
type ACLHeaderOptions = cos.ACLHeaderOptions

// ObjectPutHeaderOptions ...
type ObjectPutHeaderOptions = cos.ObjectPutHeaderOptions

// ObjectGetOptions is used for GetObject()
// details see https://intl.cloud.tencent.com/document/product/436/7753
type ObjectGetOptions = cos.ObjectGetOptions

// ObjectHeadOptions specified "IfModifiedSince" Header
type ObjectHeadOptions = cos.ObjectHeadOptions

// PutObject to bucket. This action requires WRITE permission for the Bucket.
func (c *BucketClient) PutObject(objectKey string, reader io.Reader, opt *ObjectPutOptions) error {
	if objectKey == "" {
		return errors.New("empty objectKey")
	}

	_, err := c.Client.Object.Put(context.Background(), objectKey, reader, opt)
	if err != nil {
		return err
	}

	if isLog {
		log.Printf("Put object \"%s\" in bucket \"%s\"\n", objectKey, c.Name)
	}

	return nil
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
// if the named object doesn't exit, delete operation is ok and return 204 - No Content
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

// GetObject ...
func (c *BucketClient) GetObject(objectKey string, opt *ObjectGetOptions) (*http.Response, error) {
	if objectKey == "" {
		return nil, errors.New("empty objectKey")
	}

	resp, err := c.Client.Object.Get(context.Background(), objectKey, opt)
	if err != nil {
		if opErr, ok := ErrConvert(err); ok {
			return nil, opErr
		}
		return nil, err
	}

	return resp.Response, nil
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

	return fmt.Sprintf("https://%s-%s.cos.%s.myqcloud.com/%s", c.BucketConfig.Name, c.BucketConfig.AppID, c.BucketConfig.Region, objectKey), nil
}

// HeadObject requests object meta info.
// if opt specified "IfModifiedSince" Header and object is not modified, response 304
func (c *BucketClient) HeadObject(objectKey string, opt *ObjectHeadOptions) error {
	if objectKey == "" {
		return errors.New("empty objectKey")
	}

	resp, err := c.Client.Object.Head(context.Background(), objectKey, opt)
	if resp != nil {
		switch resp.StatusCode {
		case 404:
			return &OpError{"404", "NoSuchObject", err}
		case 304:
			return &OpError{"304", "NotModified", err}
		}
	}
	if err != nil {
		return err
	}

	return nil
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
	if err != nil {
		log.Fatalln("read resp body err: ", err)

		return err
	}
	defer resp.Body.Close()

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
