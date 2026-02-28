package tarfs_test

import (
	"archive/tar"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"testing/fstest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/tarfs"
)

type errCloser struct {
	io.Reader
	closeErr error
}

func (e *errCloser) Close() error {
	return e.closeErr
}

var _ = Describe("Fs", func() {
	Describe("Open", func() {
		It("should open a tar file", func() {
			tfs, err := tarfs.Open("../testdata/test.tar")

			Expect(err).NotTo(HaveOccurred())
			Expect(tfs).NotTo(BeNil())
		})

		It("should return error for nonexistent file", func() {
			tfs, err := tarfs.Open("../testdata/nonexistent.tar")

			Expect(err).To(HaveOccurred())
			Expect(tfs).To(BeNil())
		})
	})

	Describe("Name", func() {
		It("should return the tar file name", func() {
			tfs, err := tarfs.Open("../testdata/test.tar")
			Expect(err).NotTo(HaveOccurred())

			name := tfs.Name()

			Expect(name).To(Equal("../testdata/test.tar"))
		})
	})

	Describe("Close", func() {
		It("should close the tar file", func() {
			tfs, err := tarfs.Open("../testdata/test.tar")
			Expect(err).NotTo(HaveOccurred())

			err = tfs.Close()
			Expect(err).NotTo(HaveOccurred())
		})

		It("should be idempotent", func() {
			tfs, err := tarfs.Open("../testdata/test.tar")
			Expect(err).NotTo(HaveOccurred())

			Expect(tfs.Close()).To(Succeed())
			Expect(tfs.Close()).To(Succeed())
		})

		It("should return error when opening file after Close", func() {
			tfs, err := tarfs.Open("../testdata/test.tar")
			Expect(err).NotTo(HaveOccurred())

			Expect(tfs.Close()).To(Succeed())

			file, err := tfs.Open("tartest/test.txt")
			Expect(err).To(MatchError(fs.ErrNotExist))
			Expect(file).To(BeNil())
		})
	})

	Describe("FromReader", func() {
		It("should create Fs from io.Reader", func() {
			var buf bytes.Buffer
			tw := tar.NewWriter(&buf)

			err := tw.WriteHeader(&tar.Header{
				Name: "test.txt",
				Mode: 0644,
				Size: 4,
			})
			Expect(err).NotTo(HaveOccurred())
			_, err = tw.Write([]byte("test"))
			Expect(err).NotTo(HaveOccurred())
			Expect(tw.Close()).To(Succeed())

			// Create from plain io.Reader (not io.ReadCloser)
			reader := bytes.NewReader(buf.Bytes())
			tfs := tarfs.FromReader("test.tar", reader)

			Expect(tfs).NotTo(BeNil())
			Expect(tfs.Name()).To(Equal("test.tar"))

			file, err := tfs.Open("test.txt")
			Expect(err).NotTo(HaveOccurred())
			content, err := io.ReadAll(file)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal("test"))
		})

		It("should create Fs from io.ReadCloser", func() {
			file, err := os.Open("../testdata/test.tar")
			Expect(err).NotTo(HaveOccurred())
			defer file.Close()

			tfs := tarfs.FromReader("test.tar", file)

			Expect(tfs).NotTo(BeNil())
			Expect(tfs.Name()).To(Equal("test.tar"))
		})
	})

	Describe("Open file", func() {
		var tfs *tarfs.TarFile

		BeforeEach(func() {
			var err error
			tfs, err = tarfs.Open("../testdata/test.tar")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should open a file from the tar archive", func() {
			file, err := tfs.Open("tartest/test.txt")

			Expect(err).NotTo(HaveOccurred())
			Expect(file).NotTo(BeNil())
		})

		It("should read file contents", func() {
			file, err := tfs.Open("tartest/test.txt")
			Expect(err).NotTo(HaveOccurred())

			content, err := io.ReadAll(file)

			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal("test content\n"))
		})

		It("should return error for nonexistent file in tar", func() {
			file, err := tfs.Open("nonexistent.txt")

			Expect(err).To(MatchError(fs.ErrNotExist))
			Expect(file).To(BeNil())
		})

		It("should return cached file on subsequent opens", func() {
			file1, err := tfs.Open("tartest/test.txt")
			Expect(err).NotTo(HaveOccurred())

			file2, err := tfs.Open("tartest/test.txt")
			Expect(err).NotTo(HaveOccurred())

			Expect(file2).To(Equal(file1))
		})

		It("should open multiple files from the archive", func() {
			file1, err := tfs.Open("tartest/test.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(file1).NotTo(BeNil())

			file2, err := tfs.Open("tartest/another.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(file2).NotTo(BeNil())

			content1, err := io.ReadAll(file1)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content1)).To(Equal("test content\n"))

			content2, err := io.ReadAll(file2)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content2)).To(Equal("another file\n"))
		})
	})

	Describe("OpenFS", func() {
		It("should open a tar file from a custom FS", func() {
			tfs, err := tarfs.OpenFS(os.DirFS(".."), "testdata/test.tar")

			Expect(err).NotTo(HaveOccurred())
			Expect(tfs).NotTo(BeNil())
		})

		It("should read files from tar opened with custom FS", func() {
			tfs, err := tarfs.OpenFS(os.DirFS(".."), "testdata/test.tar")
			Expect(err).NotTo(HaveOccurred())

			file, err := tfs.Open("tartest/test.txt")
			Expect(err).NotTo(HaveOccurred())

			content, err := io.ReadAll(file)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal("test content\n"))
		})
	})

	Describe("error handling", func() {
		It("should handle broken tar with incomplete content", func() {
			var buf bytes.Buffer
			tw := tar.NewWriter(&buf)
			tw.WriteHeader(&tar.Header{
				Name: "test.txt",
				Size: 1000,
				Mode: 0600,
			})
			tw.Write([]byte("short"))

			tmpDir := GinkgoT().TempDir()
			testPath := tmpDir + "/incomplete.tar"
			err := os.WriteFile(testPath, buf.Bytes(), 0644)
			Expect(err).NotTo(HaveOccurred())

			tfs, err := tarfs.Open(testPath)
			Expect(err).NotTo(HaveOccurred())

			file, err := tfs.Open("test.txt")
			Expect(err).To(MatchError(fs.ErrNotExist))
			Expect(err).To(MatchError(io.ErrUnexpectedEOF))
			Expect(file).To(BeNil())
		})

		It("should format TarError correctly", func() {
			err := &tarfs.TarError{
				Archive: "test.tar",
				Name:    "test.txt",
				Err:     fs.ErrNotExist,
				Cause:   io.ErrUnexpectedEOF,
			}

			Expect(err.Error()).To(Equal("test.tar(test.txt): file does not exist: unexpected EOF"))
		})

		It("should handle close error when reaching EOF", func() {
			var buf bytes.Buffer
			tw := tar.NewWriter(&buf)

			err := tw.WriteHeader(&tar.Header{
				Name: "file1.txt",
				Mode: 0644,
				Size: 5,
			})
			Expect(err).NotTo(HaveOccurred())
			_, err = tw.Write([]byte("data1"))
			Expect(err).NotTo(HaveOccurred())
			Expect(tw.Close()).To(Succeed())

			closeErr := errors.New("close failed")
			reader := &errCloser{
				Reader:   bytes.NewReader(buf.Bytes()),
				closeErr: closeErr,
			}

			tfs := tarfs.FromReader("test.tar", reader)

			file, err := tfs.Open("nonexistent.txt")

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(fs.ErrNotExist))
			Expect(err).To(MatchError(closeErr))
			Expect(file).To(BeNil())
		})
	})

	Describe("lazy loading", func() {
		It("should only read tar entries as needed", func() {
			var buf bytes.Buffer
			tw := tar.NewWriter(&buf)

			// Write multiple files
			for i := 1; i <= 5; i++ {
				num := fmt.Sprintf("%d", i)
				name := "file" + num + ".txt"
				content := "content " + num

				err := tw.WriteHeader(&tar.Header{
					Name: name,
					Mode: 0644,
					Size: int64(len(content)),
				})
				Expect(err).NotTo(HaveOccurred())

				_, err = tw.Write([]byte(content))
				Expect(err).NotTo(HaveOccurred())
			}
			Expect(tw.Close()).To(Succeed())

			tmpDir := GinkgoT().TempDir()
			testPath := tmpDir + "/lazy.tar"
			err := os.WriteFile(testPath, buf.Bytes(), 0644)
			Expect(err).NotTo(HaveOccurred())

			tfs, err := tarfs.Open(testPath)
			Expect(err).NotTo(HaveOccurred())

			// Open first file - should only read up to it
			file1, err := tfs.Open("file1.txt")
			Expect(err).NotTo(HaveOccurred())
			content1, err := io.ReadAll(file1)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content1)).To(Equal("content 1"))

			// Open last file - should read through entire archive
			file5, err := tfs.Open("file5.txt")
			Expect(err).NotTo(HaveOccurred())
			content5, err := io.ReadAll(file5)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content5)).To(Equal("content 5"))

			// Open middle file - should be cached now
			file3, err := tfs.Open("file3.txt")
			Expect(err).NotTo(HaveOccurred())
			content3, err := io.ReadAll(file3)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content3)).To(Equal("content 3"))
		})
	})

	Describe("directory handling", func() {
		var tfs *tarfs.TarFile
		var testPath string

		BeforeEach(func() {
			var buf bytes.Buffer
			tw := tar.NewWriter(&buf)

			err := tw.WriteHeader(&tar.Header{
				Name:     "mydir",
				Mode:     0755,
				Typeflag: tar.TypeDir,
			})
			Expect(err).NotTo(HaveOccurred())

			err = tw.WriteHeader(&tar.Header{
				Name: "mydir/file.txt",
				Mode: 0644,
				Size: 14,
			})
			Expect(err).NotTo(HaveOccurred())
			_, err = tw.Write([]byte("file in subdir"))
			Expect(err).NotTo(HaveOccurred())

			err = tw.WriteHeader(&tar.Header{
				Name:     "emptydir",
				Mode:     0755,
				Typeflag: tar.TypeDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(tw.Close()).To(Succeed())

			tmpDir := GinkgoT().TempDir()
			testPath = tmpDir + "/test-with-dirs.tar"
			err = os.WriteFile(testPath, buf.Bytes(), 0644)
			Expect(err).NotTo(HaveOccurred())

			tfs, err = tarfs.Open(testPath)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should open a directory entry", func() {
			file, err := tfs.Open("mydir")

			Expect(err).NotTo(HaveOccurred())
			Expect(file).NotTo(BeNil())
			Expect(file.Close()).To(Succeed())
		})

		It("should return directory info for directory entry", func() {
			file, err := tfs.Open("mydir")
			Expect(err).NotTo(HaveOccurred())
			defer file.Close()

			info, err := file.Stat()

			Expect(err).NotTo(HaveOccurred())
			Expect(info.IsDir()).To(BeTrue())
			Expect(info.Name()).To(Equal("mydir"))
		})

		It("should open an empty directory", func() {
			file, err := tfs.Open("emptydir")

			Expect(err).NotTo(HaveOccurred())
			Expect(file).NotTo(BeNil())
			defer file.Close()

			info, err := file.Stat()
			Expect(err).NotTo(HaveOccurred())
			Expect(info.IsDir()).To(BeTrue())
		})

		It("should open files within directories", func() {
			file, err := tfs.Open("mydir/file.txt")

			Expect(err).NotTo(HaveOccurred())
			Expect(file).NotTo(BeNil())

			content, err := io.ReadAll(file)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal("file in subdir"))
		})

		It("should return error when reading from directory", func() {
			file, err := tfs.Open("mydir")
			Expect(err).NotTo(HaveOccurred())

			buf := make([]byte, 10)
			n, err := file.Read(buf)

			Expect(n).To(Equal(0))
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(fs.ErrInvalid))
		})
	})

	Describe("synthetic directory handling", func() {
		var tfs *tarfs.TarFile

		BeforeEach(func() {
			var err error
			tfs, err = tarfs.Open("../testdata/test.tar")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return error when reading from synthetic directory", func() {
			file, err := tfs.Open("tartest")
			Expect(err).NotTo(HaveOccurred())
			defer file.Close()

			buf := make([]byte, 10)
			n, err := file.Read(buf)

			Expect(n).To(Equal(0))
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(fs.ErrInvalid))
		})

		It("should return error when reading from root directory", func() {
			file, err := tfs.Open(".")
			Expect(err).NotTo(HaveOccurred())
			defer file.Close()

			buf := make([]byte, 10)
			n, err := file.Read(buf)

			Expect(n).To(Equal(0))
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(fs.ErrInvalid))
		})

		It("should return nil from Sys() on directory FileInfo", func() {
			file, err := tfs.Open("tartest")
			Expect(err).NotTo(HaveOccurred())
			defer file.Close()

			info, err := file.Stat()
			Expect(err).NotTo(HaveOccurred())
			Expect(info.Sys()).To(BeNil())
		})

		It("should handle paginated ReadDir when requesting more than available", func() {
			file, err := tfs.Open("tartest")
			Expect(err).NotTo(HaveOccurred())
			defer file.Close()

			// Cast to ReadDirFile to access ReadDir
			rdFile, ok := file.(fs.ReadDirFile)
			Expect(ok).To(BeTrue())

			// Request more entries than exist (tartest only has 2 files)
			entries, err := rdFile.ReadDir(10)
			Expect(err).NotTo(HaveOccurred())
			Expect(entries).To(HaveLen(2))
		})

		It("should handle error when opening root with corrupted tar", func() {
			var buf bytes.Buffer
			tw := tar.NewWriter(&buf)

			// Write a valid header but corrupt the data
			err := tw.WriteHeader(&tar.Header{
				Name: "file.txt",
				Mode: 0644,
				Size: 1000, // Claim large size
			})
			Expect(err).NotTo(HaveOccurred())
			// Write only a few bytes, making it incomplete
			_, err = tw.Write([]byte("short"))
			Expect(err).NotTo(HaveOccurred())
			// Don't close tw properly to create corrupt archive

			tmpDir := GinkgoT().TempDir()
			testPath := tmpDir + "/corrupt.tar"
			err = os.WriteFile(testPath, buf.Bytes(), 0644)
			Expect(err).NotTo(HaveOccurred())

			corruptTfs, err := tarfs.Open(testPath)
			Expect(err).NotTo(HaveOccurred())

			// Opening root should try to load all entries and hit the corruption
			file, err := corruptTfs.Open(".")
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(ihfs.ErrInvalid))
			Expect(file).To(BeNil())
		})
	})

	Describe("concurrent access", func() {
		It("should handle concurrent Open calls on different files", func() {
			tfs, err := tarfs.Open("../testdata/test.tar")
			Expect(err).NotTo(HaveOccurred())

			done := make(chan bool)
			const goroutines = 10

			for range goroutines {
				go func() {
					defer GinkgoRecover()
					file, err := tfs.Open("tartest/test.txt")
					Expect(err).NotTo(HaveOccurred())
					Expect(file).NotTo(BeNil())
					defer file.Close()

					content, err := io.ReadAll(file)
					Expect(err).NotTo(HaveOccurred())
					Expect(string(content)).To(Equal("test content\n"))

					done <- true
				}()
			}

			for range goroutines {
				<-done
			}
		})

		It("should handle concurrent Open calls on the same file", func() {
			tfs, err := tarfs.Open("../testdata/test.tar")
			Expect(err).NotTo(HaveOccurred())

			done := make(chan bool)
			const goroutines = 20

			for range goroutines {
				go func() {
					defer GinkgoRecover()
					file, err := tfs.Open("tartest/test.txt")
					Expect(err).NotTo(HaveOccurred())
					Expect(file).NotTo(BeNil())
					defer file.Close()
					done <- true
				}()
			}

			for range goroutines {
				<-done
			}
		})

		It("should handle concurrent Open calls on multiple files", func() {
			tfs, err := tarfs.Open("../testdata/test.tar")
			Expect(err).NotTo(HaveOccurred())

			done := make(chan bool)
			const goroutines = 20

			for i := range goroutines {
				fileName := "tartest/test.txt"
				if i%2 == 0 {
					fileName = "tartest/another.txt"
				}

				go func(name string) {
					defer GinkgoRecover()
					file, err := tfs.Open(name)
					Expect(err).NotTo(HaveOccurred())
					Expect(file).NotTo(BeNil())
					defer file.Close()
					done <- true
				}(fileName)
			}

			for range goroutines {
				<-done
			}
		})

		It("should handle concurrent reads from cached files", func() {
			tfs, err := tarfs.Open("../testdata/test.tar")
			Expect(err).NotTo(HaveOccurred())

			// Pre-cache the file
			_, err = tfs.Open("tartest/test.txt")
			Expect(err).NotTo(HaveOccurred())

			done := make(chan bool)
			const goroutines = 20

			for range goroutines {
				go func() {
					defer GinkgoRecover()
					// This should hit the cache
					file, err := tfs.Open("tartest/test.txt")
					Expect(err).NotTo(HaveOccurred())
					Expect(file).NotTo(BeNil())
					defer file.Close()

					content, err := io.ReadAll(file)
					Expect(err).NotTo(HaveOccurred())
					Expect(string(content)).To(Equal("test content\n"))

					done <- true
				}()
			}

			for range goroutines {
				<-done
			}
		})

		It("should handle race when one goroutine hits EOF and closes", func() {
			var buf bytes.Buffer
			tw := tar.NewWriter(&buf)

			err := tw.WriteHeader(&tar.Header{
				Name: "file1.txt",
				Mode: 0644,
				Size: 5,
			})
			Expect(err).NotTo(HaveOccurred())
			_, err = tw.Write([]byte("data1"))
			Expect(err).NotTo(HaveOccurred())

			err = tw.WriteHeader(&tar.Header{
				Name: "file2.txt",
				Mode: 0644,
				Size: 5,
			})
			Expect(err).NotTo(HaveOccurred())
			_, err = tw.Write([]byte("data2"))
			Expect(err).NotTo(HaveOccurred())
			Expect(tw.Close()).To(Succeed())

			tmpDir := GinkgoT().TempDir()
			testPath := tmpDir + "/small.tar"
			err = os.WriteFile(testPath, buf.Bytes(), 0644)
			Expect(err).NotTo(HaveOccurred())

			tfs, err := tarfs.Open(testPath)
			Expect(err).NotTo(HaveOccurred())

			done := make(chan error, 10)
			const goroutines = 10

			for i := range goroutines {
				go func(idx int) {
					defer GinkgoRecover()
					fileName := "nonexistent.txt"
					if idx < 2 {
						fileName = fmt.Sprintf("file%d.txt", idx+1)
					}
					_, err := tfs.Open(fileName)
					done <- err
				}(i)
			}

			var closedErrors, notExistErrors int
			for range goroutines {
				err := <-done
				if err != nil {
					if errors.Is(err, fs.ErrClosed) {
						closedErrors++
					} else if errors.Is(err, fs.ErrNotExist) {
						notExistErrors++
					}
				}
			}

			Expect(closedErrors).To(Equal(7))
			Expect(notExistErrors).To(Equal(1))
		})
	})

	var tfs *tarfs.TarFile

	BeforeEach(func() {
		var err error
		tfs, err = tarfs.Open("../testdata/test.tar")
		Expect(err).NotTo(HaveOccurred())
	})

	Context("directory entries implement ReadDirFile", func() {
		It("should allow casting directory to fs.ReadDirFile", func() {
			file, err := tfs.Open("tartest")
			Expect(err).NotTo(HaveOccurred())
			defer file.Close()

			rdFile, ok := file.(fs.ReadDirFile)
			Expect(ok).To(BeTrue(), "directory should implement fs.ReadDirFile")

			entries, err := rdFile.ReadDir(-1)
			Expect(err).NotTo(HaveOccurred())
			Expect(entries).To(HaveLen(2))
		})

		It("should allow casting explicit directory entry to fs.ReadDirFile", func() {
			file, err := tfs.Open(".")
			Expect(err).NotTo(HaveOccurred())
			defer file.Close()

			rdFile, ok := file.(fs.ReadDirFile)
			Expect(ok).To(BeTrue(), "root directory should implement fs.ReadDirFile")

			entries, err := rdFile.ReadDir(-1)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(entries)).To(BeNumerically(">=", 1), "root should contain at least tartest")
		})
	})

	Context("FileInfo.Name returns base name only", func() {
		It("should return base name for nested synthetic directory", func() {
			file, err := tfs.Open("tartest")
			Expect(err).NotTo(HaveOccurred())
			defer file.Close()

			info, err := file.Stat()
			Expect(err).NotTo(HaveOccurred())
			Expect(info.Name()).To(Equal("tartest"), "should be base name, not full path")
		})

		It("should return base name for synthetic subdirectory entries", func() {
			file, err := tfs.Open("tartest")
			Expect(err).NotTo(HaveOccurred())
			defer file.Close()

			rdFile := file.(fs.ReadDirFile)
			entries, err := rdFile.ReadDir(-1)
			Expect(err).NotTo(HaveOccurred())

			for _, entry := range entries {
				name := entry.Name()
				Expect(name).NotTo(ContainSubstring("/"), "entry name should not contain path separator")

				info, err := entry.Info()
				Expect(err).NotTo(HaveOccurred())
				Expect(info.Name()).To(Equal(name), "Info().Name() should match entry.Name()")
			}
		})
	})

	Context("resource cleanup", func() {
		It("should close tar reader when opening root directory", func() {
			tfs, err := tarfs.Open("../testdata/test.tar")
			Expect(err).NotTo(HaveOccurred())

			// Open root directory which reads entire tar
			f, err := tfs.Open(".")
			Expect(err).NotTo(HaveOccurred())
			Expect(f.Close()).To(Succeed())

			// Explicitly close the tarfs - should not leak
			err = tfs.Close()
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("directory with trailing slash", func() {
		It("should detect non-synthetic directories with trailing slash in tar", func() {
			// Create a tar with explicit directory entry ending in "/"
			var buf bytes.Buffer
			tw := tar.NewWriter(&buf)

			err := tw.WriteHeader(&tar.Header{
				Name:     "mydir/",
				Typeflag: tar.TypeDir,
				Mode:     0755,
			})
			Expect(err).NotTo(HaveOccurred())

			// Write file under directory
			err = tw.WriteHeader(&tar.Header{
				Name: "mydir/file.txt",
				Size: 5,
			})
			Expect(err).NotTo(HaveOccurred())
			_, err = tw.Write([]byte("hello"))
			Expect(err).NotTo(HaveOccurred())

			err = tw.Close()
			Expect(err).NotTo(HaveOccurred())

			tfs := tarfs.FromReader("test.tar", bytes.NewReader(buf.Bytes()))

			f, err := tfs.Open("mydir")
			Expect(err).NotTo(HaveOccurred())
			defer f.Close()

			info, err := f.Stat()
			Expect(err).NotTo(HaveOccurred())
			Expect(info.IsDir()).To(BeTrue())
			// Should have non-nil Sys since it's in the tar
			Expect(info.Sys()).NotTo(BeNil())
		})
	})

	Context("error message clarity", func() {
		It("should not include nil in error messages", func() {
			tfs, err := tarfs.Open("../testdata/test.tar")
			Expect(err).NotTo(HaveOccurred())

			// Try to open with invalid path
			_, err = tfs.Open("../invalid")
			Expect(err).To(HaveOccurred())

			// Error message should not contain ": <nil>"
			Expect(err.Error()).NotTo(ContainSubstring(": <nil>"))
		})
	})

	It("should return error when calling ReadDir on a regular file", func() {
		file, err := tfs.Open("tartest/test.txt")
		Expect(err).NotTo(HaveOccurred())
		defer file.Close()

		// Try to cast to ReadDirFile
		rdFile, ok := file.(fs.ReadDirFile)
		Expect(ok).To(BeTrue(), "File should implement ReadDirFile interface")

		// But calling ReadDir should fail
		_, err = rdFile.ReadDir(-1)
		Expect(err).To(HaveOccurred())
		Expect(err).To(MatchError(ContainSubstring("invalid argument")))
	})

	It("should return Sys() for real tar entries", func() {
		file, err := tfs.Open("tartest/test.txt")
		Expect(err).NotTo(HaveOccurred())
		defer file.Close()

		info, err := file.Stat()
		Expect(err).NotTo(HaveOccurred())

		// Real tar entry should have non-nil Sys
		Expect(info.Sys()).NotTo(BeNil())
	})

	It("should skip entries outside directory prefix", func() {
		// Create tar with multiple directories
		var buf bytes.Buffer
		tw := tar.NewWriter(&buf)

		// dir1/file.txt
		err := tw.WriteHeader(&tar.Header{Name: "dir1/file.txt", Size: 5})
		Expect(err).NotTo(HaveOccurred())
		_, err = tw.Write([]byte("hello"))
		Expect(err).NotTo(HaveOccurred())

		// dir2/file.txt
		err = tw.WriteHeader(&tar.Header{Name: "dir2/file.txt", Size: 5})
		Expect(err).NotTo(HaveOccurred())
		_, err = tw.Write([]byte("world"))
		Expect(err).NotTo(HaveOccurred())

		Expect(tw.Close()).To(Succeed())

		tfs := tarfs.FromReader("test.tar", bytes.NewReader(buf.Bytes()))

		// Open dir1 and verify it only shows its own files
		file, err := tfs.Open("dir1")
		Expect(err).NotTo(HaveOccurred())
		defer file.Close()

		rdFile := file.(fs.ReadDirFile)
		entries, err := rdFile.ReadDir(-1)
		Expect(err).NotTo(HaveOccurred())
		Expect(entries).To(HaveLen(1))
		Expect(entries[0].Name()).To(Equal("file.txt"))
	})

	It("should return nil Sys() for synthetic directories", func() {
		file, err := tfs.Open("tartest")
		Expect(err).NotTo(HaveOccurred())
		defer file.Close()

		info, err := file.Stat()
		Expect(err).NotTo(HaveOccurred())

		// Synthetic directory should have nil Sys
		Expect(info.Sys()).To(BeNil())
	})

	It("should handle root directory with Open(.)", func() {
		// fs.ValidPath rejects ".", so Open must handle it before validation
		file, err := tfs.Open(".")
		Expect(err).NotTo(HaveOccurred())
		defer file.Close()

		info, err := file.Stat()
		Expect(err).NotTo(HaveOccurred())
		Expect(info.IsDir()).To(BeTrue())
		Expect(info.Name()).To(Equal("."))
	})

	It("should handle directories with trailing slashes in tar", func() {
		// Create tar with directory entry ending in "/"
		var buf bytes.Buffer
		tw := tar.NewWriter(&buf)

		// Add directory with trailing slash
		err := tw.WriteHeader(&tar.Header{
			Name:     "mydir/",
			Typeflag: tar.TypeDir,
			Mode:     0755,
		})
		Expect(err).NotTo(HaveOccurred())

		// Add file in directory
		err = tw.WriteHeader(&tar.Header{Name: "mydir/file.txt", Size: 5})
		Expect(err).NotTo(HaveOccurred())
		_, err = tw.Write([]byte("hello"))
		Expect(err).NotTo(HaveOccurred())

		Expect(tw.Close()).To(Succeed())

		tfs := tarfs.FromReader("test.tar", bytes.NewReader(buf.Bytes()))

		// Should be able to open "mydir" (without trailing slash)
		file, err := tfs.Open("mydir")
		Expect(err).NotTo(HaveOccurred())
		defer file.Close()

		info, err := file.Stat()
		Expect(err).NotTo(HaveOccurred())
		Expect(info.IsDir()).To(BeTrue())
		// Should have non-nil Sys since it's a real tar entry
		Expect(info.Sys()).NotTo(BeNil())
	})

	Describe("fstest", func() {
		It("should pass fstest.TestFS", func() {
			tfs, err := tarfs.Open("../testdata/test.tar")
			Expect(err).NotTo(HaveOccurred())

			err = fstest.TestFS(tfs, "tartest/test.txt", "tartest/another.txt")
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
