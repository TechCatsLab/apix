### Usage

```
import "github.com/TechCatsLab/apix/cloud/tencent/tcos"
```

#### AuthorizationConfig

Get APPID, SecretID, SecretKey from https://console.cloud.tencent.com/cam/capi

```
var authorizationConfig = &tcos.AuthorizationConfig{
    AppID:     YourAPPID,
    SecretID:  YourSecretID,
    SecretKey: YourSecretKey,
}
```

#### GetServiceClient

```
authorizationClient, err := tcos.CreateAuthorizationClient(authorizationConfig)

buckets, err := authorizationClient.ListBuckets()
for _, bucket := range buckets {
    fmt.Printf("Bucket: %+v\n", bucket)
}
```

#### GetBucketClient

- put new bucket

```
bucketConfig := &tcos.BucketConfig{{
    AuthorizationConfig: authorizationConfig,
    Name:                NewBucketName,
    Region:              Region(e.g. "ap-shanghai"),
}

bucketClient, err := tcos.PutBucket(config, nil)
```

- from authorizationClient

```
buckets, err := authorizationClient.ListBuckets()

bucketClient, err := authorizationClient.CreateBucketClient(&buckets[0])
```

- from CreateBucketClient

```
bucketConfig := &tcos.BucketConfig{{
    AuthorizationConfig: authorizationConfig,
    Name:                BucketName,
    Region:              Region(e.g. "ap-shanghai"),
}

bucketClient, err := tcos.CreateBucketClient(bucketConfig)
```

#### BucketClient Operations

- PutCORS(opt *tcos.BucketPutCORSOptions)

```
opt := &tcos.BucketPutCORSOptions{
    Rules: []tcos.BucketCORSRule{
        {
            AllowedOrigins: []string{"http://www.qq.com"},
            AllowedMethods: []string{"PUT", "GET"},
            AllowedHeaders: []string{"x-cos-meta-test", "x-cos-xx"},
            MaxAgeSeconds:  500,
            ExposeHeaders:  []string{"x-cos-meta-test1"},
        },
        {
            ID:             "1234",
            AllowedOrigins: []string{"http://www.google.com", "twitter.com"},
            AllowedMethods: []string{"PUT", "GET"},
            MaxAgeSeconds:  500,
        },
    },
}
err = bucketClient.PutCORS(opt)
```

- GetCORS()

```
rules, err := bucketClient.GetCORS()
for _, r := range rules {
    fmt.Printf("%+v\n", r)
}
```

- PutObject(objectKey string, reader io.Reader, force bool, opt *ObjectPutOptions)

```
f, err := os.Open(filename)
s, err := f.Stat()

opt := &ObjectPutOptions{
    ObjectPutHeaderOptions: &ObjectPutHeaderOptions{
        ContentLength: int(s.Size()),
    },
}
resp, err = bucketClient.PutObject(s.Name(), f, false, opt)
```

- GetObject(objectKey string, opt *ObjectGetOptions)

```
resp, err := bucketClient.GetObject(objectKey, nil)
data, err := ioutil.ReadAll(resp.Body)
defer resp.Body.Close()
err = ioutil.WriteFile(filepath.Join(localPath, filename), data, 0666)
```

- Copy(sourceKey, destKey string, force bool, opt *ObjectCopyOptions)

```
res, resp, err = bucketClient.Copy("filename_1", "files/filename_1", true, nil)
```

- Move(sourceKey, destKey string, force bool, opt *ObjectCopyOptions)

```
res, resp, err = bucketClient.Move("filename_1", "files/filename_1", true, nil)
```

- Rename(sourceKey, fileName string, opt *ObjectCopyOptions)

```
res, resp, err = bucketClient.Rename("filename_1", "filename_2", true, nil)
```

- DeleteObject(objectKey string)

```
err = bucketClient.DeleteObject(filename)
```

- ListObjects(opt *BucketGetOptions)

```
list, err := bucketClient.ListObjects(nil)
for _, obj := range list {
    t.Logf("Object: %+v\n", obj)
}
```

- ObjectDownloadURL(objectKey string)

```
url, err := bucketClient.ObjectDownloadURL(objectKey)
```

- ObjectStaticURL(objectKey string) - NOTICE: This action need enable static website in basicconfig of bucket

```
url, err := bucketClient.ObjectStaticURL(objectKey)
```

- HeadObject(objectKey string, opt *ObjectHeadOptions)

```
resp, err = bucketClient.HeadObject(filename, nil)
```
