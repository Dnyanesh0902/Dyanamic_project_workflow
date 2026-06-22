package util

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func GetAwsSession() (*session.Session, error) {
	secretName := os.Getenv("AWS_BUCKET_SECRET_CODE")
	region := os.Getenv("AWS_DEFAULT_REGION")
	config, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		logrus.Error("@GetAwsSession error in config ", err)
		return nil, err
	}

	svc := secretsmanager.NewFromConfig(config)
	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String("AWSCURRENT"),
	}

	result, err := svc.GetSecretValue(context.TODO(), input)
	if err != nil {
		logrus.Error("@GetAwsSession error in get secret ", err)
		return nil, err
	}
	var secretString string = *result.SecretString
	var awsDetails map[string]interface{}
	json.Unmarshal([]byte(secretString), &awsDetails)

	var accessKey, secretKey interface{}
	if awsDetails["AccessKeyId"] != nil {
		accessKey = awsDetails["AccessKeyId"]
	}
	if awsDetails["SecretAccessKey"] != nil {
		secretKey = awsDetails["SecretAccessKey"]
	}
	session, newSessionErr := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(fmt.Sprintf("%v", accessKey), fmt.Sprintf("%v", secretKey), ""),
	})
	if newSessionErr != nil {
		logrus.Error("@GetAwsSession newSessionErr ", newSessionErr, session)
		return nil, newSessionErr
	}
	return session, nil
}

func StoreToS3WithKMS(storedFilePath string, s3filePath string) (string, error) {
	awsBucketName := os.Getenv("AWS_BUCKET")
	// fmt.Println("🚀 ~ funcStoreToS3WithKMS ~ awsBucketName:", awsBucketName)
	logrus.Info("Local Storage Stored File Path:", storedFilePath)
	logrus.Info("File path to store in S3:", s3filePath)
	file, err := os.Open(storedFilePath)
	logrus.Info("File Opened Successfully")
	if err != nil {
		logrus.Error("UploadStoreToS3@Failed To Open the File, May be It is Corrupted")
		return "", fmt.Errorf("Failed To Open the File, May be It is Corrupted: %v", err)
	}
	defer file.Close()
	session, awsErr := GetAwsSession()
	fmt.Println("🚀 ~ funcStoreToS3WithKMS ~ session:", session)
	if awsErr != nil {
		logrus.Error("Failed to Get AWS Session ", awsErr)
		return "", awsErr
	}

	buffer, err := os.ReadFile(storedFilePath)
	if err != nil {
		logrus.Fatal(err)
	}
	uploader := s3manager.NewUploader(session)
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(awsBucketName),
		Key:         aws.String(s3filePath),
		Body:        file,
		ContentType: aws.String(http.DetectContentType(buffer)),
	})
	if err != nil {
		logrus.Error("S3 Upload Error ", err)
		return "", fmt.Errorf("S3 Upload Error: %v", err)
	}
	logrus.Info("S3 Upload Success ", result)
	return s3filePath, nil
}

func UploadFileToS3WithoutKMS(storedFilePath string, s3filePath string) (string, error) {
	awsRegion := os.Getenv("AWS_DEFAULT_REGION")
	awsAccessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	awsSecretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	awsBucket := os.Getenv("AWS_BUCKET")
	if awsRegion == "" || awsAccessKey == "" || awsSecretKey == "" || awsBucket == "" {
		log.Println("AWS environment variables are not set")
		return "", fmt.Errorf("missing AWS S3 environment variables")
	}
	file, err := os.Open(storedFilePath)
	if err != nil {
		fmt.Println("UploadStoreToS3@Failed To Open the File, May be It is Corrupted")
		return "", fmt.Errorf("Failed To Open the File, May be It is Corrupted: %v", err)
	}
	defer file.Close()
	logrus.Info("File Opened Successfully")
	awsSession, err := session.NewSession(&aws.Config{
		Region:      aws.String(awsRegion),
		Credentials: credentials.NewStaticCredentials(awsAccessKey, awsSecretKey, ""),
	})
	if err != nil {
		log.Printf("Error creating AWS session: %v", err)
		return "", fmt.Errorf("failed to create AWS session: %v", err)
	}
	svc := s3.New(awsSession)
	buffer, err := os.ReadFile(storedFilePath)
	if err != nil {
		log.Fatal(err)
	}
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(awsBucket),
		Key:         aws.String(s3filePath),
		Body:        nil,
		ContentType: aws.String(http.DetectContentType(buffer)),
	})

	if err != nil {
		logrus.Printf("Error uploading to S3: %v", err)
		return "", fmt.Errorf("failed to upload file to S3: %v", err)
	}

	logrus.Println("File uploaded successfully to S3")
	return s3filePath, nil
}

func SanitizeDocumentName(docName string) string {
	regExp := regexp.MustCompile(`[\x00-\x1F\x7F]`)
	docName = regExp.ReplaceAllString(docName, "")
	docName = regexp.MustCompile(`[^A-Za-z0-9.\-]`).ReplaceAllString(docName, " ")
	docName = strings.ReplaceAll(docName, "-", "")
	docName = strings.ReplaceAll(docName, ",", "")
	docName = strings.ReplaceAll(docName, " ", "_")
	return docName
}

func UploadFileToS3(c *gin.Context, fileHeader *multipart.FileHeader, uploadFolder string) (string, error) {
	env := os.Getenv("APP_ENV")
	files := []*multipart.FileHeader{fileHeader}

	// Generate unique filename
	originalFileName := strings.ReplaceAll(fileHeader.Filename, " ", "_")
	fileNameWithoutExt := strings.TrimSuffix(originalFileName, filepath.Ext(originalFileName))
	docName := SanitizeDocName(fileNameWithoutExt)
	uniqueFileName := fmt.Sprintf("%v_%s%s", MakeTimestampMilli(), docName, filepath.Ext(originalFileName))

	if env == "uat" {
		s3Path := filepath.Join(env, uploadFolder, uniqueFileName)

		file, err := fileHeader.Open()
		if err != nil {
			return "", fmt.Errorf("failed to open file: %w", err)
		}
		defer file.Close()

		uploaded, err := UploadToFileToS3(s3Path, file)
		if err != nil {
			logrus.Error("Error uploading file to S3: ", err)
			return "", err
		}
		return uploaded, nil
	}

	currentTime := time.Now()
	year := currentTime.Format("2006")
	month := currentTime.Format("01")

	datedPath := filepath.Join(year, month, uploadFolder)
	fullUploadDir := filepath.Join("uploads", datedPath)

	_, err := UploadAndSaveFiles(c, files, fullUploadDir, []string{uniqueFileName})
	if err != nil {
		return "", fmt.Errorf("failed to save file locally: %w", err)
	}
	// Construct local file path
	localFilePath := filepath.Join(fullUploadDir, uniqueFileName)

	// Remove the local file after successful save
	if removeErr := RemoveLocalStorageFile(localFilePath); removeErr != nil {
		logrus.Warnf("Failed to remove local file %s: %v", localFilePath, removeErr)
	}

	return filepath.Join("uploads", datedPath, uniqueFileName), nil
}

func SaveFileToLocalStorage(c *gin.Context, file *multipart.FileHeader) (string, error) {
	uploadedFile, err := file.Open()
	if err != nil {
		logrus.Fatalf("Error While Opening File:", err)
		return "", err
	}
	defer uploadedFile.Close()

	fileBytes, err := io.ReadAll(uploadedFile)
	if err != nil {
		logrus.Fatalf("Failed To Read File Content:", err)
		return "", err
	}

	fileSize := len(fileBytes)
	fmt.Println("fileSize", fileSize)

	filePath := fmt.Sprintf("uploads/%s", file.Filename)
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		return "", err
	}
	return filePath, nil
}

func RemoveLocalStorageFile(filePath string) error {
	err := os.Remove(filePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func UploadToFileToS3(s3filePath string, file multipart.File) (string, error) {
	awsBucket := os.Getenv("AWS_BUCKET")
	awsSession, err := GetAwsSession()
	if err != nil {
		logrus.Errorf("Failed to create AWS session: %v", err)
		return "", fmt.Errorf("failed to create AWS session: %w", err)
	}

	contentType := "application/octet-stream"
	switch ext := strings.ToLower(filepath.Ext(s3filePath)); ext {
	case ".png":
		contentType = "image/png"
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".gif":
		contentType = "image/gif"
	case ".pdf":
		contentType = "application/pdf"
	}

	logrus.Infof("Uploading file: %s with content type: %s", s3filePath, contentType)

	uploader := s3manager.NewUploader(awsSession)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(awsBucket),
		Key:         aws.String(s3filePath),
		Body:        file,
		ContentType: aws.String(contentType),
	})

	if err != nil {
		logrus.Errorf("Failed to upload file to S3: %v", err)
		return "", fmt.Errorf("failed to upload file to S3: %w", err)
	}

	return s3filePath, nil
}

func UploadAndSaveFiles(c *gin.Context, files []*multipart.FileHeader, uploadDir string, fileNames []string) ([]string, error) {
	var savedPaths []string

	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %v", err)
	}

	for i, fileHeader := range files {
		fileName := fileNames[i]
		finalPath := filepath.Join(uploadDir, fileName)

		file, err := fileHeader.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %v", err)
		}
		defer file.Close()

		out, err := os.Create(finalPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create file: %v", err)
		}
		defer out.Close()

		if _, err = io.Copy(out, file); err != nil {
			return nil, fmt.Errorf("failed to save file: %v", err)
		}

		savedPaths = append(savedPaths, finalPath)
	}

	return savedPaths, nil
}

func UploadFileToS3WithKMS(c *gin.Context, file *multipart.FileHeader, s3FolderPath string) (string, string, error) {
	sanitizedFileName := SanitizeDocumentName(file.Filename)
	s3FilePath := fmt.Sprintf("%s/%s", s3FolderPath, sanitizedFileName)
	localFilePath := fmt.Sprintf("%s/%s", "storage", sanitizedFileName)

	uploadedFile, err := file.Open()
	if err != nil {
		logrus.Fatalf("Error While Opening File: %v", err)
		return "", "", err
	}
	defer uploadedFile.Close()

	fileBytes, err := io.ReadAll(uploadedFile)
	if err != nil {
		logrus.Fatalf("Failed To Read File Content: %v", err)
		return "", "", err
	}

	fileSize := len(fileBytes)
	fmt.Println("fileSize", fileSize)

	if err := c.SaveUploadedFile(file, localFilePath); err != nil {
		return "", "", err
	}

	s3UploadedPath, err := StoreToS3WithKMS(localFilePath, s3FilePath)
	if err != nil {
		logrus.Error("Failed to upload file to S3:", err)
		_ = os.Remove(localFilePath)
		return "", "", err
	}

	if err := os.Remove(localFilePath); err != nil {
		logrus.Error("Failed to delete local file after upload:", err)
	}

	return s3UploadedPath, sanitizedFileName, nil
}
