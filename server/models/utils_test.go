package models

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestIsValidBucketName(t *testing.T) {
	a := assert.New(t)

	// Test cases for valid bucket names
	a.True(isValidBucketName("mybucket"), "Bucket Name")
	a.True(isValidBucketName("my-bucket"), "Bucket Name with hyphens")
	a.True(isValidBucketName("1234"), "Bucket Name with numbers")
	a.True(isValidBucketName("a.b.c"), "Bucket Name with dots")
	a.True(isValidBucketName("my.bucket.name"), "Bucket Name with dots")

	// Test cases for invalid bucket names
	a.False(isValidBucketName(""), "Bucket Name Empty String")
	a.False(isValidBucketName("a"), "Bucket Name Too short")
	a.False(isValidBucketName(strings.Repeat("a", 64)), "Bucket Name Too long")
	a.False(isValidBucketName("-invalid"), "Bucket Name Starts with -")
	a.False(isValidBucketName("invalid-"), "Bucket Name Ends with -")
	a.False(isValidBucketName("in valid"), "Bucket Name Contains space")
	a.False(isValidBucketName("invalid@"), "Bucket Name Contains @")
	a.False(isValidBucketName("my_bucket_name"), "Bucket Name with underscores")
	a.False(isValidBucketName("my\tbucket\tname"), "Bucket Name with tabs")
	a.False(isValidBucketName("my\nbucket\nname"), "Bucket Name with newlines")
	a.False(isValidBucketName("my\tbucket\nname"), "Bucket Name with tabs and newlines")
}

func TestIsValidObjectName(t *testing.T) {
	a := assert.New(t)

	// Test cases for valid object names
	a.True(isValidObjectName("file.txt"), "Object Name")
	a.True(isValidObjectName("folder/file.txt"), "Object Name with folder")
	a.True(isValidObjectName("folder/subfolder/file.txt"), "Object Name with folder and subfolders")

	// Test cases for invalid object names
	a.False(isValidObjectName(""), "Object Name Empty String")
	a.False(isValidObjectName("/invalid"), "Object Name Starts with /")
	a.False(isValidObjectName("invalid/"), "Object Name Ends with /")
	a.False(isValidObjectName("inva\nlid/hello/world.txt"), "Object Name Contains newline")
	a.False(isValidObjectName("invalid\thello/world.txt"), "Object Name Contains tab")
	a.False(isValidObjectName("invalid\t/hello\n/world.txt"), "Object Name Contains newline and tab")
	a.False(isValidObjectName(strings.Repeat("a", 962)), "Object Name Too long")
}

func TestIsValidMimeType(t *testing.T) {
	a := assert.New(t)

	// Test cases for valid mime types
	a.True(isValidMimeType("text/plain"), "Mime Type Text Plain")
	a.True(isValidMimeType("image/jpeg"), "Mime Type Image Jpeg")
	a.True(isValidMimeType("application/json"), "Mime Type Application Json")

	// Test cases for invalid mime types
	a.False(isValidMimeType(""))
	a.False(isValidMimeType("text/ plain"), "Mime Type Contains space")
	a.False(isValidMimeType("invalid"), "Mime Type Missing /")
	a.False(isValidMimeType("text/"), "Mime Type Missing subtype")
	a.False(isValidMimeType("/plain"), "Mime Type Missing type")
}

func TestIsNotEmptyTrimmedString(t *testing.T) {
	a := assert.New(t)

	// Test cases for non-empty trimmed strings
	a.True(isNotEmptyTrimmedString("hello"))
	a.True(isNotEmptyTrimmedString("  hello  "))

	// Test cases for empty or whitespace strings
	a.False(isNotEmptyTrimmedString(""))
	a.False(isNotEmptyTrimmedString(" "))
	a.False(isNotEmptyTrimmedString("    "))
}
