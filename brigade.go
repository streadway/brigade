package main

import (
	"encoding/json"
	"encoding/xml"
	"strconv"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"expvar"
	"time"
)

const maxKeys = 1000

var listen = flag.String("http", ":8080", "Listen address")

var FeedCounters = expvar.NewMap("FeedCounters")

type Content struct {
	Key          string `json:"key"`
	LastModified string `json:"lastModified"`
	ETag         string `json:"eTag"`
	Size         uint64 `json:"size"`
}

type ListBucketResult struct {
	Name        string
	Prefix      string
	Marker      string
	MaxKeys     int
	IsTruncated bool
	Contents    []Content
}

type feed struct {
	id       string
	bucket   string
	msg      string
	contents chan Content
	marker   string
	proto    string
}

type procs struct {
  m sync.Mutex
  running map[string]*feed
}

var feeds = &procs{
  running: make(map[string]*feed),
}

func (f *feed) close(msg string) {
	f.msg = msg
	close(f.contents)
}

func (f *feed) run() {
	for retries := 0; retries < 1000; {
		batch := fmt.Sprintf("%s://%s.s3.amazonaws.com/?max-keys=%d&marker=%s",
			f.proto, f.bucket, maxKeys, f.marker)

		res, err := http.Get(batch)

		if err != nil {
			// IO error
			f.close(err.Error())
			break
		} else if res.StatusCode >= 400 && res.StatusCode < 500 {
			// Fatal error code range
			f.close(res.Status)
			break
		} else if res.StatusCode >= 500 {
			// Exponential backoff, will get service eventually
			backoff := rand.Int63n(int64(math.Pow(2, float64(retries))))
			time.Sleep(time.Second * time.Duration(backoff))
			retries += 1
			continue
		}

		bucket := ListBucketResult{}

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			// Broken transport reads retry immediately
			continue
		}

		err = xml.Unmarshal(body, &bucket)
		if err != nil {
			f.close(err.Error())
		}

		// Termination clause
		if len(bucket.Contents) == 0 {
			f.close("")
			break
		}

		for _, content := range bucket.Contents {
			f.contents <- content
      FeedCounters.Add(f.key(), 1)
		}

		f.marker = bucket.Contents[len(bucket.Contents)-1].Key

		retries = 0
	}
}

func (f *feed) key() string {
  return f.bucket + "/" + f.id
}

func (p *procs) join(bucketName, jobId string) *feed {
	p.m.Lock()
	defer p.m.Unlock()

  var (
    started, proto *feed
    ok bool
  )

  proto = &feed{
    id:       jobId,
    proto:    "https",
    bucket:   bucketName,
    contents: make(chan Content, maxKeys*2), // avoid partial initialization
  }

	if started, ok = p.running[proto.key()]; !ok {
		started, p.running[proto.key()] = proto, proto
		go started.run()
	}

	return started
}

func consume(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache, no-store, private")
	w.Header().Set("Content-Type", "text/json")

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 3 {
		http.Error(w, "path expected to be /BUCKET/JOBID", http.StatusBadRequest)
		return
	}

  max, err := strconv.ParseUint(r.URL.Query().Get("max"), 10, 0)
  if err != nil {
    max = math.MaxUint64
  }
  
	feed := feeds.join(parts[1], parts[2])

	enc := json.NewEncoder(w)

	delivered := 0
	for content := range feed.contents {
		if err := enc.Encode(content); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		delivered += 1

    max -= 1
    if max <= 0 {
      break
    }
	}

	if delivered == 0 && feed.msg != "" {
		http.Error(w, feed.msg, http.StatusGone)
	}
}

func main() {
  flag.Parse()
	log.Fatal(http.ListenAndServe(*listen, http.HandlerFunc(consume)))
}
