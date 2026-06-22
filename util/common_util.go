package util

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"

	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/sirupsen/logrus"
)

func RequestBodyLogger(c *gin.Context) string {
	requestBody, _ := c.GetRawData()
	c.Request.Body = io.NopCloser(bytes.NewReader(requestBody))
	return string(requestBody)
}

func GetServerHostName() string {
	hostname, err := os.Hostname()
	if err != nil {
		logrus.Warn("Failed to retrieve hostname: ", err)
		return "unknown-host"
	}
	return hostname
}

func InArray(item string, array []string, caseSensitive bool) bool {
	for _, element := range array {
		if (caseSensitive && element == item) || (!caseSensitive && strings.EqualFold(element, item)) {
			return true
		}
	}
	return false
}

func GetPaginationParams(c *gin.Context) (int, int, error) {
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		return 0, 0, err
	}

	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		return 0, 0, err
	}

	return limit, offset, nil
}

func GenerateSluge(input string) string {
	var slugBuilder strings.Builder
	for _, r := range input {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			slugBuilder.WriteRune(unicode.ToLower(r))
		} else if unicode.IsSpace(r) {
			slugBuilder.WriteRune('_')
		}
	}
	return slugBuilder.String()
}

func GetFilePathAndName(Clienttxnid string) (string, string) {
	// Base directory where the files will be saved (you can change this as needed)
	baseDir := "/path/to/save/pdf" // Modify the base directory accordingly

	// Create a unique file name based on Clienttxnid and the current timestamp
	timestamp := time.Now().Format("20060102_150405")
	fileName := fmt.Sprintf("investor_%d_%s.pdf", Clienttxnid, timestamp)

	// Generate the full file path by combining the base directory and file name
	filePath := filepath.Join(baseDir, fileName)

	return filePath, fileName
}

func DecryptAndSavePdf(pdfBase64 string, filePath string, fileName string) error {
	// Decode the base64 string into a byte array
	decodedData, err := base64.StdEncoding.DecodeString(pdfBase64)
	if err != nil {
		return errors.New("failed to decode base64 string")
	}

	// Optionally, you can add decryption logic here if the PDF is encrypted

	// Write the decoded data to the file at the specified path
	err = ioutil.WriteFile(filePath, decodedData, 0644)
	if err != nil {
		return errors.New("failed to save the PDF file")
	}

	// Return nil if everything succeeds
	return nil
}

func EncryptAES(plainText string) (string, error) {
	key := os.Getenv("AES_KEY_FOR_AADHAR_PAN")
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	cipherText := aesGCM.Seal(nil, nonce, []byte(plainText), nil)
	result := append(nonce, cipherText...)

	return base64.StdEncoding.EncodeToString(result), nil
}

func DecryptAES(encryptedText string) (string, error) {
	key := os.Getenv("AES_KEY_FOR_AADHAR_PAN")
	data, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	if len(data) < 12 {
		return "", errors.New("invalid ciphertext")
	}

	nonce, cipherText := data[:12], data[12:]

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	plainText, err := aesGCM.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return "", err
	}

	return string(plainText), nil
}

func GenerateHMACBase64(message string) (string, error) {

	secretKey := os.Getenv("HMAC_SECRET_KEY")
	if secretKey == "" {
		logrus.Error("HMAC Secret key not set in environment variables")
		return "", errors.New("HMAC secret key not set in environment variables")
	}

	key := []byte(secretKey)
	data := []byte(message)
	h := hmac.New(sha512.New, key)
	_, err := h.Write(data)
	if err != nil {
		return "", fmt.Errorf("error writing data to HMAC: %v", err)
	}
	hash := h.Sum(nil)
	return base64.StdEncoding.EncodeToString(hash), nil
}

func GenerateUnikCode() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(1e8))
	randomPart := fmt.Sprintf("%08d", n.Int64())

	return "0000" + randomPart + "M"
}

func GetAuthenticatedUserID(c *gin.Context) (int, error) {
	authUserIDValue, exists := c.Get("user_id")
	if !exists {
		logrus.Error("user_id not found in context")
		return 0, fmt.Errorf("unauthorized access")
	}

	authUserIDStr, ok := authUserIDValue.(string)
	if !ok {
		logrus.Error("Invalid user_id format")
		return 0, fmt.Errorf("invalid user ID format")
	}

	authUserID, err := strconv.Atoi(authUserIDStr)
	if err != nil {
		logrus.Error("Error converting user_id to int:", err)
		return 0, fmt.Errorf("invalid user ID format")
	}

	return authUserID, nil
}

func GenerateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	re := regexp.MustCompile(`[^a-z0-9-]`)
	slug = re.ReplaceAllString(slug, "")
	return slug
}

func InArrayInt(item int, array []int) bool {
	var found bool

	for _, num := range array {
		if num == item {
			return true
		}
	}

	return found
}
func SanitizeDocName(docName string) string {
	re := regexp.MustCompile(`[\x00-\x1F\x7F]`)
	docName = re.ReplaceAllString(docName, "")
	docName = regexp.MustCompile(`[^A-Za-z0-9.\-]`).ReplaceAllString(docName, " ")

	docName = strings.ReplaceAll(docName, "-", "")
	docName = strings.ReplaceAll(docName, ",", "")

	docName = strings.ReplaceAll(docName, " ", "_")

	return docName
}

func MakeTimestampMilli() string {
	unixMilli := unixMilli(time.Now())
	return strconv.FormatInt(unixMilli, 10)
}
func unixMilli(t time.Time) int64 {
	return t.Round(time.Millisecond).UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}

func GeneratePDF(templatePath string, data map[string]interface{}) ([]byte, error) {
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %v", err)
	}

	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %v", err)
	}

	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		return nil, fmt.Errorf("failed to create PDF generator: %v", err)
	}

	page := wkhtmltopdf.NewPageReader(bytes.NewReader(tpl.Bytes()))
	pdfg.AddPage(page)

	if err := pdfg.Create(); err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %v", err)
	}

	return pdfg.Bytes(), nil
}

var (
	ErrDuplicateEntry = errors.New("duplicate entry")
	ErrInvalidParent  = errors.New("invalid parent")
	ErrInvalidType    = errors.New("invalid node type")
	ErrNotFound       = errors.New("record not found")
)

var (
	ErrInvalidNodeType  = errors.New("invalid node type")
	ErrInvalidNode      = errors.New("invalid inventory node")
	ErrAlreadyAllocated = errors.New("already allocated")
	ErrNodeNotFound     = errors.New("node not found")
	ErrInvalidLocation  = errors.New("invalid location id")
)

const (
	PathHallRoomCot = "HALL_ROOM_COT"
	PathRoomCot     = "ROOM_COT"
	PathHallCot     = "HALL_COT"

	NodeHall = "HALL"
	NodeRoom = "ROOM"
	NodeCot  = "COT"
)
const (
	AllocationStatusActive   = "ACTIVE"
	AllocationStatusReleased = "RELEASED"
)
const (
	InventoryStatusAvailable   = "AVAILABLE"
	InventoryStatusOccupied    = "OCCUPIED"
	InventoryStatusMaintenance = "MAINTENANCE"
	InventoryStatusInactive    = "INACTIVE"
	InventoryStatusDeleted     = "DELETED"
)
const (
	ModeStrict = "STRICT"
	ModeForce  = "FORCE"
)
