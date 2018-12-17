package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/santrancisco/cque"
	"github.com/santrancisco/graphflowlog/jobs"
	"github.com/santrancisco/logutils"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	verbose      = kingpin.Flag("verbose", "Enable debug mode.").Default("false").Short('v').Bool()
	cache        = kingpin.Flag("cache", "cache location for flowlog files").Default("flowlogcache").Short('c').String()
	bucketname   = kingpin.Flag("bucket", "bucket name").Default("\000").Short('b').String()
	downloadonly = kingpin.Flag("downloadonly", "Enable downloadonly mode.").Default("false").Bool()
	bucketpath   = kingpin.Flag("path", "s3 bucket path to where flow logs for different interfaces are stored").Default("").Short('p').String()
	depth        = kingpin.Flag("depth", "How deep we want to crawl for each interface").Default("2").Short('d').Int()
	threads      = kingpin.Flag("threads", "How many worker thread should we spawn").Default("4").Short('t').Int()
	outfolder    = kingpin.Flag("out", "location where we want to save the result and used for the frontend code").Default("").Short('o').String()
	// Ignore the following arguments for now :), Using environment variable for AWS credentials is much better/easier.
	// accesskey    = kingpin.Flag("accesskey", "access key for s3 access").Default("").OverrideDefaultFromEnvar("AWS_ACCESS_KEY_ID").String()
	// secretkey    = kingpin.Flag("secretkey", "secret key for s3 access").Default("").OverrideDefaultFromEnvar("AWS_SECRET_ACCESS_KEY").String()
	// sessiontoken = kingpin.Flag("sessiontoken", "aws_session_token for s3 access").Default("").OverrideDefaultFromEnvar("AWS_SESSION_TOKEN").String()
	// region       = kingpin.Flag("region", "aws region").Default("").OverrideDefaultFromEnvar("AWS_REGION").String()
	// sqlitedbpath = kingpin.Flag("sqlpath", "Local sqlite cache location. Default is ~/.drat.sqlite").Default("~/.drat.sqlite").Short('s').String()
)

func main() {
	kingpin.Version("0.0.1")
	kingpin.Parse()
	config := map[string]interface{}{}
	config["depth"] = *depth
	config["bucketname"] = *bucketname
	config["bucketpath"] = *bucketpath
	config["cache"] = *cache
	config["downloadonly"] = *downloadonly

	// Make sure cache folder exist.
	err := os.MkdirAll(*cache, 0777)
	if err != nil {
		log.Fatal(err)
	}

	// config["sqlitedbpath"] = *sqlitedbpath

	// Configuring our log level
	logfilter := "WARNING"
	if *verbose {
		logfilter = "DEBUG"
	}

	filteroutput := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "WARNING", "INFO", "ERROR"},
		MinLevel: logutils.LogLevel(logfilter),
		Writer:   os.Stderr,
	}
	log.SetOutput(filteroutput)

	ctx, cancel := context.WithCancel(context.Background())
	// defering canclation of all concurence processes
	defer cancel()

	qc := cque.NewQue()
	wpool := cque.NewWorkerPool(qc, cque.WorkMap{
		jobs.KeyCrawlBucket: (&jobs.JobFuncWrapper{
			QC:        qc,
			Logger:    log.New(filteroutput, "", log.LstdFlags),
			F:         jobs.CrawlBucket,
			AppConfig: config}).Run,

		jobs.KeyParseS3Log: (&jobs.JobFuncWrapper{
			QC:        qc,
			Logger:    log.New(filteroutput, "", log.LstdFlags),
			F:         jobs.ParseS3Log,
			AppConfig: config}).Run,
	}, *threads)

	wpool.Start(ctx)
	qc.Enqueue(cque.Job{
		Type: jobs.KeyCrawlBucket,
		Args: jobs.CrawlBucketJob{Bucketname: *bucketname, Bucketpath: *bucketpath, Depth: *depth},
	})

	rh := ResultHandler{
		WaitForResult: true,
		WaitTimeStart: time.Now(),
		qc:            qc,
		ctx:           ctx,
	}
	go rh.Run()
	running := true
	for running {
		// If we have been waiting for result for more than 1s.
		if rh.WaitForResult && (time.Since(rh.WaitTimeStart) > (time.Duration(1) * time.Second)) {
			shouldnotrunning := qc.IsQueueEmpty && rh.WaitForResult
			running = !shouldnotrunning
		} else {
			running = true
		}
		// when queue is not empty we loop mainthread.
		time.Sleep(1 * time.Second)
	}
	log.Printf("[INFO] Creating data.json")
	// After collecting all our data, time to spit out our graph data.
	rh.Data.Updatexasis()
	rh.Data.AddNodeIndex()
	rh.Data.AddEdgeIndexes(rh.TempEdges)
	b, err := json.MarshalIndent(rh.Data, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fname := filepath.Join(*outfolder, "data.json")
	err = ioutil.WriteFile(fname, b, 0777)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("[INFO] %s is ready", fname)
	// log.Print(string(b))

	// if *outtofile != "" {
	// 	f, err := os.OpenFile(*outtofile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	// 	if err != nil {
	// 		fmt.Println(string(b))
	// 		fmt.Printf("Having issue with opening file %s. The result is printed to stdout \n", *outtofile)
	// 		log.Fatal(err)
	// 	}
	// 	f.Write(b)
	// } else {
	// 	// If we don't specify a file output, dump it to stdout
	// 	fmt.Println(string(b))
	// }

}
