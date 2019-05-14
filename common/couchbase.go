package common

import (
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/couchbase/gocb.v1"
)

func OpenBucket(connectionString string) *gocb.Bucket {
	result, err := url.Parse(connectionString)
	if err != nil {
		panic(errors.Wrapf(err, "initialize meta client failed: %s", connectionString))
	}
	bucketName := strings.Trim(result.Path, "/")
	user := result.User
	result.Scheme = "http"
	result.User = nil
	result.Path = ""

	cluster, err := gocb.Connect(result.String())
	if err != nil {
		panic(err)
	}
	if password, passwordSet := user.Password(); passwordSet {
		err = cluster.Authenticate(gocb.PasswordAuthenticator{
			Username: user.Username(),
			Password: password,
		})
		if err != nil {
			panic(err)
		}
	}

	bucket, err := cluster.OpenBucket(bucketName, "")
	if err != nil {
		panic(err)
	}
	return bucket
}
