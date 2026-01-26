package tarfs_test

import (
	"archive/tar"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

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

		It("should handle corrupt tar header", func() {
			// Create an invalid tar that causes tar.Reader.Next() to fail
			var buf bytes.Buffer
			tw := tar.NewWriter(&buf)

			// Write a valid entry
			err := tw.WriteHeader(&tar.Header{
				Name: "file1.txt",
				Mode: 0644,
				Size: 5,
			})
			Expect(err).NotTo(HaveOccurred())
			_, err = tw.Write([]byte("data1"))
			Expect(err).NotTo(HaveOccurred())

			// Write another partial/corrupt entry - append invalid tar header bytes
			// A tar header is 512 bytes, add garbage that looks like a header but has bad checksum
			invalidHeader := make([]byte, 512)
			copy(invalidHeader[:100], "file2.txt")  // Name field
			invalidHeader[156] = '0'  // Type flag for regular file
			// Leave checksum field (offset 148-155) as zeros, which will be invalid
			buf.Write(invalidHeader)
			
			tmpDir := GinkgoT().TempDir()
			testPath := tmpDir + "/corrupt.tar"
			err = os.WriteFile(testPath, buf.Bytes(), 0644)
			Expect(err).NotTo(HaveOccurred())

			tfs, err := tarfs.Open(testPath)
			Expect(err).NotTo(HaveOccurred())

			// Try to open a file - this should cause Next() to fail with checksum error
			file, err := tfs.Open("file2.txt")
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(fs.ErrNotExist))
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

			// Open middle file - will fail because archive is exhausted
			// and file3 was not cached since it wasn't directly opened
			file3, err := tfs.Open("file3.txt")
			Expect(err).To(MatchError(fs.ErrNotExist))
			Expect(file3).To(BeNil())
		})

		It("should cache entries while looking for a file", func() {
			var buf bytes.Buffer
			tw := tar.NewWriter(&buf)

			// Write files file1, file2, file3
			for i := 1; i <= 3; i++ {
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
			testPath := tmpDir + "/cache-test.tar"
			err := os.WriteFile(testPath, buf.Bytes(), 0644)
			Expect(err).NotTo(HaveOccurred())

			tfs, err := tarfs.Open(testPath)
			Expect(err).NotTo(HaveOccurred())

			// Open file3 - this will read file1, file2, and file3,
			// but only cache file3 (lazy caching)
			file3, err := tfs.Open("file3.txt")
			Expect(err).NotTo(HaveOccurred())
			content3, err := io.ReadAll(file3)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content3)).To(Equal("content 3"))

			// Now file3 is cached but file1 and file2 are not
			// Try to open file1 - should fail since tar reader is exhausted
			file1, err := tfs.Open("file1.txt")
			Expect(err).To(MatchError(fs.ErrNotExist))
			Expect(file1).To(BeNil())

			// But file3 should still be accessible from cache
			file3Again, err := tfs.Open("file3.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(file3Again).NotTo(BeNil())
		})
	})

	Describe("directory handling", func() {
		var tfs *tarfs.TarFile
		var testPath string

		BeforeEach(func() {
			var buf bytes.Buffer
			tw := tar.NewWriter(&buf)

			err := tw.WriteHeader(&tar.Header{
				Name:     "mydir/",
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
				Name:     "emptydir/",
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
			file, err := tfs.Open("mydir/")

			Expect(err).NotTo(HaveOccurred())
			Expect(file).NotTo(BeNil())
			Expect(file.Close()).To(Succeed())
		})

		It("should return directory info for directory entry", func() {
			file, err := tfs.Open("mydir/")
			Expect(err).NotTo(HaveOccurred())
			defer file.Close()

			info, err := file.Stat()

			Expect(err).NotTo(HaveOccurred())
			Expect(info.IsDir()).To(BeTrue())
			Expect(info.Name()).To(Equal("mydir"))
		})

		It("should open an empty directory", func() {
			file, err := tfs.Open("emptydir/")

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

		It("should read empty content from directory", func() {
			file, err := tfs.Open("mydir/")
			Expect(err).NotTo(HaveOccurred())

			buf := make([]byte, 10)
			n, err := file.Read(buf)

			Expect(n).To(Equal(0))
			Expect(err).To(Equal(io.EOF))
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

			var closedErrors, notExistErrors, successCount int
			for range goroutines {
				err := <-done
				if err != nil {
					if errors.Is(err, fs.ErrClosed) {
						closedErrors++
					} else if errors.Is(err, fs.ErrNotExist) {
						notExistErrors++
					}
				} else {
					successCount++
				}
			}

			// With lazy caching, only directly accessed files are cached
			// All operations should complete (either succeed or fail)
			Expect(successCount + closedErrors + notExistErrors).To(Equal(10))
			// At least one goroutine should hit an error (closed or not found)
			Expect(closedErrors + notExistErrors).To(BeNumerically(">", 0))
		})

		It("should skip over non-matching files while searching", func() {
			var buf bytes.Buffer
			tw := tar.NewWriter(&buf)

			// Write files with different names
			files := []string{"aaa.txt", "bbb.txt", "zzz.txt"}
			for _, name := range files {
				content := "content-" + name

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
			testPath := tmpDir + "/skip.tar"
			err := os.WriteFile(testPath, buf.Bytes(), 0644)
			Expect(err).NotTo(HaveOccurred())

			tfs, err := tarfs.Open(testPath)
			Expect(err).NotTo(HaveOccurred())

			// Open the middle file - should skip aaa.txt, find bbb.txt
			file, err := tfs.Open("bbb.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(file).NotTo(BeNil())

			content, err := io.ReadAll(file)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal("content-bbb.txt"))
		})

		It("should handle double-check pattern when file is cached by another goroutine", func() {
			var buf bytes.Buffer
			tw := tar.NewWriter(&buf)

			err := tw.WriteHeader(&tar.Header{
				Name: "file.txt",
				Mode: 0644,
				Size: 7,
			})
			Expect(err).NotTo(HaveOccurred())
			_, err = tw.Write([]byte("content"))
			Expect(err).NotTo(HaveOccurred())
			Expect(tw.Close()).To(Succeed())

			tmpDir := GinkgoT().TempDir()
			testPath := tmpDir + "/double-check.tar"
			err = os.WriteFile(testPath, buf.Bytes(), 0644)
			Expect(err).NotTo(HaveOccurred())

			tfs, err := tarfs.Open(testPath)
			Expect(err).NotTo(HaveOccurred())

			done := make(chan bool, 2)

			// Start two goroutines that try to open the same file simultaneously
			for range 2 {
				go func() {
					defer GinkgoRecover()
					file, err := tfs.Open("file.txt")
					Expect(err).NotTo(HaveOccurred())
					Expect(file).NotTo(BeNil())
					done <- true
				}()
			}

			// Both should succeed - one will load it, the other will get it from cache
			<-done
			<-done
		})
	})
})
