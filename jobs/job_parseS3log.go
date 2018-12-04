package jobs

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/santrancisco/cque"
)

type ParseS3LogJob struct {
	Bucketname string
	Objectpath string
}

type Connection struct {
	From string
	To   string
	Port int64
}

func (c Connection) String() string {
	o := c.From + " - " + c.To + " - " + string(c.Port)
	return o
}

const (
	KeyParseS3Log = "parse_s3_log"
)

func ParseS3Log(logger *log.Logger, qc *cque.Client, j *cque.Job, appconfig map[string]interface{}) error {
	o := j.Args.(ParseS3LogJob)
	log.Printf("[DEBUG] Parsing S3 log from: %s/%s \n", o.Bucketname, o.Objectpath)
	data, err := Getlogsfroms3(o.Bucketname, o.Objectpath, appconfig)
	if err != nil {
		return err
	}
	downloadonly := appconfig["downloadonly"].(bool)
	if downloadonly {
		return nil
	}
	connections := []Connection{}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		s := strings.Split(scanner.Text(), " ")
		fromport, _ := strconv.ParseInt(s[6], 10, 0)
		toport, _ := strconv.ParseInt(s[7], 10, 0)
		if fromport > toport {
			fromip := s[4]
			toip := s[5]
			if Isinternal(toip) {
				if Isinternal(fromip) {
					isnew := true
					for _, i := range connections {
						if (i.From == fromip) && (i.To == toip) && (i.Port == toport) {
							isnew = false
							break
						}
					}
					if isnew {
						c := Connection{
							From: fromip,
							To:   toip,
							Port: toport,
						}
						connections = append(connections)
						qc.Result <- cque.Result{
							JobType: j.Type,
							Result:  c,
						}
					}
				}
			}
			// fmt.Printf("%s -> %s:%s\n", s[4], s[5], s[7])
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

// Getlogsfroms3 - Download flowlog from s3, decompress using gzip
func Getlogsfroms3(bucket, key string, appconfig map[string]interface{}) ([]byte, error) {
	var gzdata []byte
	cache := appconfig["cache"].(string)
	cachefile := filepath.Join(cache, bucket, key)
	_ = os.MkdirAll(filepath.Dir(cachefile), 0777)
	log.Printf("[INFO] Getting object from S3 %s", filepath.Join(bucket, key))
	// Get from cache if exists.
	downloadonly := appconfig["downloadonly"].(bool)
	if _, err := os.Stat(cachefile); !os.IsNotExist(err) {
		log.Printf("[INFO] Found cache file for %s", filepath.Join(bucket, key))
		if downloadonly {
			return nil, nil
		}
		gzdata, err = ioutil.ReadFile(cachefile)
		if err != nil {
			return nil, err
		}
	} else {
		// Download from S3
		sess, _ := session.NewSession()
		downloader := s3manager.NewDownloader(sess)
		buff := &aws.WriteAtBuffer{}
		_, err := downloader.Download(buff,
			&s3.GetObjectInput{
				Bucket: aws.String(bucket),
				Key:    aws.String(key),
			})
		if err != nil {
			return nil, err
		}
		gzdata = buff.Bytes()
		ioutil.WriteFile(cachefile, gzdata, 0777)
	}
	if downloadonly {
		return nil, nil
	}
	// Decompress gz files
	r := bytes.NewReader(gzdata)

	gzf, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(gzf)
	if err != nil {
		return nil, err
	}
	return data, nil

}

// Check If ip is not between internal ranges
// 	10.0.0.0 – 10.255.255.255
//  172.16.0.0 – 172.31.255.255
//  192.168.0.0 – 192.168.255.255
func Isinternal(ip string) bool {
	i := iptoint(ip)
	if !(((i > 167772160) && (i < 184549375)) || ((i > 2886729728) && (i < 2887778303)) || ((i > 3232235520) && (i < 3232301055))) {
		return false
	}
	return true
}

// Convert ip string to an integer
func iptoint(ips string) int64 {
	ip := net.ParseIP(ips)
	IPv4Int := big.NewInt(0)
	IPv4Int.SetBytes(ip.To4())
	return IPv4Int.Int64()
}
