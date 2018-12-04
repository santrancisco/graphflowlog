package jobs

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/santrancisco/cque"
)

type CrawlBucketJob struct {
	Bucketname string
	Bucketpath string
	Depth      int
}

const (
	KeyCrawlBucket = "crawl_bucket"
)

// Example - we can wrap the other function
func CrawlBucket(logger *log.Logger, qc *cque.Client, j *cque.Job, appconfig map[string]interface{}) error {
	b := j.Args.(CrawlBucketJob)
	log.Printf("[INFO] Getting flow log from bucket %s - %s", b.Bucketname, b.Bucketpath)
	subfolders, err := Getsubfolder(b.Bucketname, b.Bucketpath)
	if err != nil {
		return err
	}

	if len(subfolders) > 0 {
		for _, networkinterfacefolder := range subfolders {
			files, err := Listfileinfolder(b.Bucketname, networkinterfacefolder, b.Depth)
			if err != nil {
				return err
			}
			for _, file := range files {
				qc.Enqueue(cque.Job{
					Type: KeyParseS3Log,
					Args: ParseS3LogJob{Bucketname: b.Bucketname, Objectpath: file},
				})
			}
		}
	}

	return nil
}

func Getsubfolder(bucket, path string) ([]string, error) {
	// Initialize a session in us-west-2 that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials.
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}

	_, err = sess.Config.Credentials.Get()
	if err != nil {
		return nil, err
	}

	// Create S3 service client
	svc := s3.New(sess)

	// Get the list of items
	resp, err := svc.ListObjects(&s3.ListObjectsInput{Bucket: aws.String(bucket), Prefix: aws.String(path), Delimiter: aws.String("/")})
	if err != nil {
		return nil, fmt.Errorf("Unable to list items in bucket %q, %v", bucket, err)
	}

	subfolders := []string{}
	for _, folder := range resp.CommonPrefixes {
		subfolders = append(subfolders, *folder.Prefix)
	}

	return subfolders, nil

}

func Listfileinfolder(bucket, path string, depth int) ([]string, error) {
	// Initialize a session in us-west-2 that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials.
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}

	_, err = sess.Config.Credentials.Get()
	if err != nil {
		return nil, err
	}

	// Create S3 service client
	svc := s3.New(sess)

	// Get the list of items
	resp, err := svc.ListObjects(&s3.ListObjectsInput{Bucket: aws.String(bucket), Prefix: aws.String(path), Delimiter: aws.String("/")})
	if err != nil {
		return nil, fmt.Errorf("Unable to list items in bucket %q, %v", bucket, err)
	}
	files := []string{}
	l := len(resp.Contents)
	if l >= depth {
		for i := (l - depth); i < l; i++ {
			files = append(files, *resp.Contents[i].Key)
		}
	} else {
		for i := 0; i < l; i++ {
			files = append(files, *resp.Contents[i].Key)

		}
	}
	return files, nil
	// fmt.Println("Found", len(resp.Contents), "items in bucket", bucket)
	// fmt.Println("")
	// return nil, nil
}
