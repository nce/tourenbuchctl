package pdfexport

import (
	"context"
	"fmt"
	"os"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/objectstorage"
)

type S3Export struct {
	Client     objectstorage.ObjectStorageClient
	BucketName string
	ObjectName string
}

func initClient() (objectstorage.ObjectStorageClient, error) {
	provider := common.DefaultConfigProvider()

	client, err := objectstorage.NewObjectStorageClientWithConfigurationProvider(provider)
	if err != nil {
		return objectstorage.ObjectStorageClient{}, fmt.Errorf("failed to init ObjectstorageClient: %w", err)
	}

	return client, nil
}

func (s3 *S3Export) Save(srcFile string) error {
	var err error

	s3.Client, err = initClient()
	if err != nil {
		return err
	}

	ctx := context.Background()

	nsResp, err := s3.Client.GetNamespace(ctx, objectstorage.GetNamespaceRequest{})
	if err != nil {
		return fmt.Errorf("failed to query bucket namespace: %w", err)
	}

	namespace := *nsResp.Value

	file, err := os.Open(srcFile)
	if err != nil {
		return fmt.Errorf("failed to open file: %s; %w", srcFile, err)
	}
	defer file.Close()

	putReq := objectstorage.PutObjectRequest{
		NamespaceName: &namespace,
		BucketName:    &s3.BucketName,
		ObjectName:    &s3.ObjectName,
		PutObjectBody: file,
		//nolint: modernize
		ContentType: common.String("application/pdf"),
	}

	_, err = s3.Client.PutObject(ctx, putReq)
	if err != nil {
		return fmt.Errorf("failed to upload object: %s; %w", s3.ObjectName, err)
	}

	return nil
}
