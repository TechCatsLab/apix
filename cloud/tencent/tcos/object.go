/*
 * Revision History:
 *     Initial: 2018/05/28        Wang RiYu
 */

package tcos

import (
	"os"
	"context"
	"log"
	"io/ioutil"
	"path/filepath"

	"github.com/mozillazg/go-cos"
	"io"
)

// Object contains Key, ETag, Size, PartNumber, LastModified, StorageClass, Owner of created object
// e.g. Object{Key:"text.txt", ETag:"\"847f4281d3e9ad10844ef37da835cfc0\"", Size:4545, PartNumber:0, LastModified:"2018-05-22T14:39:23.000Z", StorageClass:"STANDARD", Owner:(*cos.Owner)(0xc42018bb00)}
type Object = cos.Object

// ObjectPutOptions see https://intl.cloud.tencent.com/document/product/436/7749#request-header
type ObjectPutOptions = cos.ObjectPutOptions

// ACLHeaderOptions ...
type ACLHeaderOptions = cos.ACLHeaderOptions

// ObjectPutHeaderOptions ...
type ObjectPutHeaderOptions = cos.ObjectPutHeaderOptions

// PutObject to bucket.
func (c *BucketClient) PutObject(name string, reader io.Reader, opt *ObjectPutOptions) error {
	_, err := c.Client.Object.Put(context.Background(), name, reader, opt)
	if err != nil {
		return err
	}

	return nil
}

// ListObjects is used to get all objects in bucket
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
