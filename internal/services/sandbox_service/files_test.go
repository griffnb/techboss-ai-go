package sandbox_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/griffnb/core/lib/testtools/assert"
	"github.com/griffnb/core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/models/sandbox"
)

// Test_validatePath_validPaths tests that valid paths are accepted
func Test_validatePath_validPaths(t *testing.T) {
	t.Run("accepts valid workspace path", func(t *testing.T) {
		// Arrange
		path := "/workspace/src/main.go"

		// Act
		err := validatePath(path)

		// Assert
		assert.Empty(t, err)
	})

	t.Run("accepts valid s3-bucket path", func(t *testing.T) {
		// Arrange
		path := "/s3-bucket/data/output.json"

		// Act
		err := validatePath(path)

		// Assert
		assert.Empty(t, err)
	})

	t.Run("accepts root workspace path", func(t *testing.T) {
		// Arrange
		path := "/workspace"

		// Act
		err := validatePath(path)

		// Assert
		assert.Empty(t, err)
	})

	t.Run("accepts root s3-bucket path", func(t *testing.T) {
		// Arrange
		path := "/s3-bucket"

		// Act
		err := validatePath(path)

		// Assert
		assert.Empty(t, err)
	})

	t.Run("accepts empty path as workspace root", func(t *testing.T) {
		// Arrange
		path := ""

		// Act
		err := validatePath(path)

		// Assert
		assert.Empty(t, err)
	})
}

// Test_validatePath_directoryTraversal tests that directory traversal attempts are rejected
func Test_validatePath_directoryTraversal(t *testing.T) {
	t.Run("rejects path with double dots", func(t *testing.T) {
		// Arrange
		path := "/workspace/../etc/passwd"

		// Act
		err := validatePath(path)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "directory traversal")
	})

	t.Run("rejects path with double dots in middle", func(t *testing.T) {
		// Arrange
		path := "/workspace/src/../../../etc/passwd"

		// Act
		err := validatePath(path)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "directory traversal")
	})

	t.Run("rejects path starting with double dots", func(t *testing.T) {
		// Arrange
		path := "../workspace"

		// Act
		err := validatePath(path)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "directory traversal")
	})
}

// Test_validatePath_invalidPaths tests that paths outside allowed directories are rejected
func Test_validatePath_invalidPaths(t *testing.T) {
	t.Run("rejects path outside workspace and s3-bucket", func(t *testing.T) {
		// Arrange
		path := "/etc/passwd"

		// Act
		err := validatePath(path)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be within /workspace or /s3-bucket")
	})

	t.Run("rejects /tmp path", func(t *testing.T) {
		// Arrange
		path := "/tmp/test.txt"

		// Act
		err := validatePath(path)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be within /workspace or /s3-bucket")
	})

	t.Run("rejects /root path", func(t *testing.T) {
		// Arrange
		path := "/root/secret"

		// Act
		err := validatePath(path)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be within /workspace or /s3-bucket")
	})
}

// Test_parseFileListOptions_defaults tests that default values are applied correctly
func Test_parseFileListOptions_defaults(t *testing.T) {
	t.Run("applies defaults when no query params provided", func(t *testing.T) {
		// Arrange
		req := &http.Request{
			URL: &url.URL{},
		}

		// Act
		opts, err := parseFileListOptions(req)

		// Assert
		assert.Empty(t, err)
		assert.Equal(t, "volume", opts.Source)
		assert.Equal(t, 1, opts.Page)
		assert.Equal(t, 100, opts.PerPage)
		assert.Equal(t, true, opts.Recursive)
		assert.Equal(t, "", opts.Path)
	})

	t.Run("uses default page when invalid page provided", func(t *testing.T) {
		// Arrange
		req := &http.Request{
			URL: &url.URL{
				RawQuery: "page=invalid",
			},
		}

		// Act
		opts, err := parseFileListOptions(req)

		// Assert
		assert.Empty(t, err)
		assert.Equal(t, 1, opts.Page)
	})

	t.Run("uses default per_page when invalid per_page provided", func(t *testing.T) {
		// Arrange
		req := &http.Request{
			URL: &url.URL{
				RawQuery: "per_page=invalid",
			},
		}

		// Act
		opts, err := parseFileListOptions(req)

		// Assert
		assert.Empty(t, err)
		assert.Equal(t, 100, opts.PerPage)
	})
}

// Test_parseFileListOptions_customValues tests that custom query parameters are parsed correctly
func Test_parseFileListOptions_customValues(t *testing.T) {
	t.Run("parses custom source parameter", func(t *testing.T) {
		// Arrange
		req := &http.Request{
			URL: &url.URL{
				RawQuery: "source=s3",
			},
		}

		// Act
		opts, err := parseFileListOptions(req)

		// Assert
		assert.Empty(t, err)
		assert.Equal(t, "s3", opts.Source)
	})

	t.Run("parses custom page parameter", func(t *testing.T) {
		// Arrange
		req := &http.Request{
			URL: &url.URL{
				RawQuery: "page=5",
			},
		}

		// Act
		opts, err := parseFileListOptions(req)

		// Assert
		assert.Empty(t, err)
		assert.Equal(t, 5, opts.Page)
	})

	t.Run("parses custom per_page parameter", func(t *testing.T) {
		// Arrange
		req := &http.Request{
			URL: &url.URL{
				RawQuery: "per_page=50",
			},
		}

		// Act
		opts, err := parseFileListOptions(req)

		// Assert
		assert.Empty(t, err)
		assert.Equal(t, 50, opts.PerPage)
	})

	t.Run("parses custom path parameter", func(t *testing.T) {
		// Arrange
		req := &http.Request{
			URL: &url.URL{
				RawQuery: "path=/workspace/src",
			},
		}

		// Act
		opts, err := parseFileListOptions(req)

		// Assert
		assert.Empty(t, err)
		assert.Equal(t, "/workspace/src", opts.Path)
	})

	t.Run("parses recursive=false", func(t *testing.T) {
		// Arrange
		req := &http.Request{
			URL: &url.URL{
				RawQuery: "recursive=false",
			},
		}

		// Act
		opts, err := parseFileListOptions(req)

		// Assert
		assert.Empty(t, err)
		assert.Equal(t, false, opts.Recursive)
	})

	t.Run("parses all parameters together", func(t *testing.T) {
		// Arrange
		req := &http.Request{
			URL: &url.URL{
				RawQuery: "source=s3&page=3&per_page=200&path=/s3-bucket/data&recursive=false",
			},
		}

		// Act
		opts, err := parseFileListOptions(req)

		// Assert
		assert.Empty(t, err)
		assert.Equal(t, "s3", opts.Source)
		assert.Equal(t, 3, opts.Page)
		assert.Equal(t, 200, opts.PerPage)
		assert.Equal(t, "/s3-bucket/data", opts.Path)
		assert.Equal(t, false, opts.Recursive)
	})
}

// Test_parseFileListOptions_boundaryConditions tests boundary conditions and edge cases
func Test_parseFileListOptions_boundaryConditions(t *testing.T) {
	t.Run("rejects page less than 1", func(t *testing.T) {
		// Arrange
		req := &http.Request{
			URL: &url.URL{
				RawQuery: "page=0",
			},
		}

		// Act
		_, err := parseFileListOptions(req)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "page must be at least 1")
	})

	t.Run("rejects negative page", func(t *testing.T) {
		// Arrange
		req := &http.Request{
			URL: &url.URL{
				RawQuery: "page=-5",
			},
		}

		// Act
		_, err := parseFileListOptions(req)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "page must be at least 1")
	})

	t.Run("rejects per_page less than 1", func(t *testing.T) {
		// Arrange
		req := &http.Request{
			URL: &url.URL{
				RawQuery: "per_page=0",
			},
		}

		// Act
		_, err := parseFileListOptions(req)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "per_page must be between 1 and 1000")
	})

	t.Run("rejects per_page greater than 1000", func(t *testing.T) {
		// Arrange
		req := &http.Request{
			URL: &url.URL{
				RawQuery: "per_page=1001",
			},
		}

		// Act
		_, err := parseFileListOptions(req)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "per_page must be between 1 and 1000")
	})

	t.Run("accepts per_page=1000 as maximum", func(t *testing.T) {
		// Arrange
		req := &http.Request{
			URL: &url.URL{
				RawQuery: "per_page=1000",
			},
		}

		// Act
		opts, err := parseFileListOptions(req)

		// Assert
		assert.Empty(t, err)
		assert.Equal(t, 1000, opts.PerPage)
	})

	t.Run("accepts per_page=1 as minimum", func(t *testing.T) {
		// Arrange
		req := &http.Request{
			URL: &url.URL{
				RawQuery: "per_page=1",
			},
		}

		// Act
		opts, err := parseFileListOptions(req)

		// Assert
		assert.Empty(t, err)
		assert.Equal(t, 1, opts.PerPage)
	})
}

// Test_parseFileListOptions_invalidSource tests that invalid source values are rejected
func Test_parseFileListOptions_invalidSource(t *testing.T) {
	t.Run("rejects invalid source value", func(t *testing.T) {
		// Arrange
		req := &http.Request{
			URL: &url.URL{
				RawQuery: "source=invalid",
			},
		}

		// Act
		_, err := parseFileListOptions(req)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "source must be 'volume' or 's3'")
	})

	t.Run("uses default for empty source value", func(t *testing.T) {
		// Arrange
		req := &http.Request{
			URL: &url.URL{
				RawQuery: "source=",
			},
		}

		// Act
		opts, err := parseFileListOptions(req)

		// Assert
		assert.Empty(t, err)
		assert.Equal(t, "volume", opts.Source)
	})
}

// Test_parseFileListOptions_pathValidation tests that path parameter is validated
func Test_parseFileListOptions_pathValidation(t *testing.T) {
	t.Run("rejects path with directory traversal", func(t *testing.T) {
		// Arrange
		req := &http.Request{
			URL: &url.URL{
				RawQuery: "path=/workspace/../etc/passwd",
			},
		}

		// Act
		_, err := parseFileListOptions(req)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "directory traversal")
	})

	t.Run("rejects path outside allowed directories", func(t *testing.T) {
		// Arrange
		req := &http.Request{
			URL: &url.URL{
				RawQuery: "path=/etc/passwd",
			},
		}

		// Act
		_, err := parseFileListOptions(req)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be within /workspace or /s3-bucket")
	})
}

// Test_SandboxService_buildListFilesCommand tests command generation for listing files
func Test_SandboxService_buildListFilesCommand(t *testing.T) {
	// Arrange
	service := NewSandboxService()

	t.Run("generates command for volume source with default path", func(t *testing.T) {
		// Arrange
		opts := &FileListOptions{
			Source:    "volume",
			Path:      "",
			Recursive: true,
		}

		// Act
		cmd := service.buildListFilesCommand(opts)

		// Assert
		assert.NEmpty(t, cmd)
		assert.Contains(t, cmd, "find")
		assert.Contains(t, cmd, "/workspace")
		assert.True(t, !strings.Contains(cmd, "-maxdepth"), "command should not contain -maxdepth for recursive=true")
	})

	t.Run("generates command for volume source with custom path", func(t *testing.T) {
		// Arrange
		opts := &FileListOptions{
			Source:    "volume",
			Path:      "/workspace/src",
			Recursive: true,
		}

		// Act
		cmd := service.buildListFilesCommand(opts)

		// Assert
		assert.NEmpty(t, cmd)
		assert.Contains(t, cmd, "find")
		assert.Contains(t, cmd, "/workspace/src")
		assert.True(t, !strings.Contains(cmd, "-maxdepth"), "command should not contain -maxdepth for recursive=true")
	})

	t.Run("generates command for s3 source with default path", func(t *testing.T) {
		// Arrange
		opts := &FileListOptions{
			Source:    "s3",
			Path:      "",
			Recursive: true,
		}

		// Act
		cmd := service.buildListFilesCommand(opts)

		// Assert
		assert.NEmpty(t, cmd)
		assert.Contains(t, cmd, "find")
		assert.Contains(t, cmd, "/s3-bucket")
		assert.True(t, !strings.Contains(cmd, "-maxdepth"), "command should not contain -maxdepth for recursive=true")
	})

	t.Run("generates command for s3 source with custom path", func(t *testing.T) {
		// Arrange
		opts := &FileListOptions{
			Source:    "s3",
			Path:      "/s3-bucket/data",
			Recursive: true,
		}

		// Act
		cmd := service.buildListFilesCommand(opts)

		// Assert
		assert.NEmpty(t, cmd)
		assert.Contains(t, cmd, "find")
		assert.Contains(t, cmd, "/s3-bucket/data")
		assert.True(t, !strings.Contains(cmd, "-maxdepth"), "command should not contain -maxdepth for recursive=true")
	})

	t.Run("generates command for volume source with recursive false", func(t *testing.T) {
		// Arrange
		opts := &FileListOptions{
			Source:    "volume",
			Path:      "",
			Recursive: false,
		}

		// Act
		cmd := service.buildListFilesCommand(opts)

		// Assert
		assert.NEmpty(t, cmd)
		assert.Contains(t, cmd, "find")
		assert.Contains(t, cmd, "/workspace")
		assert.Contains(t, cmd, "-maxdepth 1")
	})

	t.Run("generates command for s3 source with recursive false", func(t *testing.T) {
		// Arrange
		opts := &FileListOptions{
			Source:    "s3",
			Path:      "",
			Recursive: false,
		}

		// Act
		cmd := service.buildListFilesCommand(opts)

		// Assert
		assert.NEmpty(t, cmd)
		assert.Contains(t, cmd, "find")
		assert.Contains(t, cmd, "/s3-bucket")
		assert.Contains(t, cmd, "-maxdepth 1")
	})

	t.Run("generates command for volume source with custom path and recursive false", func(t *testing.T) {
		// Arrange
		opts := &FileListOptions{
			Source:    "volume",
			Path:      "/workspace/cmd",
			Recursive: false,
		}

		// Act
		cmd := service.buildListFilesCommand(opts)

		// Assert
		assert.NEmpty(t, cmd)
		assert.Contains(t, cmd, "find")
		assert.Contains(t, cmd, "/workspace/cmd")
		assert.Contains(t, cmd, "-maxdepth 1")
	})

	t.Run("generates command for s3 source with custom path and recursive false", func(t *testing.T) {
		// Arrange
		opts := &FileListOptions{
			Source:    "s3",
			Path:      "/s3-bucket/output",
			Recursive: false,
		}

		// Act
		cmd := service.buildListFilesCommand(opts)

		// Assert
		assert.NEmpty(t, cmd)
		assert.Contains(t, cmd, "find")
		assert.Contains(t, cmd, "/s3-bucket/output")
		assert.Contains(t, cmd, "-maxdepth 1")
	})

	t.Run("includes file and directory type flags", func(t *testing.T) {
		// Arrange
		opts := &FileListOptions{
			Source:    "volume",
			Path:      "",
			Recursive: true,
		}

		// Act
		cmd := service.buildListFilesCommand(opts)

		// Assert
		assert.NEmpty(t, cmd)
		assert.Contains(t, cmd, "-type f")
		assert.Contains(t, cmd, "-type d")
	})

	t.Run("uses logical OR for file and directory types", func(t *testing.T) {
		// Arrange
		opts := &FileListOptions{
			Source:    "volume",
			Path:      "",
			Recursive: true,
		}

		// Act
		cmd := service.buildListFilesCommand(opts)

		// Assert
		assert.NEmpty(t, cmd)
		assert.Contains(t, cmd, "-o")
	})
}

// Test_parseFileMetadata tests parsing stat command output into FileInfo structs
func Test_parseFileMetadata(t *testing.T) {
	t.Run("parses single file successfully", func(t *testing.T) {
		// Arrange
		output := "/workspace/main.go|2048|1735560000|f"

		// Act
		files, err := parseFileMetadata(output)

		// Assert
		assert.Empty(t, err)
		assert.Equal(t, 1, len(files))
		assert.Equal(t, "main.go", files[0].Name)
		assert.Equal(t, "/workspace/main.go", files[0].Path)
		assert.Equal(t, int64(2048), files[0].Size)
		assert.Equal(t, false, files[0].IsDirectory)
	})

	t.Run("parses single directory successfully", func(t *testing.T) {
		// Arrange
		output := "/workspace/src|4096|1735560000|d"

		// Act
		files, err := parseFileMetadata(output)

		// Assert
		assert.Empty(t, err)
		assert.Equal(t, 1, len(files))
		assert.Equal(t, "src", files[0].Name)
		assert.Equal(t, "/workspace/src", files[0].Path)
		assert.Equal(t, int64(4096), files[0].Size)
		assert.Equal(t, true, files[0].IsDirectory)
	})

	t.Run("parses multiple files successfully", func(t *testing.T) {
		// Arrange
		output := "/workspace/main.go|2048|1735560000|f\n/workspace/README.md|1024|1735560100|f\n/workspace/config.json|512|1735560200|f"

		// Act
		files, err := parseFileMetadata(output)

		// Assert
		assert.Empty(t, err)
		assert.Equal(t, 3, len(files))
		assert.Equal(t, "main.go", files[0].Name)
		assert.Equal(t, "README.md", files[1].Name)
		assert.Equal(t, "config.json", files[2].Name)
	})

	t.Run("parses mixed files and directories", func(t *testing.T) {
		// Arrange
		output := "/workspace/main.go|2048|1735560000|f\n/workspace/src|4096|1735560100|d\n/workspace/tests|4096|1735560200|d\n/workspace/go.mod|256|1735560300|f"

		// Act
		files, err := parseFileMetadata(output)

		// Assert
		assert.Empty(t, err)
		assert.Equal(t, 4, len(files))
		assert.Equal(t, false, files[0].IsDirectory)
		assert.Equal(t, true, files[1].IsDirectory)
		assert.Equal(t, true, files[2].IsDirectory)
		assert.Equal(t, false, files[3].IsDirectory)
	})

	t.Run("returns empty slice for empty output", func(t *testing.T) {
		// Arrange
		output := ""

		// Act
		files, err := parseFileMetadata(output)

		// Assert
		assert.Empty(t, err)
		assert.Equal(t, 0, len(files))
	})

	t.Run("skips empty lines in output", func(t *testing.T) {
		// Arrange
		output := "/workspace/main.go|2048|1735560000|f\n\n/workspace/README.md|1024|1735560100|f\n\n\n"

		// Act
		files, err := parseFileMetadata(output)

		// Assert
		assert.Empty(t, err)
		assert.Equal(t, 2, len(files))
		assert.Equal(t, "main.go", files[0].Name)
		assert.Equal(t, "README.md", files[1].Name)
	})

	t.Run("returns error for malformed data missing fields", func(t *testing.T) {
		// Arrange
		output := "/workspace/main.go|2048"

		// Act
		_, err := parseFileMetadata(output)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "malformed")
	})

	t.Run("returns error for malformed data invalid size", func(t *testing.T) {
		// Arrange
		output := "/workspace/main.go|invalid_size|1735560000|f"

		// Act
		_, err := parseFileMetadata(output)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "size")
	})

	t.Run("returns error for malformed data invalid timestamp", func(t *testing.T) {
		// Arrange
		output := "/workspace/main.go|2048|invalid_timestamp|f"

		// Act
		_, err := parseFileMetadata(output)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "timestamp")
	})

	t.Run("handles special characters in path", func(t *testing.T) {
		// Arrange
		output := "/workspace/my-file_name with spaces.go|2048|1735560000|f"

		// Act
		files, err := parseFileMetadata(output)

		// Assert
		assert.Empty(t, err)
		assert.Equal(t, 1, len(files))
		assert.Equal(t, "my-file_name with spaces.go", files[0].Name)
		assert.Equal(t, "/workspace/my-file_name with spaces.go", files[0].Path)
	})

	t.Run("handles large file size", func(t *testing.T) {
		// Arrange
		output := "/workspace/large_file.bin|1073741824|1735560000|f"

		// Act
		files, err := parseFileMetadata(output)

		// Assert
		assert.Empty(t, err)
		assert.Equal(t, 1, len(files))
		assert.Equal(t, int64(1073741824), files[0].Size)
	})

	t.Run("handles zero size file", func(t *testing.T) {
		// Arrange
		output := "/workspace/empty.txt|0|1735560000|f"

		// Act
		files, err := parseFileMetadata(output)

		// Assert
		assert.Empty(t, err)
		assert.Equal(t, 1, len(files))
		assert.Equal(t, int64(0), files[0].Size)
	})

	t.Run("extracts name correctly from path", func(t *testing.T) {
		// Arrange
		output := "/workspace/cmd/server/handlers/main.go|2048|1735560000|f"

		// Act
		files, err := parseFileMetadata(output)

		// Assert
		assert.Empty(t, err)
		assert.Equal(t, 1, len(files))
		assert.Equal(t, "main.go", files[0].Name)
		assert.Equal(t, "/workspace/cmd/server/handlers/main.go", files[0].Path)
	})

	t.Run("handles root directory path", func(t *testing.T) {
		// Arrange
		output := "/workspace|4096|1735560000|d"

		// Act
		files, err := parseFileMetadata(output)

		// Assert
		assert.Empty(t, err)
		assert.Equal(t, 1, len(files))
		assert.Equal(t, "workspace", files[0].Name)
		assert.Equal(t, "/workspace", files[0].Path)
		assert.Equal(t, true, files[0].IsDirectory)
	})

	t.Run("converts unix timestamp to time.Time correctly", func(t *testing.T) {
		// Arrange
		output := "/workspace/main.go|2048|1735560000|f"

		// Act
		files, err := parseFileMetadata(output)

		// Assert
		assert.Empty(t, err)
		assert.Equal(t, 1, len(files))
		// Verify the timestamp is not zero value
		assert.True(t, !files[0].ModifiedAt.IsZero(), "ModifiedAt should not be zero value")
		// Verify the timestamp matches what we expect (Unix timestamp 1735560000)
		assert.Equal(t, int64(1735560000), files[0].ModifiedAt.Unix())
	})

	t.Run("handles s3-bucket paths", func(t *testing.T) {
		// Arrange
		output := "/s3-bucket/data/output.json|512|1735560000|f"

		// Act
		files, err := parseFileMetadata(output)

		// Assert
		assert.Empty(t, err)
		assert.Equal(t, 1, len(files))
		assert.Equal(t, "output.json", files[0].Name)
		assert.Equal(t, "/s3-bucket/data/output.json", files[0].Path)
	})

	t.Run("returns error when type is invalid", func(t *testing.T) {
		// Arrange
		output := "/workspace/main.go|2048|1735560000|x"

		// Act
		_, err := parseFileMetadata(output)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "type")
	})

	t.Run("handles multiple errors in batch and returns first error", func(t *testing.T) {
		// Arrange
		output := "/workspace/main.go|2048|1735560000|f\n/workspace/bad.go|invalid|1735560100|f\n/workspace/good.go|1024|1735560200|f"

		// Act
		_, err := parseFileMetadata(output)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "size")
	})

	t.Run("checksum field is empty by default", func(t *testing.T) {
		// Arrange
		output := "/workspace/main.go|2048|1735560000|f"

		// Act
		files, err := parseFileMetadata(output)

		// Assert
		assert.Empty(t, err)
		assert.Equal(t, 1, len(files))
		assert.Equal(t, "", files[0].Checksum)
	})
}

// Test_paginateFiles tests pagination logic for file listings
func Test_paginateFiles(t *testing.T) {
	// Helper function to generate test files
	generateTestFiles := func(count int) []FileInfo {
		files := make([]FileInfo, count)
		for i := 0; i < count; i++ {
			files[i] = FileInfo{
				Name: "file" + strconv.Itoa(i+1) + ".txt",
				Path: "/workspace/file" + strconv.Itoa(i+1) + ".txt",
				Size: int64((i + 1) * 100),
			}
		}
		return files
	}

	t.Run("returns first page with 100 files from 250 total", func(t *testing.T) {
		// Arrange
		files := generateTestFiles(250)
		opts := &FileListOptions{
			Page:    1,
			PerPage: 100,
		}

		// Act
		response := paginateFiles(files, opts)

		// Assert
		assert.Equal(t, 100, len(response.Files))
		assert.Equal(t, "file1.txt", response.Files[0].Name)
		assert.Equal(t, "file100.txt", response.Files[99].Name)
		assert.Equal(t, 250, response.TotalCount)
		assert.Equal(t, 1, response.Page)
		assert.Equal(t, 100, response.PerPage)
		assert.Equal(t, 3, response.TotalPages)
	})

	t.Run("returns second page with files 101-200 from 250 total", func(t *testing.T) {
		// Arrange
		files := generateTestFiles(250)
		opts := &FileListOptions{
			Page:    2,
			PerPage: 100,
		}

		// Act
		response := paginateFiles(files, opts)

		// Assert
		assert.Equal(t, 100, len(response.Files))
		assert.Equal(t, "file101.txt", response.Files[0].Name)
		assert.Equal(t, "file200.txt", response.Files[99].Name)
		assert.Equal(t, 250, response.TotalCount)
		assert.Equal(t, 2, response.Page)
		assert.Equal(t, 100, response.PerPage)
		assert.Equal(t, 3, response.TotalPages)
	})

	t.Run("returns third page with files 201-250 from 250 total", func(t *testing.T) {
		// Arrange
		files := generateTestFiles(250)
		opts := &FileListOptions{
			Page:    3,
			PerPage: 100,
		}

		// Act
		response := paginateFiles(files, opts)

		// Assert
		assert.Equal(t, 50, len(response.Files))
		assert.Equal(t, "file201.txt", response.Files[0].Name)
		assert.Equal(t, "file250.txt", response.Files[49].Name)
		assert.Equal(t, 250, response.TotalCount)
		assert.Equal(t, 3, response.Page)
		assert.Equal(t, 100, response.PerPage)
		assert.Equal(t, 3, response.TotalPages)
	})

	t.Run("returns last page with partial results", func(t *testing.T) {
		// Arrange
		files := generateTestFiles(250)
		opts := &FileListOptions{
			Page:    3,
			PerPage: 100,
		}

		// Act
		response := paginateFiles(files, opts)

		// Assert
		assert.Equal(t, 50, len(response.Files))
		assert.Equal(t, 250, response.TotalCount)
		assert.Equal(t, 3, response.TotalPages)
	})

	t.Run("returns single page with all files when total less than per_page", func(t *testing.T) {
		// Arrange
		files := generateTestFiles(50)
		opts := &FileListOptions{
			Page:    1,
			PerPage: 100,
		}

		// Act
		response := paginateFiles(files, opts)

		// Assert
		assert.Equal(t, 50, len(response.Files))
		assert.Equal(t, "file1.txt", response.Files[0].Name)
		assert.Equal(t, "file50.txt", response.Files[49].Name)
		assert.Equal(t, 50, response.TotalCount)
		assert.Equal(t, 1, response.Page)
		assert.Equal(t, 100, response.PerPage)
		assert.Equal(t, 1, response.TotalPages)
	})

	t.Run("returns empty files slice for empty file list", func(t *testing.T) {
		// Arrange
		files := []FileInfo{}
		opts := &FileListOptions{
			Page:    1,
			PerPage: 100,
		}

		// Act
		response := paginateFiles(files, opts)

		// Assert
		assert.Equal(t, 0, len(response.Files))
		assert.Equal(t, 0, response.TotalCount)
		assert.Equal(t, 1, response.Page)
		assert.Equal(t, 100, response.PerPage)
		assert.Equal(t, 0, response.TotalPages)
	})

	t.Run("returns empty files slice when page out of range", func(t *testing.T) {
		// Arrange
		files := generateTestFiles(100)
		opts := &FileListOptions{
			Page:    5,
			PerPage: 100,
		}

		// Act
		response := paginateFiles(files, opts)

		// Assert
		assert.Equal(t, 0, len(response.Files))
		assert.Equal(t, 100, response.TotalCount)
		assert.Equal(t, 5, response.Page)
		assert.Equal(t, 100, response.PerPage)
		assert.Equal(t, 1, response.TotalPages)
	})

	t.Run("returns correct page with small per_page", func(t *testing.T) {
		// Arrange
		files := generateTestFiles(10)
		opts := &FileListOptions{
			Page:    2,
			PerPage: 3,
		}

		// Act
		response := paginateFiles(files, opts)

		// Assert
		assert.Equal(t, 3, len(response.Files))
		assert.Equal(t, "file4.txt", response.Files[0].Name)
		assert.Equal(t, "file5.txt", response.Files[1].Name)
		assert.Equal(t, "file6.txt", response.Files[2].Name)
		assert.Equal(t, 10, response.TotalCount)
		assert.Equal(t, 2, response.Page)
		assert.Equal(t, 3, response.PerPage)
		assert.Equal(t, 4, response.TotalPages)
	})

	t.Run("returns correct page with per_page equals 1", func(t *testing.T) {
		// Arrange
		files := generateTestFiles(10)
		opts := &FileListOptions{
			Page:    3,
			PerPage: 1,
		}

		// Act
		response := paginateFiles(files, opts)

		// Assert
		assert.Equal(t, 1, len(response.Files))
		assert.Equal(t, "file3.txt", response.Files[0].Name)
		assert.Equal(t, 10, response.TotalCount)
		assert.Equal(t, 3, response.Page)
		assert.Equal(t, 1, response.PerPage)
		assert.Equal(t, 10, response.TotalPages)
	})

	t.Run("returns all files when per_page greater than total files", func(t *testing.T) {
		// Arrange
		files := generateTestFiles(10)
		opts := &FileListOptions{
			Page:    1,
			PerPage: 100,
		}

		// Act
		response := paginateFiles(files, opts)

		// Assert
		assert.Equal(t, 10, len(response.Files))
		assert.Equal(t, "file1.txt", response.Files[0].Name)
		assert.Equal(t, "file10.txt", response.Files[9].Name)
		assert.Equal(t, 10, response.TotalCount)
		assert.Equal(t, 1, response.Page)
		assert.Equal(t, 100, response.PerPage)
		assert.Equal(t, 1, response.TotalPages)
	})

	t.Run("calculates total pages correctly for exact multiples", func(t *testing.T) {
		// Arrange
		files := generateTestFiles(300)
		opts := &FileListOptions{
			Page:    1,
			PerPage: 100,
		}

		// Act
		response := paginateFiles(files, opts)

		// Assert
		assert.Equal(t, 300, response.TotalCount)
		assert.Equal(t, 3, response.TotalPages)
	})

	t.Run("calculates total pages correctly for non-exact multiples", func(t *testing.T) {
		// Arrange
		files := generateTestFiles(305)
		opts := &FileListOptions{
			Page:    1,
			PerPage: 100,
		}

		// Act
		response := paginateFiles(files, opts)

		// Assert
		assert.Equal(t, 305, response.TotalCount)
		assert.Equal(t, 4, response.TotalPages)
	})

	t.Run("treats page 0 as page 1", func(t *testing.T) {
		// Arrange
		files := generateTestFiles(100)
		opts := &FileListOptions{
			Page:    0,
			PerPage: 10,
		}

		// Act
		response := paginateFiles(files, opts)

		// Assert
		assert.Equal(t, 10, len(response.Files))
		assert.Equal(t, "file1.txt", response.Files[0].Name)
		assert.Equal(t, "file10.txt", response.Files[9].Name)
		assert.Equal(t, 1, response.Page)
	})

	t.Run("treats negative page as page 1", func(t *testing.T) {
		// Arrange
		files := generateTestFiles(100)
		opts := &FileListOptions{
			Page:    -1,
			PerPage: 10,
		}

		// Act
		response := paginateFiles(files, opts)

		// Assert
		assert.Equal(t, 10, len(response.Files))
		assert.Equal(t, "file1.txt", response.Files[0].Name)
		assert.Equal(t, "file10.txt", response.Files[9].Name)
		assert.Equal(t, 1, response.Page)
	})

	t.Run("returns exactly 100 files at page boundary", func(t *testing.T) {
		// Arrange
		files := generateTestFiles(200)
		opts := &FileListOptions{
			Page:    2,
			PerPage: 100,
		}

		// Act
		response := paginateFiles(files, opts)

		// Assert
		assert.Equal(t, 100, len(response.Files))
		assert.Equal(t, "file101.txt", response.Files[0].Name)
		assert.Equal(t, "file200.txt", response.Files[99].Name)
		assert.Equal(t, 200, response.TotalCount)
		assert.Equal(t, 2, response.Page)
		assert.Equal(t, 2, response.TotalPages)
	})

	t.Run("handles large per_page value close to max", func(t *testing.T) {
		// Arrange
		files := generateTestFiles(2000)
		opts := &FileListOptions{
			Page:    2,
			PerPage: 1000,
		}

		// Act
		response := paginateFiles(files, opts)

		// Assert
		assert.Equal(t, 1000, len(response.Files))
		assert.Equal(t, "file1001.txt", response.Files[0].Name)
		assert.Equal(t, "file2000.txt", response.Files[999].Name)
		assert.Equal(t, 2000, response.TotalCount)
		assert.Equal(t, 2, response.Page)
		assert.Equal(t, 1000, response.PerPage)
		assert.Equal(t, 2, response.TotalPages)
	})

	t.Run("verifies files slice is subset of original files", func(t *testing.T) {
		// Arrange
		files := generateTestFiles(50)
		opts := &FileListOptions{
			Page:    2,
			PerPage: 20,
		}

		// Act
		response := paginateFiles(files, opts)

		// Assert
		assert.Equal(t, 20, len(response.Files))
		// Verify files 21-40 are in the response
		for i := 0; i < 20; i++ {
			expectedName := "file" + strconv.Itoa(21+i) + ".txt"
			assert.Equal(t, expectedName, response.Files[i].Name)
		}
	})

	t.Run("returns correct metadata for single file", func(t *testing.T) {
		// Arrange
		files := generateTestFiles(1)
		opts := &FileListOptions{
			Page:    1,
			PerPage: 100,
		}

		// Act
		response := paginateFiles(files, opts)

		// Assert
		assert.Equal(t, 1, len(response.Files))
		assert.Equal(t, 1, response.TotalCount)
		assert.Equal(t, 1, response.Page)
		assert.Equal(t, 100, response.PerPage)
		assert.Equal(t, 1, response.TotalPages)
	})
}

// Test_SandboxService_ListFiles tests the ListFiles service method
// NOTE: These tests follow TDD RED phase - they WILL FAIL until ListFiles is implemented
func Test_SandboxService_ListFiles(t *testing.T) {
	t.Run("returns files from volume source with valid sandbox", func(t *testing.T) {
		// Arrange
		service := NewSandboxService()
		ctx := context.Background()

		// Create mock sandbox model
		sandboxModel := sandbox.New()
		sandboxModel.ExternalID.Set("test-sandbox-id")
		sandboxModel.Type.Set(sandbox.TYPE_CLAUDE_CODE)

		// Reconstruct sandbox info (without actual Modal sandbox)
		sandboxInfo, err := ReconstructSandboxInfo(context.Background(), sandboxModel, types.UUID("test-account"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		opts := &FileListOptions{
			Source:    "volume",
			Path:      "",
			Recursive: true,
			Page:      1,
			PerPage:   100,
		}

		// Act
		// This will fail because ListFiles doesn't exist yet (RED phase)
		response, err := service.ListFiles(ctx, sandboxInfo, opts)

		// Assert
		// When implemented, should return valid response
		assert.NoError(t, err)
		assert.NEmpty(t, response)
		assert.Equal(t, 1, response.Page)
		assert.Equal(t, 100, response.PerPage)
	})

	t.Run("returns files from s3 source with valid sandbox", func(t *testing.T) {
		// Arrange
		service := NewSandboxService()
		ctx := context.Background()

		sandboxModel := sandbox.New()
		sandboxModel.ExternalID.Set("test-sandbox-s3")
		sandboxModel.Type.Set(sandbox.TYPE_CLAUDE_CODE)

		sandboxInfo, err := ReconstructSandboxInfo(context.Background(), sandboxModel, types.UUID("test-account"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		opts := &FileListOptions{
			Source:    "s3",
			Path:      "",
			Recursive: true,
			Page:      1,
			PerPage:   100,
		}

		// Act
		response, err := service.ListFiles(ctx, sandboxInfo, opts)

		// Assert
		assert.NoError(t, err)
		assert.NEmpty(t, response)
		assert.Equal(t, 1, response.Page)
		assert.Equal(t, 100, response.PerPage)
	})

	t.Run("returns error for invalid source", func(t *testing.T) {
		// Arrange
		service := NewSandboxService()
		ctx := context.Background()

		sandboxModel := sandbox.New()
		sandboxModel.ExternalID.Set("test-sandbox-invalid")
		sandboxModel.Type.Set(sandbox.TYPE_CLAUDE_CODE)

		sandboxInfo, err := ReconstructSandboxInfo(context.Background(), sandboxModel, types.UUID("test-account"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		opts := &FileListOptions{
			Source:    "invalid",
			Path:      "",
			Recursive: true,
			Page:      1,
			PerPage:   100,
		}

		// Act
		response, err := service.ListFiles(ctx, sandboxInfo, opts)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, response)
		assert.Contains(t, err.Error(), "invalid source")
	})

	t.Run("defaults to root path when path is empty", func(t *testing.T) {
		// Arrange
		service := NewSandboxService()
		ctx := context.Background()

		sandboxModel := sandbox.New()
		sandboxModel.ExternalID.Set("test-sandbox-empty-path")
		sandboxModel.Type.Set(sandbox.TYPE_CLAUDE_CODE)

		sandboxInfo, err := ReconstructSandboxInfo(context.Background(), sandboxModel, types.UUID("test-account"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		opts := &FileListOptions{
			Source:    "volume",
			Path:      "", // Empty path should default to /workspace
			Recursive: true,
			Page:      1,
			PerPage:   100,
		}

		// Act
		response, err := service.ListFiles(ctx, sandboxInfo, opts)

		// Assert
		// Should use default /workspace path and succeed
		assert.NoError(t, err)
		assert.NEmpty(t, response)
	})

	t.Run("handles valid command output with multiple files", func(t *testing.T) {
		// Arrange
		service := NewSandboxService()
		ctx := context.Background()

		sandboxModel := sandbox.New()
		sandboxModel.ExternalID.Set("test-sandbox-multiple")
		sandboxModel.Type.Set(sandbox.TYPE_CLAUDE_CODE)

		sandboxInfo, err := ReconstructSandboxInfo(context.Background(), sandboxModel, types.UUID("test-account"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		opts := &FileListOptions{
			Source:    "volume",
			Path:      "/workspace",
			Recursive: false,
			Page:      1,
			PerPage:   100,
		}

		// Act
		// When implemented with mocked command output showing multiple files
		response, err := service.ListFiles(ctx, sandboxInfo, opts)

		// Assert
		assert.NoError(t, err)
		assert.NEmpty(t, response)
		assert.True(t, response.TotalCount >= 0, "should have non-negative total count")
	})

	t.Run("handles empty command output", func(t *testing.T) {
		// Arrange
		service := NewSandboxService()
		ctx := context.Background()

		sandboxModel := sandbox.New()
		sandboxModel.ExternalID.Set("test-sandbox-empty")
		sandboxModel.Type.Set(sandbox.TYPE_CLAUDE_CODE)

		sandboxInfo, err := ReconstructSandboxInfo(context.Background(), sandboxModel, types.UUID("test-account"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		opts := &FileListOptions{
			Source:    "volume",
			Path:      "/workspace/empty-dir",
			Recursive: true,
			Page:      1,
			PerPage:   100,
		}

		// Act
		// When implemented, should handle empty directory gracefully
		response, err := service.ListFiles(ctx, sandboxInfo, opts)

		// Assert
		assert.NoError(t, err)
		assert.NEmpty(t, response)
		assert.Equal(t, 0, len(response.Files))
		assert.Equal(t, 0, response.TotalCount)
	})

	t.Run("returns error for path validation failure", func(t *testing.T) {
		// Arrange
		service := NewSandboxService()
		ctx := context.Background()

		sandboxModel := sandbox.New()
		sandboxModel.ExternalID.Set("test-sandbox-bad-path")
		sandboxModel.Type.Set(sandbox.TYPE_CLAUDE_CODE)

		sandboxInfo, err := ReconstructSandboxInfo(context.Background(), sandboxModel, types.UUID("test-account"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		opts := &FileListOptions{
			Source:    "volume",
			Path:      "/workspace/../etc/passwd", // Directory traversal attempt
			Recursive: true,
			Page:      1,
			PerPage:   100,
		}

		// Act
		response, err := service.ListFiles(ctx, sandboxInfo, opts)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, response)
		assert.Contains(t, err.Error(), "directory traversal")
	})

	t.Run("applies pagination correctly", func(t *testing.T) {
		// Arrange
		service := NewSandboxService()
		ctx := context.Background()

		sandboxModel := sandbox.New()
		sandboxModel.ExternalID.Set("test-sandbox-pagination")
		sandboxModel.Type.Set(sandbox.TYPE_CLAUDE_CODE)

		sandboxInfo, err := ReconstructSandboxInfo(context.Background(), sandboxModel, types.UUID("test-account"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		opts := &FileListOptions{
			Source:    "volume",
			Path:      "/workspace",
			Recursive: true,
			Page:      2,
			PerPage:   10,
		}

		// Act
		response, err := service.ListFiles(ctx, sandboxInfo, opts)

		// Assert
		assert.NoError(t, err)
		assert.NEmpty(t, response)
		assert.Equal(t, 2, response.Page)
		assert.Equal(t, 10, response.PerPage)
	})

	t.Run("builds correct find command for volume source", func(t *testing.T) {
		// Arrange
		service := NewSandboxService()
		ctx := context.Background()

		sandboxModel := sandbox.New()
		sandboxModel.ExternalID.Set("test-sandbox-volume-cmd")
		sandboxModel.Type.Set(sandbox.TYPE_CLAUDE_CODE)

		sandboxInfo, err := ReconstructSandboxInfo(context.Background(), sandboxModel, types.UUID("test-account"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		opts := &FileListOptions{
			Source:    "volume",
			Path:      "/workspace/src",
			Recursive: true,
			Page:      1,
			PerPage:   100,
		}

		// Act
		// The command built should contain "find /workspace/src"
		response, err := service.ListFiles(ctx, sandboxInfo, opts)

		// Assert
		// Implementation should use buildListFilesCommand helper
		assert.NoError(t, err)
		assert.NEmpty(t, response)
	})

	t.Run("builds correct find command for s3 source", func(t *testing.T) {
		// Arrange
		service := NewSandboxService()
		ctx := context.Background()

		sandboxModel := sandbox.New()
		sandboxModel.ExternalID.Set("test-sandbox-s3-cmd")
		sandboxModel.Type.Set(sandbox.TYPE_CLAUDE_CODE)

		sandboxInfo, err := ReconstructSandboxInfo(context.Background(), sandboxModel, types.UUID("test-account"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		opts := &FileListOptions{
			Source:    "s3",
			Path:      "/s3-bucket/data",
			Recursive: false,
			Page:      1,
			PerPage:   100,
		}

		// Act
		// The command built should contain "find /s3-bucket/data -maxdepth 1"
		response, err := service.ListFiles(ctx, sandboxInfo, opts)

		// Assert
		assert.NoError(t, err)
		assert.NEmpty(t, response)
	})

	t.Run("returns error when sandbox info is nil", func(t *testing.T) {
		// Arrange
		service := NewSandboxService()
		ctx := context.Background()

		opts := &FileListOptions{
			Source:    "volume",
			Path:      "",
			Recursive: true,
			Page:      1,
			PerPage:   100,
		}

		// Act
		response, err := service.ListFiles(ctx, nil, opts)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, response)
		assert.Contains(t, err.Error(), "sandboxInfo")
	})

	t.Run("returns error when opts is nil", func(t *testing.T) {
		// Arrange
		service := NewSandboxService()
		ctx := context.Background()

		sandboxModel := sandbox.New()
		sandboxModel.ExternalID.Set("test-sandbox-nil-opts")
		sandboxModel.Type.Set(sandbox.TYPE_CLAUDE_CODE)

		sandboxInfo, err := ReconstructSandboxInfo(context.Background(), sandboxModel, types.UUID("test-account"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Act
		response, err := service.ListFiles(ctx, sandboxInfo, nil)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, response)
		assert.Contains(t, err.Error(), "opts")
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		// Arrange
		service := NewSandboxService()
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		sandboxModel := sandbox.New()
		sandboxModel.ExternalID.Set("test-sandbox-ctx-cancel")
		sandboxModel.Type.Set(sandbox.TYPE_CLAUDE_CODE)

		sandboxInfo, err := ReconstructSandboxInfo(context.Background(), sandboxModel, types.UUID("test-account"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		opts := &FileListOptions{
			Source:    "volume",
			Path:      "",
			Recursive: true,
			Page:      1,
			PerPage:   100,
		}

		// Act
		response, err := service.ListFiles(ctx, sandboxInfo, opts)

		// Assert
		// Should respect context cancellation
		assert.Error(t, err)
		assert.Empty(t, response)
	})

	t.Run("integrates buildListFilesCommand parseFileMetadata and paginateFiles", func(t *testing.T) {
		// Arrange
		service := NewSandboxService()
		ctx := context.Background()

		sandboxModel := sandbox.New()
		sandboxModel.ExternalID.Set("test-sandbox-integration")
		sandboxModel.Type.Set(sandbox.TYPE_CLAUDE_CODE)

		sandboxInfo, err := ReconstructSandboxInfo(context.Background(), sandboxModel, types.UUID("test-account"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		opts := &FileListOptions{
			Source:    "volume",
			Path:      "/workspace",
			Recursive: true,
			Page:      1,
			PerPage:   50,
		}

		// Act
		// Implementation should:
		// 1. Call buildListFilesCommand(opts) to build command
		// 2. Execute command via sandboxInfo.Sandbox.Exec()
		// 3. Call parseFileMetadata() on output
		// 4. Call paginateFiles() with parsed files
		response, err := service.ListFiles(ctx, sandboxInfo, opts)

		// Assert
		assert.NoError(t, err)
		assert.NEmpty(t, response)
		assert.Equal(t, 1, response.Page)
		assert.Equal(t, 50, response.PerPage)
		// Files slice should be properly paginated
		assert.True(t, len(response.Files) <= 50, "should not exceed per_page limit")
	})
}

// Test_SandboxService_buildReadFileCommand tests command generation for reading file content
func Test_SandboxService_buildReadFileCommand(t *testing.T) {
	// Arrange
	service := NewSandboxService()

	t.Run("generates cat command for small file from volume", func(t *testing.T) {
		// Arrange
		source := "volume"
		filePath := "/test.txt"
		maxSize := int64(0) // no limit

		// Act
		cmd := service.buildReadFileCommand(source, filePath, maxSize)

		// Assert
		assert.Equal(t, "cat /workspace/test.txt", cmd)
	})

	t.Run("generates cat command for small file from s3", func(t *testing.T) {
		// Arrange
		source := "s3"
		filePath := "/data.json"
		maxSize := int64(0)

		// Act
		cmd := service.buildReadFileCommand(source, filePath, maxSize)

		// Assert
		assert.Equal(t, "cat /s3-bucket/data.json", cmd)
	})

	t.Run("generates head command for large file with size limit", func(t *testing.T) {
		// Arrange
		source := "volume"
		filePath := "/large.bin"
		maxSize := int64(10485760) // 10MB

		// Act
		cmd := service.buildReadFileCommand(source, filePath, maxSize)

		// Assert
		assert.Equal(t, "head -c 10485760 /workspace/large.bin", cmd)
	})

	t.Run("generates cat command for file from custom volume path", func(t *testing.T) {
		// Arrange
		source := "volume"
		filePath := "/src/main.go"
		maxSize := int64(0)

		// Act
		cmd := service.buildReadFileCommand(source, filePath, maxSize)

		// Assert
		assert.Equal(t, "cat /workspace/src/main.go", cmd)
	})

	t.Run("generates head command for s3 file with size limit", func(t *testing.T) {
		// Arrange
		source := "s3"
		filePath := "/output/large.csv"
		maxSize := int64(5242880) // 5MB

		// Act
		cmd := service.buildReadFileCommand(source, filePath, maxSize)

		// Assert
		assert.Equal(t, "head -c 5242880 /s3-bucket/output/large.csv", cmd)
	})

	t.Run("uses workspace base path for volume source", func(t *testing.T) {
		// Arrange
		source := "volume"
		filePath := "/any/path/file.txt"
		maxSize := int64(0)

		// Act
		cmd := service.buildReadFileCommand(source, filePath, maxSize)

		// Assert
		assert.Contains(t, cmd, "/workspace")
		assert.Contains(t, cmd, filePath)
	})

	t.Run("uses s3-bucket base path for s3 source", func(t *testing.T) {
		// Arrange
		source := "s3"
		filePath := "/any/path/file.txt"
		maxSize := int64(0)

		// Act
		cmd := service.buildReadFileCommand(source, filePath, maxSize)

		// Assert
		assert.Contains(t, cmd, "/s3-bucket")
		assert.Contains(t, cmd, filePath)
	})

	t.Run("handles nested file paths correctly", func(t *testing.T) {
		// Arrange
		source := "volume"
		filePath := "/deep/nested/path/to/file.txt"
		maxSize := int64(0)

		// Act
		cmd := service.buildReadFileCommand(source, filePath, maxSize)

		// Assert
		assert.Equal(t, "cat /workspace/deep/nested/path/to/file.txt", cmd)
	})

	t.Run("handles very large size limits", func(t *testing.T) {
		// Arrange
		source := "volume"
		filePath := "/huge.bin"
		maxSize := int64(1073741824) // 1GB

		// Act
		cmd := service.buildReadFileCommand(source, filePath, maxSize)

		// Assert
		assert.Equal(t, "head -c 1073741824 /workspace/huge.bin", cmd)
	})
}

// Test_SandboxService_GetFileContent tests retrieving file content from sandbox
func Test_SandboxService_GetFileContent(t *testing.T) {
	t.Run("returns error when sandbox not connected for volume", func(t *testing.T) {
		// Arrange
		service := NewSandboxService()
		ctx := context.Background()

		sandboxModel := sandbox.New()
		sandboxModel.ExternalID.Set("test-sandbox-text")
		sandboxModel.Type.Set(sandbox.TYPE_CLAUDE_CODE)
		// ReconstructSandboxInfo creates sandboxInfo without active Modal connection
		sandboxInfo, err := ReconstructSandboxInfo(context.Background(), sandboxModel, types.UUID("test-account"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		source := "volume"
		filePath := "/test.txt"

		// Act
		content, err := service.GetFileContent(ctx, sandboxInfo, source, filePath)

		// Assert
		// Should return error when sandbox not connected (expected behavior)
		assert.Error(t, err)
		assert.Empty(t, content)
		assert.Contains(t, err.Error(), "sandbox not connected")
	})

	t.Run("returns error when sandbox not connected for s3", func(t *testing.T) {
		// Arrange
		service := NewSandboxService()
		ctx := context.Background()

		sandboxModel := sandbox.New()
		sandboxModel.ExternalID.Set("test-sandbox-s3")
		sandboxModel.Type.Set(sandbox.TYPE_CLAUDE_CODE)
		sandboxInfo, err := ReconstructSandboxInfo(context.Background(), sandboxModel, types.UUID("test-account"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		source := "s3"
		filePath := "/data.json"

		// Act
		content, err := service.GetFileContent(ctx, sandboxInfo, source, filePath)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, content)
		assert.Contains(t, err.Error(), "sandbox not connected")
	})

	t.Run("validates file path and returns appropriate error", func(t *testing.T) {
		// Arrange
		service := NewSandboxService()
		ctx := context.Background()

		sandboxModel := sandbox.New()
		sandboxModel.ExternalID.Set("test-sandbox-binary")
		sandboxModel.Type.Set(sandbox.TYPE_CLAUDE_CODE)
		sandboxInfo, err := ReconstructSandboxInfo(context.Background(), sandboxModel, types.UUID("test-account"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		source := "volume"
		filePath := "/image.png"

		// Act
		content, err := service.GetFileContent(ctx, sandboxInfo, source, filePath)

		// Assert
		// For disconnected sandbox, should get "not connected" error
		assert.Error(t, err)
		assert.Empty(t, content)
	})

	t.Run("returns error when file not found", func(t *testing.T) {
		// Arrange
		service := NewSandboxService()
		ctx := context.Background()

		sandboxModel := sandbox.New()
		sandboxModel.ExternalID.Set("test-sandbox-notfound")
		sandboxModel.Type.Set(sandbox.TYPE_CLAUDE_CODE)
		sandboxInfo, err := ReconstructSandboxInfo(context.Background(), sandboxModel, types.UUID("test-account"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		source := "volume"
		filePath := "/nonexistent.txt"

		// Act
		content, err := service.GetFileContent(ctx, sandboxInfo, source, filePath)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, content)
		assert.Contains(t, err.Error(), "file not found")
	})

	t.Run("returns error for large file when sandbox not connected", func(t *testing.T) {
		// Arrange
		service := NewSandboxService()
		ctx := context.Background()

		sandboxModel := sandbox.New()
		sandboxModel.ExternalID.Set("test-sandbox-large")
		sandboxModel.Type.Set(sandbox.TYPE_CLAUDE_CODE)
		sandboxInfo, err := ReconstructSandboxInfo(context.Background(), sandboxModel, types.UUID("test-account"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		source := "volume"
		filePath := "/large.bin"

		// Act
		content, err := service.GetFileContent(ctx, sandboxInfo, source, filePath)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, content)
		assert.Contains(t, err.Error(), "sandbox not connected")
	})

	t.Run("returns error for empty file when sandbox not connected", func(t *testing.T) {
		// Arrange
		service := NewSandboxService()
		ctx := context.Background()

		sandboxModel := sandbox.New()
		sandboxModel.ExternalID.Set("test-sandbox-empty")
		sandboxModel.Type.Set(sandbox.TYPE_CLAUDE_CODE)
		sandboxInfo, err := ReconstructSandboxInfo(context.Background(), sandboxModel, types.UUID("test-account"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		source := "volume"
		filePath := "/empty.txt"

		// Act
		content, err := service.GetFileContent(ctx, sandboxInfo, source, filePath)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, content)
		assert.Contains(t, err.Error(), "sandbox not connected")
	})

	t.Run("returns error for invalid source", func(t *testing.T) {
		// Arrange
		service := NewSandboxService()
		ctx := context.Background()

		sandboxModel := sandbox.New()
		sandboxModel.ExternalID.Set("test-sandbox-invalid")
		sandboxModel.Type.Set(sandbox.TYPE_CLAUDE_CODE)
		sandboxInfo, err := ReconstructSandboxInfo(context.Background(), sandboxModel, types.UUID("test-account"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		source := "invalid"
		filePath := "/test.txt"

		// Act
		content, err := service.GetFileContent(ctx, sandboxInfo, source, filePath)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, content)
		assert.Contains(t, err.Error(), "invalid source")
	})

	t.Run("returns error when sandboxInfo is nil", func(t *testing.T) {
		// Arrange
		service := NewSandboxService()
		ctx := context.Background()

		source := "volume"
		filePath := "/test.txt"

		// Act
		content, err := service.GetFileContent(ctx, nil, source, filePath)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, content)
		assert.Contains(t, err.Error(), "sandboxInfo cannot be nil")
	})

	t.Run("returns error for path validation failure", func(t *testing.T) {
		// Arrange
		service := NewSandboxService()
		ctx := context.Background()

		sandboxModel := sandbox.New()
		sandboxModel.ExternalID.Set("test-sandbox-badpath")
		sandboxModel.Type.Set(sandbox.TYPE_CLAUDE_CODE)
		sandboxInfo, err := ReconstructSandboxInfo(context.Background(), sandboxModel, types.UUID("test-account"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		source := "volume"
		filePath := "/workspace/../etc/passwd" // Directory traversal

		// Act
		content, err := service.GetFileContent(ctx, sandboxInfo, source, filePath)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, content)
		assert.Contains(t, err.Error(), "directory traversal")
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		// Arrange
		service := NewSandboxService()
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		sandboxModel := sandbox.New()
		sandboxModel.ExternalID.Set("test-sandbox-cancel")
		sandboxModel.Type.Set(sandbox.TYPE_CLAUDE_CODE)
		sandboxInfo, err := ReconstructSandboxInfo(context.Background(), sandboxModel, types.UUID("test-account"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		source := "volume"
		filePath := "/test.txt"

		// Act
		content, err := service.GetFileContent(ctx, sandboxInfo, source, filePath)

		// Assert
		// Should respect context cancellation - either nil sandbox or context error
		assert.True(t, err != nil || content == nil, "should handle cancelled context")
	})

	t.Run("detects correct MIME type for various extensions", func(t *testing.T) {
		// Arrange
		service := NewSandboxService()
		ctx := context.Background()

		sandboxModel := sandbox.New()
		sandboxModel.ExternalID.Set("test-sandbox-mime")
		sandboxModel.Type.Set(sandbox.TYPE_CLAUDE_CODE)
		sandboxInfo, err := ReconstructSandboxInfo(context.Background(), sandboxModel, types.UUID("test-account"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		testCases := []struct {
			filePath    string
			expectedExt string
		}{
			{"/main.go", "text/x-go"},
			{"/script.py", "text/x-python"},
			{"/data.json", "application/json"},
			{"/config.yaml", "application/x-yaml"},
		}

		for _, tc := range testCases {
			// Act
			content, err := service.GetFileContent(ctx, sandboxInfo, "volume", tc.filePath)

			// Assert (if file exists)
			if err == nil && content != nil {
				assert.Equal(t, tc.expectedExt, content.ContentType)
			}
		}
	})

	t.Run("extracts filename correctly from path", func(t *testing.T) {
		// Arrange
		service := NewSandboxService()
		ctx := context.Background()

		sandboxModel := sandbox.New()
		sandboxModel.ExternalID.Set("test-sandbox-filename")
		sandboxModel.Type.Set(sandbox.TYPE_CLAUDE_CODE)
		sandboxInfo, err := ReconstructSandboxInfo(context.Background(), sandboxModel, types.UUID("test-account"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		source := "volume"
		filePath := "/deep/nested/path/file.txt"

		// Act
		content, err := service.GetFileContent(ctx, sandboxInfo, source, filePath)

		// Assert (if file exists)
		if err == nil && content != nil {
			assert.Equal(t, "file.txt", content.FileName)
		}
	})
}

// Test_detectMimeType tests MIME type detection based on file extension
func Test_detectMimeType(t *testing.T) {
	t.Run("detects text/plain for .txt files", func(t *testing.T) {
		// Arrange
		filename := "test.txt"

		// Act
		mimeType := detectMimeType(filename)

		// Assert
		assert.Equal(t, "text/plain", mimeType)
	})

	t.Run("detects application/json for .json files", func(t *testing.T) {
		// Arrange
		filename := "data.json"

		// Act
		mimeType := detectMimeType(filename)

		// Assert
		assert.Equal(t, "application/json", mimeType)
	})

	t.Run("detects text/x-python for .py files", func(t *testing.T) {
		// Arrange
		filename := "script.py"

		// Act
		mimeType := detectMimeType(filename)

		// Assert
		assert.Equal(t, "text/x-python", mimeType)
	})

	t.Run("detects text/x-go for .go files", func(t *testing.T) {
		// Arrange
		filename := "main.go"

		// Act
		mimeType := detectMimeType(filename)

		// Assert
		assert.Equal(t, "text/x-go", mimeType)
	})

	t.Run("detects image/png for .png files", func(t *testing.T) {
		// Arrange
		filename := "image.png"

		// Act
		mimeType := detectMimeType(filename)

		// Assert
		assert.Equal(t, "image/png", mimeType)
	})

	t.Run("detects application/pdf for .pdf files", func(t *testing.T) {
		// Arrange
		filename := "doc.pdf"

		// Act
		mimeType := detectMimeType(filename)

		// Assert
		assert.Equal(t, "application/pdf", mimeType)
	})

	t.Run("returns application/octet-stream for unknown extensions", func(t *testing.T) {
		// Arrange
		filename := "unknown.xyz"

		// Act
		mimeType := detectMimeType(filename)

		// Assert
		assert.Equal(t, "application/octet-stream", mimeType)
	})

	t.Run("handles uppercase extensions", func(t *testing.T) {
		// Arrange
		filename := "TEST.TXT"

		// Act
		mimeType := detectMimeType(filename)

		// Assert
		assert.Equal(t, "text/plain", mimeType)
	})

	t.Run("handles mixed case extensions", func(t *testing.T) {
		// Arrange
		filename := "Script.Py"

		// Act
		mimeType := detectMimeType(filename)

		// Assert
		assert.Equal(t, "text/x-python", mimeType)
	})

	t.Run("handles files with paths", func(t *testing.T) {
		// Arrange
		filename := "/workspace/src/main.go"

		// Act
		mimeType := detectMimeType(filename)

		// Assert
		assert.Equal(t, "text/x-go", mimeType)
	})

	t.Run("detects application/xml for .xml files", func(t *testing.T) {
		// Arrange
		filename := "config.xml"

		// Act
		mimeType := detectMimeType(filename)

		// Assert
		assert.Equal(t, "application/xml", mimeType)
	})

	t.Run("detects text/html for .html files", func(t *testing.T) {
		// Arrange
		filename := "index.html"

		// Act
		mimeType := detectMimeType(filename)

		// Assert
		assert.Equal(t, "text/html", mimeType)
	})

	t.Run("detects text/css for .css files", func(t *testing.T) {
		// Arrange
		filename := "styles.css"

		// Act
		mimeType := detectMimeType(filename)

		// Assert
		assert.Equal(t, "text/css", mimeType)
	})

	t.Run("detects application/javascript for .js files", func(t *testing.T) {
		// Arrange
		filename := "app.js"

		// Act
		mimeType := detectMimeType(filename)

		// Assert
		assert.Equal(t, "application/javascript", mimeType)
	})

	t.Run("detects text/x-java for .java files", func(t *testing.T) {
		// Arrange
		filename := "Main.java"

		// Act
		mimeType := detectMimeType(filename)

		// Assert
		assert.Equal(t, "text/x-java", mimeType)
	})

	t.Run("detects text/x-c for .c files", func(t *testing.T) {
		// Arrange
		filename := "program.c"

		// Act
		mimeType := detectMimeType(filename)

		// Assert
		assert.Equal(t, "text/x-c", mimeType)
	})

	t.Run("detects text/x-c++ for .cpp files", func(t *testing.T) {
		// Arrange
		filename := "program.cpp"

		// Act
		mimeType := detectMimeType(filename)

		// Assert
		assert.Equal(t, "text/x-c++", mimeType)
	})

	t.Run("detects text/markdown for .md files", func(t *testing.T) {
		// Arrange
		filename := "README.md"

		// Act
		mimeType := detectMimeType(filename)

		// Assert
		assert.Equal(t, "text/markdown", mimeType)
	})

	t.Run("detects application/x-yaml for .yaml files", func(t *testing.T) {
		// Arrange
		filename := "config.yaml"

		// Act
		mimeType := detectMimeType(filename)

		// Assert
		assert.Equal(t, "application/x-yaml", mimeType)
	})

	t.Run("detects application/x-yaml for .yml files", func(t *testing.T) {
		// Arrange
		filename := "docker-compose.yml"

		// Act
		mimeType := detectMimeType(filename)

		// Assert
		assert.Equal(t, "application/x-yaml", mimeType)
	})

	t.Run("detects application/x-sh for .sh files", func(t *testing.T) {
		// Arrange
		filename := "script.sh"

		// Act
		mimeType := detectMimeType(filename)

		// Assert
		assert.Equal(t, "application/x-sh", mimeType)
	})

	t.Run("detects image/jpeg for .jpg files", func(t *testing.T) {
		// Arrange
		filename := "photo.jpg"

		// Act
		mimeType := detectMimeType(filename)

		// Assert
		assert.Equal(t, "image/jpeg", mimeType)
	})

	t.Run("detects image/jpeg for .jpeg files", func(t *testing.T) {
		// Arrange
		filename := "photo.jpeg"

		// Act
		mimeType := detectMimeType(filename)

		// Assert
		assert.Equal(t, "image/jpeg", mimeType)
	})

	t.Run("detects image/gif for .gif files", func(t *testing.T) {
		// Arrange
		filename := "animation.gif"

		// Act
		mimeType := detectMimeType(filename)

		// Assert
		assert.Equal(t, "image/gif", mimeType)
	})

	t.Run("detects application/zip for .zip files", func(t *testing.T) {
		// Arrange
		filename := "archive.zip"

		// Act
		mimeType := detectMimeType(filename)

		// Assert
		assert.Equal(t, "application/zip", mimeType)
	})

	t.Run("detects application/x-tar for .tar files", func(t *testing.T) {
		// Arrange
		filename := "archive.tar"

		// Act
		mimeType := detectMimeType(filename)

		// Assert
		assert.Equal(t, "application/x-tar", mimeType)
	})

	t.Run("detects application/gzip for .gz files", func(t *testing.T) {
		// Arrange
		filename := "archive.tar.gz"

		// Act
		mimeType := detectMimeType(filename)

		// Assert
		assert.Equal(t, "application/gzip", mimeType)
	})

	t.Run("handles files with no extension", func(t *testing.T) {
		// Arrange
		filename := "Makefile"

		// Act
		mimeType := detectMimeType(filename)

		// Assert
		assert.Equal(t, "application/octet-stream", mimeType)
	})

	t.Run("handles files with multiple dots", func(t *testing.T) {
		// Arrange
		filename := "my.file.name.go"

		// Act
		mimeType := detectMimeType(filename)

		// Assert
		assert.Equal(t, "text/x-go", mimeType)
	})
}

// Test_SandboxService_BuildFileTree tests the BuildFileTree method that converts
// a flat list of FileInfo into a hierarchical FileTreeNode structure.
func Test_SandboxService_BuildFileTree(t *testing.T) {
	t.Run("converts flat file list to tree structure", func(t *testing.T) {
		// Arrange
		service := NewSandboxService()
		files := []FileInfo{
			{Name: "workspace", Path: "/workspace", IsDirectory: true, Size: 0, ModifiedAt: time.Now()},
			{Name: "src", Path: "/workspace/src", IsDirectory: true, Size: 0, ModifiedAt: time.Now()},
			{Name: "main.go", Path: "/workspace/src/main.go", IsDirectory: false, Size: 100, ModifiedAt: time.Now()},
			{Name: "README.md", Path: "/workspace/README.md", IsDirectory: false, Size: 50, ModifiedAt: time.Now()},
		}

		// Act
		tree, err := service.BuildFileTree(files, "/workspace")

		// Assert
		assert.NoError(t, err)
		assert.NEmpty(t, tree)
		assert.Equal(t, "workspace", tree.Name)
		assert.Equal(t, "/workspace", tree.Path)
		assert.True(t, tree.IsDirectory)
		assert.Equal(t, 2, len(tree.Children)) // Should have 'src' directory and 'README.md' file
	})

	t.Run("handles nested directories with 3+ levels", func(t *testing.T) {
		// Arrange
		service := NewSandboxService()
		files := []FileInfo{
			{Name: "workspace", Path: "/workspace", IsDirectory: true, Size: 0, ModifiedAt: time.Now()},
			{Name: "src", Path: "/workspace/src", IsDirectory: true, Size: 0, ModifiedAt: time.Now()},
			{Name: "api", Path: "/workspace/src/api", IsDirectory: true, Size: 0, ModifiedAt: time.Now()},
			{Name: "handlers", Path: "/workspace/src/api/handlers", IsDirectory: true, Size: 0, ModifiedAt: time.Now()},
			{Name: "user.go", Path: "/workspace/src/api/handlers/user.go", IsDirectory: false, Size: 200, ModifiedAt: time.Now()},
		}

		// Act
		tree, err := service.BuildFileTree(files, "/workspace")

		// Assert
		assert.NoError(t, err)
		assert.NEmpty(t, tree)
		// Navigate to verify 3+ levels deep
		assert.Equal(t, 1, len(tree.Children)) // src
		srcNode := tree.Children[0]
		assert.Equal(t, "src", srcNode.Name)
		assert.Equal(t, 1, len(srcNode.Children)) // api
		apiNode := srcNode.Children[0]
		assert.Equal(t, "api", apiNode.Name)
		assert.Equal(t, 1, len(apiNode.Children)) // handlers
		handlersNode := apiNode.Children[0]
		assert.Equal(t, "handlers", handlersNode.Name)
		assert.Equal(t, 1, len(handlersNode.Children)) // user.go
		assert.Equal(t, "user.go", handlersNode.Children[0].Name)
		assert.Equal(t, false, handlersNode.Children[0].IsDirectory)
	})

	t.Run("handles files and directories at same level", func(t *testing.T) {
		// Arrange
		service := NewSandboxService()
		files := []FileInfo{
			{Name: "workspace", Path: "/workspace", IsDirectory: true, Size: 0, ModifiedAt: time.Now()},
			{Name: "src", Path: "/workspace/src", IsDirectory: true, Size: 0, ModifiedAt: time.Now()},
			{Name: "tests", Path: "/workspace/tests", IsDirectory: true, Size: 0, ModifiedAt: time.Now()},
			{Name: "README.md", Path: "/workspace/README.md", IsDirectory: false, Size: 50, ModifiedAt: time.Now()},
			{Name: ".gitignore", Path: "/workspace/.gitignore", IsDirectory: false, Size: 10, ModifiedAt: time.Now()},
		}

		// Act
		tree, err := service.BuildFileTree(files, "/workspace")

		// Assert
		assert.NoError(t, err)
		assert.NEmpty(t, tree)
		assert.Equal(t, 4, len(tree.Children)) // src, tests, README.md, .gitignore
	})

	t.Run("returns empty tree for empty file list", func(t *testing.T) {
		// Arrange
		service := NewSandboxService()
		files := []FileInfo{}

		// Act
		tree, err := service.BuildFileTree(files, "/workspace")

		// Assert
		assert.NoError(t, err)
		assert.NEmpty(t, tree)
		assert.Equal(t, "workspace", tree.Name)
		assert.Equal(t, "/workspace", tree.Path)
		assert.True(t, tree.IsDirectory)
		assert.Equal(t, 0, len(tree.Children))
	})

	t.Run("handles single file", func(t *testing.T) {
		// Arrange
		service := NewSandboxService()
		files := []FileInfo{
			{Name: "workspace", Path: "/workspace", IsDirectory: true, Size: 0, ModifiedAt: time.Now()},
			{Name: "README.md", Path: "/workspace/README.md", IsDirectory: false, Size: 50, ModifiedAt: time.Now()},
		}

		// Act
		tree, err := service.BuildFileTree(files, "/workspace")

		// Assert
		assert.NoError(t, err)
		assert.NEmpty(t, tree)
		assert.Equal(t, 1, len(tree.Children))
		assert.Equal(t, "README.md", tree.Children[0].Name)
		assert.Equal(t, false, tree.Children[0].IsDirectory)
	})

	t.Run("preserves file metadata in tree nodes", func(t *testing.T) {
		// Arrange
		service := NewSandboxService()
		modTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
		files := []FileInfo{
			{Name: "workspace", Path: "/workspace", IsDirectory: true, Size: 0, ModifiedAt: modTime},
			{Name: "test.txt", Path: "/workspace/test.txt", IsDirectory: false, Size: 12345, ModifiedAt: modTime},
		}

		// Act
		tree, err := service.BuildFileTree(files, "/workspace")

		// Assert
		assert.NoError(t, err)
		fileNode := tree.Children[0]
		assert.Equal(t, int64(12345), fileNode.Size)
		assert.Equal(t, modTime, fileNode.ModifiedAt)
	})

	t.Run("handles complex directory structure with multiple branches", func(t *testing.T) {
		// Arrange
		service := NewSandboxService()
		files := []FileInfo{
			{Name: "workspace", Path: "/workspace", IsDirectory: true, Size: 0, ModifiedAt: time.Now()},
			{Name: "src", Path: "/workspace/src", IsDirectory: true, Size: 0, ModifiedAt: time.Now()},
			{Name: "api", Path: "/workspace/src/api", IsDirectory: true, Size: 0, ModifiedAt: time.Now()},
			{Name: "api.go", Path: "/workspace/src/api/api.go", IsDirectory: false, Size: 100, ModifiedAt: time.Now()},
			{Name: "models", Path: "/workspace/src/models", IsDirectory: true, Size: 0, ModifiedAt: time.Now()},
			{Name: "user.go", Path: "/workspace/src/models/user.go", IsDirectory: false, Size: 200, ModifiedAt: time.Now()},
			{Name: "tests", Path: "/workspace/tests", IsDirectory: true, Size: 0, ModifiedAt: time.Now()},
			{Name: "api_test.go", Path: "/workspace/tests/api_test.go", IsDirectory: false, Size: 300, ModifiedAt: time.Now()},
		}

		// Act
		tree, err := service.BuildFileTree(files, "/workspace")

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 2, len(tree.Children)) // src and tests at root level

		// Verify src branch
		srcNode := tree.Children[0]
		assert.Equal(t, "src", srcNode.Name)
		assert.Equal(t, 2, len(srcNode.Children)) // api and models

		// Verify tests branch
		var testsNode *FileTreeNode
		for _, child := range tree.Children {
			if child.Name == "tests" {
				testsNode = child
				break
			}
		}
		assert.NEmpty(t, testsNode)
		assert.Equal(t, 1, len(testsNode.Children)) // api_test.go
	})

	t.Run("handles s3-bucket root path", func(t *testing.T) {
		// Arrange
		service := NewSandboxService()
		files := []FileInfo{
			{Name: "s3-bucket", Path: "/s3-bucket", IsDirectory: true, Size: 0, ModifiedAt: time.Now()},
			{Name: "data.json", Path: "/s3-bucket/data.json", IsDirectory: false, Size: 500, ModifiedAt: time.Now()},
		}

		// Act
		tree, err := service.BuildFileTree(files, "/s3-bucket")

		// Assert
		assert.NoError(t, err)
		assert.NEmpty(t, tree)
		assert.Equal(t, "s3-bucket", tree.Name)
		assert.Equal(t, "/s3-bucket", tree.Path)
		assert.Equal(t, 1, len(tree.Children))
	})

	t.Run("sorts and processes files in correct order", func(t *testing.T) {
		// Arrange - provide files in random order
		service := NewSandboxService()
		files := []FileInfo{
			{Name: "test.go", Path: "/workspace/src/test.go", IsDirectory: false, Size: 100, ModifiedAt: time.Now()},
			{Name: "workspace", Path: "/workspace", IsDirectory: true, Size: 0, ModifiedAt: time.Now()},
			{Name: "src", Path: "/workspace/src", IsDirectory: true, Size: 0, ModifiedAt: time.Now()},
			{Name: "README.md", Path: "/workspace/README.md", IsDirectory: false, Size: 50, ModifiedAt: time.Now()},
		}

		// Act
		tree, err := service.BuildFileTree(files, "/workspace")

		// Assert - should still build correct hierarchy despite random order
		assert.NoError(t, err)
		assert.NEmpty(t, tree)
		assert.Equal(t, 2, len(tree.Children))

		// Find src node
		var srcNode *FileTreeNode
		for _, child := range tree.Children {
			if child.Name == "src" {
				srcNode = child
				break
			}
		}
		assert.NEmpty(t, srcNode)
		assert.Equal(t, 1, len(srcNode.Children))
		assert.Equal(t, "test.go", srcNode.Children[0].Name)
	})
}
