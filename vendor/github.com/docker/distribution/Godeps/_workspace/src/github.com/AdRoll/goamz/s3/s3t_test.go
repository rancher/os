package s3_test

import (
	"github.com/AdRoll/goamz/aws"
	"github.com/AdRoll/goamz/s3"
	"github.com/AdRoll/goamz/s3/s3test"
	"github.com/AdRoll/goamz/testutil"
	"gopkg.in/check.v1"
)

type LocalServer struct {
	auth   aws.Auth
	region aws.Region
	srv    *s3test.Server
	config *s3test.Config
}

func (s *LocalServer) SetUp(c *check.C) {
	srv, err := s3test.NewServer(s.config)
	c.Assert(err, check.IsNil)
	c.Assert(srv, check.NotNil)

	s.srv = srv
	s.region = aws.Region{
		Name:                 "faux-region-1",
		S3Endpoint:           srv.URL(),
		S3LocationConstraint: true, // s3test server requires a LocationConstraint
	}
}

// LocalServerSuite defines tests that will run
// against the local s3test server. It includes
// selected tests from ClientTests;
// when the s3test functionality is sufficient, it should
// include all of them, and ClientTests can be simply embedded.
type LocalServerSuite struct {
	srv         LocalServer
	clientTests ClientTests
}

var (
	// run tests twice, once in us-east-1 mode, once not.
	_ = check.Suite(&LocalServerSuite{})
	_ = check.Suite(&LocalServerSuite{
		srv: LocalServer{
			config: &s3test.Config{
				Send409Conflict: true,
			},
		},
	})
)

func (s *LocalServerSuite) SetUpSuite(c *check.C) {
	s.srv.SetUp(c)
	s.clientTests.s3 = s3.New(s.srv.auth, s.srv.region)

	// TODO Sadly the fake server ignores auth completely right now. :-(
	s.clientTests.authIsBroken = true
	s.clientTests.Cleanup()
}

func (s *LocalServerSuite) TearDownTest(c *check.C) {
	s.clientTests.Cleanup()
}

func (s *LocalServerSuite) TestBasicFunctionality(c *check.C) {
	s.clientTests.TestBasicFunctionality(c)
}

func (s *LocalServerSuite) TestGetNotFound(c *check.C) {
	s.clientTests.TestGetNotFound(c)
}

func (s *LocalServerSuite) TestBucketList(c *check.C) {
	s.clientTests.TestBucketList(c)
}

func (s *LocalServerSuite) TestDoublePutBucket(c *check.C) {
	s.clientTests.TestDoublePutBucket(c)
}

func (s *LocalServerSuite) TestMultiComplete(c *check.C) {
	if !testutil.Amazon {
		c.Skip("live tests against AWS disabled (no -amazon)")
	}
	s.clientTests.TestMultiComplete(c)
}
