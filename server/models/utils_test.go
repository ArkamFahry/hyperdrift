package models

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestIsValidBucketName(t *testing.T) {
	a := assert.New(t)

	// Test cases for valid bucket names
	a.True(IsValidBucketName("mybucket"), "Bucket Name")
	a.True(IsValidBucketName("my-bucket"), "Bucket Name with hyphens")
	a.True(IsValidBucketName("1234"), "Bucket Name with numbers")
	a.True(IsValidBucketName("a.b.c"), "Bucket Name with dots")
	a.True(IsValidBucketName("my.bucket.name"), "Bucket Name with dots")

	// Test cases for invalid bucket names
	a.False(IsValidBucketName(""), "Bucket Name Empty String")
	a.False(IsValidBucketName("a"), "Bucket Name Too short")
	a.False(IsValidBucketName(strings.Repeat("a", 64)), "Bucket Name Too long")
	a.False(IsValidBucketName("-invalid"), "Bucket Name Starts with -")
	a.False(IsValidBucketName("invalid-"), "Bucket Name Ends with -")
	a.False(IsValidBucketName("in valid"), "Bucket Name Contains space")
	a.False(IsValidBucketName("invalid@"), "Bucket Name Contains @")
	a.False(IsValidBucketName("my_bucket_name"), "Bucket Name with underscores")
	a.False(IsValidBucketName("my\tbucket\tname"), "Bucket Name with tabs")
	a.False(IsValidBucketName("my\nbucket\nname"), "Bucket Name with newlines")
	a.False(IsValidBucketName("my\tbucket\nname"), "Bucket Name with tabs and newlines")
}

func TestIsValidObjectName(t *testing.T) {
	a := assert.New(t)

	// Test cases for valid object names
	a.True(IsValidObjectName("file.txt"), "Object Name")
	a.True(IsValidObjectName("folder/file.txt"), "Object Name with folder")
	a.True(IsValidObjectName("folder/subfolder/file.txt"), "Object Name with folder and subfolders")

	// Test cases for invalid object names
	a.False(IsValidObjectName(""), "Object Name Empty String")
	a.False(IsValidObjectName("/invalid"), "Object Name Starts with /")
	a.False(IsValidObjectName("invalid/"), "Object Name Ends with /")
	a.False(IsValidObjectName("inva\nlid/hello/world.txt"), "Object Name Contains newline")
	a.False(IsValidObjectName("invalid\thello/world.txt"), "Object Name Contains tab")
	a.False(IsValidObjectName("invalid\t/hello\n/world.txt"), "Object Name Contains newline and tab")
	a.False(IsValidObjectName(strings.Repeat("a", 962)), "Object Name Too long")
}

func TestIsValidMimeType(t *testing.T) {
	a := assert.New(t)

	// Test cases for valid mime types
	a.True(IsValidMimeType("text/plain"), "Mime Type Text Plain")
	a.True(IsValidMimeType("image/jpeg"), "Mime Type Image Jpeg")
	a.True(IsValidMimeType("application/json"), "Mime Type Application Json")

	// Test cases for invalid mime types
	a.False(IsValidMimeType(""))
	a.False(IsValidMimeType("text/ plain"), "Mime Type Contains space")
	a.False(IsValidMimeType("invalid"), "Mime Type Missing /")
	a.False(IsValidMimeType("text/"), "Mime Type Missing subtype")
	a.False(IsValidMimeType("/plain"), "Mime Type Missing type")
}

func TestIsNotEmptyTrimmedString(t *testing.T) {
	a := assert.New(t)

	// Test cases for non-empty trimmed strings
	a.True(IsNotEmptyTrimmedString("hello"))
	a.True(IsNotEmptyTrimmedString("  hello  "))

	// Test cases for empty or whitespace strings
	a.False(IsNotEmptyTrimmedString(""))
	a.False(IsNotEmptyTrimmedString(" "))
	a.False(IsNotEmptyTrimmedString("    "))
}
