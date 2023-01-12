package enumerator

import (
	"context"
	"fmt"
	"path"
	"strings"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/pkg/errors"
	"github.com/snyk/driftctl/pkg/iac/config"
)

type GSEnumerator struct {
	config config.SupplierConfig
	client storage.Client
}

func NewGSEnumerator(config config.SupplierConfig) (*GSEnumerator, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, errors.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	return &GSEnumerator{
		config,
		*client,
	}, nil
}

func (s *GSEnumerator) Origin() string {
	return s.config.String()
}

func (s *GSEnumerator) Enumerate() ([]string, error) {
	bucketPath := strings.Split(s.config.Path, "/")
	if len(bucketPath) < 2 {
		return nil, fmt.Errorf("unable to parse GS path: %s. Must be BUCKET_NAME/PREFIX", s.config.Path)
	}

	bucketName := bucketPath[0]
	// prefix should contains everything that does not have a glob pattern
	// Pattern should be the glob matcher string
	prefix, pattern := extractPrefixAndPattern(strings.Join(bucketPath[1:], "/"))

	// We combine the prefix and pattern to match file names against.
	fullPattern := path.Join(prefix, pattern)

	files := make([]string, 0)

	bucket := s.client.Bucket(bucketName)

	it := bucket.Objects(context.Background(), &storage.Query{})
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		if attrs.Size == 0 {
			continue
		}
		if attrs.Size > 0 {
			if match, _ := doublestar.Match(fullPattern, attrs.Name); match {
				files = append(files, strings.Join([]string{bucketPath[0], attrs.Name}, "/"))
			}
		}
	}

	if len(files) == 0 {
		return files, fmt.Errorf("no Terraform state was found in %s, exiting", s.config.Path)
	}

	return files, nil
}
