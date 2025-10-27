package minioClient

import (
	"context"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Client - это наша обертка над стандартным клиентом MinIO
type Client struct {
	*minio.Client
	bucketName string
}

// New создает и настраивает новый клиент для MinIO.
func New() (*Client, error) {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	secretKey := os.Getenv("MINIO_SECRET_KEY")
	useSSL, _ := strconv.ParseBool(os.Getenv("MINIO_USE_SSL"))
	bucketName := os.Getenv("MINIO_BUCKET_NAME")

	if endpoint == "" || accessKey == "" || secretKey == "" || bucketName == "" {
		return nil, fmt.Errorf("minio environment variables are not fully set")
	}

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	exists, err := minioClient.BucketExists(ctx, bucketName)
	if err != nil {
		return nil, err
	}
	if !exists {
		err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, err
		}
		// Устанавливаем публичную политику для бакета
		policy := fmt.Sprintf(`{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"AWS":["*"]},"Action":["s3:GetObject"],"Resource":["arn:aws:s3:::%s/*"]}]}`, bucketName)
		err = minioClient.SetBucketPolicy(ctx, bucketName, policy)
		if err != nil {
			return nil, err
		}
	}

	return &Client{
		Client:     minioClient,
		bucketName: bucketName,
	}, nil
}

// UploadImage загружает файл изображения в MinIO.
// Имя объекта генерируется уникально.
func (c *Client) UploadImage(ctx context.Context, file *multipart.FileHeader) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	// Генерируем уникальное имя для объекта
	objectName := uuid.New().String() + filepath.Ext(file.Filename)

	_, err = c.PutObject(ctx, c.bucketName, objectName, src, file.Size, minio.PutObjectOptions{
		ContentType: file.Header.Get("Content-Type"),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload object: %w", err)
	}

	return objectName, nil
}

// DeleteImage удаляет объект из MinIO.
func (c *Client) DeleteImage(ctx context.Context, objectName string) error {
	if objectName == "" {
		return nil // Нечего удалять
	}
	return c.RemoveObject(ctx, c.bucketName, objectName, minio.RemoveObjectOptions{})
}

// GetImageURL возвращает полный URL к объекту в MinIO
func (c *Client) GetImageURL(objectName string) string {
	if objectName == "" {
		return ""
	}
	// Примечание: Убедитесь, что ваш MinIO доступен по этому адресу извне
	return fmt.Sprintf("http://%s/%s/%s", os.Getenv("MINIO_ENDPOINT"), c.bucketName, objectName)
}
