package s3_test

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"github.com/AdRoll/goamz/aws"
	"github.com/AdRoll/goamz/s3"
	"github.com/AdRoll/goamz/testutil"
	"gopkg.in/check.v1"
	"io/ioutil"
	"net"
	"net/http"
	"sort"
	"strings"
	"time"
)

// AmazonServer represents an Amazon S3 server.
type AmazonServer struct {
	auth aws.Auth
}

func (s *AmazonServer) SetUp(c *check.C) {
	auth, err := aws.EnvAuth()
	if err != nil {
		c.Fatal(err.Error())
	}
	s.auth = auth
}

var _ = check.Suite(&AmazonClientSuite{Region: aws.USEast})
var _ = check.Suite(&AmazonClientSuite{Region: aws.EUWest})
var _ = check.Suite(&AmazonDomainClientSuite{Region: aws.USEast})

// AmazonClientSuite tests the client against a live S3 server.
type AmazonClientSuite struct {
	aws.Region
	srv AmazonServer
	ClientTests
}

func (s *AmazonClientSuite) SetUpSuite(c *check.C) {
	if !testutil.Amazon {
		c.Skip("live tests against AWS disabled (no -amazon)")
	}
	s.srv.SetUp(c)
	s.s3 = s3.New(s.srv.auth, s.Region)
	// In case tests were interrupted in the middle before.
	s.ClientTests.Cleanup()
}

func (s *AmazonClientSuite) TearDownTest(c *check.C) {
	s.ClientTests.Cleanup()
}

// AmazonDomainClientSuite tests the client against a live S3
// server using bucket names in the endpoint domain name rather
// than the request path.
type AmazonDomainClientSuite struct {
	aws.Region
	srv AmazonServer
	ClientTests
}

func (s *AmazonDomainClientSuite) SetUpSuite(c *check.C) {
	if !testutil.Amazon {
		c.Skip("live tests against AWS disabled (no -amazon)")
	}
	s.srv.SetUp(c)
	region := s.Region
	region.S3BucketEndpoint = "https://${bucket}.s3.amazonaws.com"
	s.s3 = s3.New(s.srv.auth, region)
	s.ClientTests.Cleanup()
}

func (s *AmazonDomainClientSuite) TearDownTest(c *check.C) {
	s.ClientTests.Cleanup()
}

// ClientTests defines integration tests designed to test the client.
// It is not used as a test suite in itself, but embedded within
// another type.
type ClientTests struct {
	s3           *s3.S3
	authIsBroken bool
}

func (s *ClientTests) Cleanup() {
	killBucket(testBucket(s.s3))
}

func testBucket(s *s3.S3) *s3.Bucket {
	// Watch out! If this function is corrupted and made to match with something
	// people own, killBucket will happily remove *everything* inside the bucket.
	key := s.Auth.AccessKey
	if len(key) >= 8 {
		key = s.Auth.AccessKey[:8]
	}
	return s.Bucket(fmt.Sprintf("goamz-%s-%s", s.Region.Name, key))
}

var attempts = aws.AttemptStrategy{
	Min:   5,
	Total: 20 * time.Second,
	Delay: 100 * time.Millisecond,
}

func killBucket(b *s3.Bucket) {
	var err error
	for attempt := attempts.Start(); attempt.Next(); {
		err = b.DelBucket()
		if err == nil {
			return
		}
		if _, ok := err.(*net.DNSError); ok {
			return
		}
		e, ok := err.(*s3.Error)
		if ok && e.Code == "NoSuchBucket" {
			return
		}
		if ok && e.Code == "BucketNotEmpty" {
			// Errors are ignored here. Just retry.
			resp, err := b.List("", "", "", 1000)
			if err == nil {
				for _, key := range resp.Contents {
					_ = b.Del(key.Key)
				}
			}
			multis, _, _ := b.ListMulti("", "")
			for _, m := range multis {
				_ = m.Abort()
			}
		}
	}
	message := "cannot delete test bucket"
	if err != nil {
		message += ": " + err.Error()
	}
	panic(message)
}

func get(url string) ([]byte, error) {
	for attempt := attempts.Start(); attempt.Next(); {
		resp, err := http.Get(url)
		if err != nil {
			if attempt.HasNext() {
				continue
			}
			return nil, err
		}
		data, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			if attempt.HasNext() {
				continue
			}
			return nil, err
		}
		return data, err
	}
	panic("unreachable")
}

func (s *ClientTests) TestBasicFunctionality(c *check.C) {
	b := testBucket(s.s3)
	err := b.PutBucket(s3.PublicRead)
	c.Assert(err, check.IsNil)

	err = b.Put("name", []byte("yo!"), "text/plain", s3.PublicRead, s3.Options{})
	c.Assert(err, check.IsNil)
	defer b.Del("name")

	data, err := b.Get("name")
	c.Assert(err, check.IsNil)
	c.Assert(string(data), check.Equals, "yo!")

	data, err = get(b.URL("name"))
	c.Assert(err, check.IsNil)
	c.Assert(string(data), check.Equals, "yo!")

	buf := bytes.NewBufferString("hey!")
	err = b.PutReader("name2", buf, int64(buf.Len()), "text/plain", s3.Private, s3.Options{})
	c.Assert(err, check.IsNil)
	defer b.Del("name2")

	rc, err := b.GetReader("name2")
	c.Assert(err, check.IsNil)
	data, err = ioutil.ReadAll(rc)
	c.Check(err, check.IsNil)
	c.Check(string(data), check.Equals, "hey!")
	rc.Close()

	data, err = get(b.SignedURL("name2", time.Now().Add(time.Hour)))
	c.Assert(err, check.IsNil)
	c.Assert(string(data), check.Equals, "hey!")

	if !s.authIsBroken {
		data, err = get(b.SignedURL("name2", time.Now().Add(-time.Hour)))
		c.Assert(err, check.IsNil)
		c.Assert(string(data), check.Matches, "(?s).*AccessDenied.*")
	}

	err = b.DelBucket()
	c.Assert(err, check.NotNil)

	s3err, ok := err.(*s3.Error)
	c.Assert(ok, check.Equals, true)
	c.Assert(s3err.Code, check.Equals, "BucketNotEmpty")
	c.Assert(s3err.BucketName, check.Equals, b.Name)
	c.Assert(s3err.Message, check.Equals, "The bucket you tried to delete is not empty")

	err = b.Del("name")
	c.Assert(err, check.IsNil)
	err = b.Del("name2")
	c.Assert(err, check.IsNil)

	err = b.DelBucket()
	c.Assert(err, check.IsNil)
}

func (s *ClientTests) TestGetNotFound(c *check.C) {
	b := s.s3.Bucket("goamz-" + s.s3.Auth.AccessKey)
	data, err := b.Get("non-existent")

	s3err, _ := err.(*s3.Error)
	c.Assert(s3err, check.NotNil)
	c.Assert(s3err.StatusCode, check.Equals, 404)
	c.Assert(s3err.Code, check.Equals, "NoSuchBucket")
	c.Assert(s3err.Message, check.Equals, "The specified bucket does not exist")
	c.Assert(data, check.IsNil)
}

// Communicate with all endpoints to see if they are alive.
func (s *ClientTests) TestRegions(c *check.C) {
	errs := make(chan error, len(aws.Regions))
	for _, region := range aws.Regions {
		go func(r aws.Region) {
			s := s3.New(s.s3.Auth, r)
			b := s.Bucket("goamz-" + s.Auth.AccessKey)
			_, err := b.Get("non-existent")
			errs <- err
		}(region)
	}
	for _ = range aws.Regions {
		err := <-errs
		if err != nil {
			s3_err, ok := err.(*s3.Error)
			if ok {
				c.Check(s3_err.Code, check.Matches, "NoSuchBucket")
			} else if _, ok = err.(*net.DNSError); ok {
				// Okay as well.
			} else {
				c.Errorf("Non-S3 error: %s", err)
			}
		} else {
			c.Errorf("Test should have errored but it seems to have succeeded")
		}
	}
}

var objectNames = []string{
	"index.html",
	"index2.html",
	"photos/2006/February/sample2.jpg",
	"photos/2006/February/sample3.jpg",
	"photos/2006/February/sample4.jpg",
	"photos/2006/January/sample.jpg",
	"test/bar",
	"test/foo",
}

func keys(names ...string) []s3.Key {
	ks := make([]s3.Key, len(names))
	for i, name := range names {
		ks[i].Key = name
	}
	return ks
}

// As the ListResp specifies all the parameters to the
// request too, we use it to specify request parameters
// and expected results. The Contents field is
// used only for the key names inside it.
var listTests = []s3.ListResp{
	// normal list.
	{
		Contents: keys(objectNames...),
	}, {
		Marker:   objectNames[0],
		Contents: keys(objectNames[1:]...),
	}, {
		Marker:   objectNames[0] + "a",
		Contents: keys(objectNames[1:]...),
	}, {
		Marker: "z",
	},

	// limited results.
	{
		MaxKeys:     2,
		Contents:    keys(objectNames[0:2]...),
		IsTruncated: true,
	}, {
		MaxKeys:     2,
		Marker:      objectNames[0],
		Contents:    keys(objectNames[1:3]...),
		IsTruncated: true,
	}, {
		MaxKeys:  2,
		Marker:   objectNames[len(objectNames)-2],
		Contents: keys(objectNames[len(objectNames)-1:]...),
	},

	// with delimiter
	{
		Delimiter:      "/",
		CommonPrefixes: []string{"photos/", "test/"},
		Contents:       keys("index.html", "index2.html"),
	}, {
		Delimiter:      "/",
		Prefix:         "photos/2006/",
		CommonPrefixes: []string{"photos/2006/February/", "photos/2006/January/"},
	}, {
		Delimiter:      "/",
		Prefix:         "t",
		CommonPrefixes: []string{"test/"},
	}, {
		Delimiter:   "/",
		MaxKeys:     1,
		Contents:    keys("index.html"),
		IsTruncated: true,
	}, {
		Delimiter:      "/",
		MaxKeys:        1,
		Marker:         "index2.html",
		CommonPrefixes: []string{"photos/"},
		IsTruncated:    true,
	}, {
		Delimiter:      "/",
		MaxKeys:        1,
		Marker:         "photos/",
		CommonPrefixes: []string{"test/"},
		IsTruncated:    false,
	}, {
		Delimiter:      "Feb",
		CommonPrefixes: []string{"photos/2006/Feb"},
		Contents:       keys("index.html", "index2.html", "photos/2006/January/sample.jpg", "test/bar", "test/foo"),
	},
}

func (s *ClientTests) TestDoublePutBucket(c *check.C) {
	b := testBucket(s.s3)
	err := b.PutBucket(s3.PublicRead)
	c.Assert(err, check.IsNil)

	err = b.PutBucket(s3.PublicRead)
	if err != nil {
		c.Assert(err, check.FitsTypeOf, new(s3.Error))
		c.Assert(err.(*s3.Error).Code, check.Equals, "BucketAlreadyOwnedByYou")
	}
}

func (s *ClientTests) TestBucketList(c *check.C) {
	b := testBucket(s.s3)
	err := b.PutBucket(s3.Private)
	c.Assert(err, check.IsNil)

	objData := make(map[string][]byte)
	for i, path := range objectNames {
		data := []byte(strings.Repeat("a", i))
		err := b.Put(path, data, "text/plain", s3.Private, s3.Options{})
		c.Assert(err, check.IsNil)
		defer b.Del(path)
		objData[path] = data
	}

	for i, t := range listTests {
		c.Logf("test %d", i)
		resp, err := b.List(t.Prefix, t.Delimiter, t.Marker, t.MaxKeys)
		c.Assert(err, check.IsNil)
		c.Check(resp.Name, check.Equals, b.Name)
		c.Check(resp.Delimiter, check.Equals, t.Delimiter)
		c.Check(resp.IsTruncated, check.Equals, t.IsTruncated)
		c.Check(resp.CommonPrefixes, check.DeepEquals, t.CommonPrefixes)
		checkContents(c, resp.Contents, objData, t.Contents)
	}
}

func etag(data []byte) string {
	sum := md5.New()
	sum.Write(data)
	return fmt.Sprintf(`"%x"`, sum.Sum(nil))
}

func checkContents(c *check.C, contents []s3.Key, data map[string][]byte, expected []s3.Key) {
	c.Assert(contents, check.HasLen, len(expected))
	for i, k := range contents {
		c.Check(k.Key, check.Equals, expected[i].Key)
		// TODO mtime
		c.Check(k.Size, check.Equals, int64(len(data[k.Key])))
		c.Check(k.ETag, check.Equals, etag(data[k.Key]))
	}
}

func (s *ClientTests) TestMultiInitPutList(c *check.C) {
	b := testBucket(s.s3)
	err := b.PutBucket(s3.Private)
	c.Assert(err, check.IsNil)

	multi, err := b.InitMulti("multi", "text/plain", s3.Private, s3.Options{})
	c.Assert(err, check.IsNil)
	c.Assert(multi.UploadId, check.Matches, ".+")
	defer multi.Abort()

	var sent []s3.Part

	for i := 0; i < 5; i++ {
		p, err := multi.PutPart(i+1, strings.NewReader(fmt.Sprintf("<part %d>", i+1)))
		c.Assert(err, check.IsNil)
		c.Assert(p.N, check.Equals, i+1)
		c.Assert(p.Size, check.Equals, int64(8))
		c.Assert(p.ETag, check.Matches, ".+")
		sent = append(sent, p)
	}

	s3.SetListPartsMax(2)

	parts, err := multi.ListParts()
	c.Assert(err, check.IsNil)
	c.Assert(parts, check.HasLen, len(sent))
	for i := range parts {
		c.Assert(parts[i].N, check.Equals, sent[i].N)
		c.Assert(parts[i].Size, check.Equals, sent[i].Size)
		c.Assert(parts[i].ETag, check.Equals, sent[i].ETag)
	}

	err = multi.Complete(parts)
	s3err, failed := err.(*s3.Error)
	c.Assert(failed, check.Equals, true)
	c.Assert(s3err.Code, check.Equals, "EntityTooSmall")

	err = multi.Abort()
	c.Assert(err, check.IsNil)
	_, err = multi.ListParts()
	s3err, ok := err.(*s3.Error)
	c.Assert(ok, check.Equals, true)
	c.Assert(s3err.Code, check.Equals, "NoSuchUpload")
}

// This may take a minute or more due to the minimum size accepted S3
// on multipart upload parts.
func (s *ClientTests) TestMultiComplete(c *check.C) {
	b := testBucket(s.s3)
	err := b.PutBucket(s3.Private)
	c.Assert(err, check.IsNil)

	contentType := "text/plain"
	meta := make(map[string][]string)
	meta["X-Amz-Meta-TestField"] = []string{"testValue"}
	options := s3.Options{ContentEncoding: "identity", ContentDisposition: "inline", Meta: meta}
	multi, err := b.InitMulti("multi", contentType, s3.Private, options)
	c.Assert(err, check.IsNil)
	c.Assert(multi.UploadId, check.Matches, ".+")
	defer multi.Abort()

	// Minimum size S3 accepts for all but the last part is 5MB.
	data1 := make([]byte, 5*1024*1024)
	data2 := []byte("<part 2>")

	part1, err := multi.PutPart(1, bytes.NewReader(data1))
	c.Assert(err, check.IsNil)
	part2, err := multi.PutPart(2, bytes.NewReader(data2))
	c.Assert(err, check.IsNil)

	// Purposefully reversed. The order requirement must be handled.
	err = multi.Complete([]s3.Part{part2, part1})
	c.Assert(err, check.IsNil)

	data, err := b.Get("multi")
	c.Assert(err, check.IsNil)

	c.Assert(len(data), check.Equals, len(data1)+len(data2))
	for i := range data1 {
		if data[i] != data1[i] {
			c.Fatalf("uploaded object at byte %d: want %d, got %d", data1[i], data[i])
		}
	}
	c.Assert(string(data[len(data1):]), check.Equals, string(data2))

	resp, err := b.GetResponse("multi")
	c.Assert(resp.Header.Get("Content-Type"), check.Equals, contentType)
	c.Assert(resp.Header.Get("x-amz-acl"), check.Equals, s3.Private)
	c.Assert(resp.Header.Get("Content-MD5"), check.Equals, options.ContentMD5)
	c.Assert(resp.Header.Get("Content-Encoding"), check.Equals, options.ContentEncoding)
	c.Assert(resp.Header.Get("Content-Disposition"), check.Equals, options.ContentDisposition)
	for k, values := range meta {
		c.Assert(resp.Header.Get(k), check.Equals, strings.Join(values, ","))
	}
}

type multiList []*s3.Multi

func (l multiList) Len() int           { return len(l) }
func (l multiList) Less(i, j int) bool { return l[i].Key < l[j].Key }
func (l multiList) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }

func (s *ClientTests) TestListMulti(c *check.C) {
	b := testBucket(s.s3)
	err := b.PutBucket(s3.Private)
	c.Assert(err, check.IsNil)

	// Ensure an empty state before testing its behavior.
	multis, _, err := b.ListMulti("", "")
	for _, m := range multis {
		err := m.Abort()
		c.Assert(err, check.IsNil)
	}

	keys := []string{
		"a/multi2",
		"a/multi3",
		"b/multi4",
		"multi1",
	}
	for _, key := range keys {
		m, err := b.InitMulti(key, "", s3.Private, s3.Options{})
		c.Assert(err, check.IsNil)
		defer m.Abort()
	}

	// Amazon's implementation of the multiple-request listing for
	// multipart uploads in progress seems broken in multiple ways.
	// (next tokens are not provided, etc).
	//s3.SetListMultiMax(2)

	multis, prefixes, err := b.ListMulti("", "")
	c.Assert(err, check.IsNil)
	for attempt := attempts.Start(); attempt.Next() && len(multis) < len(keys); {
		multis, prefixes, err = b.ListMulti("", "")
		c.Assert(err, check.IsNil)
	}
	sort.Sort(multiList(multis))
	c.Assert(prefixes, check.IsNil)
	var gotKeys []string
	for _, m := range multis {
		gotKeys = append(gotKeys, m.Key)
	}
	c.Assert(gotKeys, check.DeepEquals, keys)
	for _, m := range multis {
		c.Assert(m.Bucket, check.Equals, b)
		c.Assert(m.UploadId, check.Matches, ".+")
	}

	multis, prefixes, err = b.ListMulti("", "/")
	for attempt := attempts.Start(); attempt.Next() && len(prefixes) < 2; {
		multis, prefixes, err = b.ListMulti("", "")
		c.Assert(err, check.IsNil)
	}
	c.Assert(err, check.IsNil)
	c.Assert(prefixes, check.DeepEquals, []string{"a/", "b/"})
	c.Assert(multis, check.HasLen, 1)
	c.Assert(multis[0].Bucket, check.Equals, b)
	c.Assert(multis[0].Key, check.Equals, "multi1")
	c.Assert(multis[0].UploadId, check.Matches, ".+")

	for attempt := attempts.Start(); attempt.Next() && len(multis) < 2; {
		multis, prefixes, err = b.ListMulti("", "")
		c.Assert(err, check.IsNil)
	}
	multis, prefixes, err = b.ListMulti("a/", "/")
	c.Assert(err, check.IsNil)
	c.Assert(prefixes, check.IsNil)
	c.Assert(multis, check.HasLen, 2)
	c.Assert(multis[0].Bucket, check.Equals, b)
	c.Assert(multis[0].Key, check.Equals, "a/multi2")
	c.Assert(multis[0].UploadId, check.Matches, ".+")
	c.Assert(multis[1].Bucket, check.Equals, b)
	c.Assert(multis[1].Key, check.Equals, "a/multi3")
	c.Assert(multis[1].UploadId, check.Matches, ".+")
}

func (s *ClientTests) TestMultiPutAllZeroLength(c *check.C) {
	b := testBucket(s.s3)
	err := b.PutBucket(s3.Private)
	c.Assert(err, check.IsNil)

	multi, err := b.InitMulti("multi", "text/plain", s3.Private, s3.Options{})
	c.Assert(err, check.IsNil)
	defer multi.Abort()

	// This tests an edge case. Amazon requires at least one
	// part for multiprat uploads to work, even the part is empty.
	parts, err := multi.PutAll(strings.NewReader(""), 5*1024*1024)
	c.Assert(err, check.IsNil)
	c.Assert(parts, check.HasLen, 1)
	c.Assert(parts[0].Size, check.Equals, int64(0))
	c.Assert(parts[0].ETag, check.Equals, `"d41d8cd98f00b204e9800998ecf8427e"`)

	err = multi.Complete(parts)
	c.Assert(err, check.IsNil)
}
