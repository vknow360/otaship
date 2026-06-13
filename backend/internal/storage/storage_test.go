package storage

import (
	"os"
	"testing"
)

func TestNewS3Provider_MissingEnv(t *testing.T) {
	os.Clearenv()
	_, err := NewS3Provider()
	if err == nil {
		t.Errorf("Expected error when S3 credentials are missing")
	}
}

func TestNewS3Provider_WithEnv(t *testing.T) {
	os.Clearenv()
	os.Setenv("S3_ACCESS_KEY", "dummy")
	os.Setenv("S3_SECRET_ACCESS_KEY", "dummy")
	os.Setenv("S3_REGION", "us-east-1")
	os.Setenv("S3_BUCKET_NAME", "mybucket")

	p, err := NewS3Provider()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if p.Name() != "s3" {
		t.Errorf("Expected name 's3', got %s", p.Name())
	}
}

func TestNewCloudinaryProvider_MissingEnv(t *testing.T) {
	os.Clearenv()
	_, err := NewCloudinaryProvider()
	if err == nil {
		t.Errorf("Expected error when Cloudinary credentials are missing")
	}
}

func TestNewCloudinaryProvider_WithEnv(t *testing.T) {
	os.Clearenv()
	os.Setenv("CLOUDINARY_CLOUD_NAME", "dummy")
	os.Setenv("CLOUDINARY_API_KEY", "12345")
	os.Setenv("CLOUDINARY_API_SECRET", "dummy")

	p, err := NewCloudinaryProvider()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if p.Name() != "cloudinary" {
		t.Errorf("Expected name 'cloudinary', got %s", p.Name())
	}
}
