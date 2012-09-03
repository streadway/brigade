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

# TODO

 * Add AWS authentication

# LICENSE

Copyright (C) 2012 Sean Treadway <http://github.com/streadway>, SoundCloud Ltd.

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
of the Software, and to permit persons to whom the Software is furnished to do
so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
