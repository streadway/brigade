# Brigade

Demultiplexes an S3 bucket listing across multiple consumers in order to parallelize work.

# Usage

./brigade --http :8080

# Requests

Continuation of a stream, can be performed any number of times, no retry or
acknowledgment is possible.

All items are delivered when the transport is closed.

A new listing will begin with every unique job.

```
GET /BUCKET/JOB?max=10000 HTTP/1.1
Host: brigade

HTTP/1.1 200 OK
Content-Type: text/json
Transfer-Encoding: chunked

{...}
{...}
{...}x998
```

Stream is complete, no more items to list, does not contain a payload

```
GET /BUCKET/JOB/CLIENT HTTP/1.1
Host: Brigade

HTTP/1.1 200
```

Missing bucket or other failure

```
GET /MISSING-BUCKET/JOB/CLIENT HTTP/1.1
Host: Brigade

HTTP/1.1 410

Descriptive upstream message
```

# Schema

Each streaming JSON object will represent the "Contents" section of the paged
bucket listing, defined in
http://docs.amazonwebservices.com/AmazonS3/latest/API/RESTBucketGET.html

{
  "key": "name-of-object/in/bucket",
  "lastModified": "2010-10-20T21:38:07.000Z",
  "eTag": "\"2bbd1d84df61f9662d859326a6ed0972\"",
  "size": 5024,
}
