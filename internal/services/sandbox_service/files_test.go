package sandbox_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"

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
		sandboxModel.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)

		// Reconstruct sandbox info (without actual Modal sandbox)
		sandboxInfo := ReconstructSandboxInfo(sandboxModel, types.UUID("test-account"))

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
		sandboxModel.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)

		sandboxInfo := ReconstructSandboxInfo(sandboxModel, types.UUID("test-account"))

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
		sandboxModel.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)

		sandboxInfo := ReconstructSandboxInfo(sandboxModel, types.UUID("test-account"))

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
		sandboxModel.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)

		sandboxInfo := ReconstructSandboxInfo(sandboxModel, types.UUID("test-account"))

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
		sandboxModel.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)

		sandboxInfo := ReconstructSandboxInfo(sandboxModel, types.UUID("test-account"))

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
		sandboxModel.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)

		sandboxInfo := ReconstructSandboxInfo(sandboxModel, types.UUID("test-account"))

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
		sandboxModel.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)

		sandboxInfo := ReconstructSandboxInfo(sandboxModel, types.UUID("test-account"))

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
		sandboxModel.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)

		sandboxInfo := ReconstructSandboxInfo(sandboxModel, types.UUID("test-account"))

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
		sandboxModel.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)

		sandboxInfo := ReconstructSandboxInfo(sandboxModel, types.UUID("test-account"))

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
		sandboxModel.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)

		sandboxInfo := ReconstructSandboxInfo(sandboxModel, types.UUID("test-account"))

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
		sandboxModel.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)

		sandboxInfo := ReconstructSandboxInfo(sandboxModel, types.UUID("test-account"))

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
		sandboxModel.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)

		sandboxInfo := ReconstructSandboxInfo(sandboxModel, types.UUID("test-account"))

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
		sandboxModel.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)

		sandboxInfo := ReconstructSandboxInfo(sandboxModel, types.UUID("test-account"))

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
