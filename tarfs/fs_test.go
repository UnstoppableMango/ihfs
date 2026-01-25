package tarfs_test

import (
	"archive/tar"
	"bytes"
	"io"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs/tarfs"
)

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

	Describe("Open file", func() {
		var tfs *tarfs.Fs

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

			Expect(err).To(MatchError(ContainSubstring("file does not exist")))
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
			Expect(err).To(MatchError(io.ErrUnexpectedEOF))
			Expect(file).To(BeNil())
		})
	})

	Describe("directory handling", func() {
		var tfs *tarfs.Fs
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
	})
})
