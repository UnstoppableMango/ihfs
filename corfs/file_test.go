package corfs_test

import (
	"errors"
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs"
	"github.com/unstoppablemango/ihfs/corfs"
	"github.com/unstoppablemango/ihfs/testfs"
)

var _ = Describe("File", func() {
	Describe("Close", func() {
		It("should close both base and layer files", func() {
			baseClosed := false
			layerClosed := false

			baseFile := &testfs.File{
				CloseFunc: func() error {
					baseClosed = true
					return nil
				},
			}
			layerFile := &testfs.File{
				CloseFunc: func() error {
					layerClosed = true
					return nil
				},
			}

			file := corfs.NewFile(baseFile, layerFile)
			err := file.Close()

			Expect(err).ToNot(HaveOccurred())
			Expect(baseClosed).To(BeTrue())
			Expect(layerClosed).To(BeTrue())
		})

		It("should return error when both files are nil", func() {
			file := corfs.NewFile(nil, nil)
			err := file.Close()
			Expect(err).To(HaveOccurred())
		})

		It("should return joined errors when both fail", func() {
			baseFile := &testfs.File{
				CloseFunc: func() error {
					return errors.New("base error")
				},
			}
			layerFile := &testfs.File{
				CloseFunc: func() error {
					return errors.New("layer error")
				},
			}

			file := corfs.NewFile(baseFile, layerFile)
			err := file.Close()

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("base error"))
			Expect(err.Error()).To(ContainSubstring("layer error"))
		})
	})

	Describe("Read", func() {
		It("should read from layer if available", func() {
			baseFile := &testfs.File{
				ReadFunc: func(p []byte) (int, error) {
					return copy(p, []byte("base")), io.EOF
				},
			}
			layerFile := &testfs.File{
				ReadFunc: func(p []byte) (int, error) {
					return copy(p, []byte("layer")), io.EOF
				},
			}

			file := corfs.NewFile(baseFile, layerFile)
			buf := make([]byte, 100)
			n, _ := file.Read(buf)

			Expect(string(buf[:n])).To(Equal("layer"))
		})

		It("should read from base if layer is nil", func() {
			baseFile := &testfs.File{
				ReadFunc: func(p []byte) (int, error) {
					return copy(p, []byte("base")), io.EOF
				},
			}

			file := corfs.NewFile(baseFile, nil)
			buf := make([]byte, 100)
			n, _ := file.Read(buf)

			Expect(string(buf[:n])).To(Equal("base"))
		})

		It("should return error when both files are nil", func() {
			file := corfs.NewFile(nil, nil)
			buf := make([]byte, 100)
			_, err := file.Read(buf)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Stat", func() {
		It("should stat layer if available", func() {
			baseFile := &testfs.File{
				StatFunc: func() (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo()
					fi.NameFunc = func() string { return "base.txt" }
					return fi, nil
				},
			}
			layerFile := &testfs.File{
				StatFunc: func() (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo()
					fi.NameFunc = func() string { return "layer.txt" }
					return fi, nil
				},
			}

			file := corfs.NewFile(baseFile, layerFile)
			info, err := file.Stat()

			Expect(err).ToNot(HaveOccurred())
			Expect(info.Name()).To(Equal("layer.txt"))
		})

		It("should stat base if layer is nil", func() {
			baseFile := &testfs.File{
				StatFunc: func() (ihfs.FileInfo, error) {
					fi := testfs.NewFileInfo()
					fi.NameFunc = func() string { return "base.txt" }
					return fi, nil
				},
			}

			file := corfs.NewFile(baseFile, nil)
			info, err := file.Stat()

			Expect(err).ToNot(HaveOccurred())
			Expect(info.Name()).To(Equal("base.txt"))
		})

		It("should return error when both files are nil", func() {
			file := corfs.NewFile(nil, nil)
			_, err := file.Stat()
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("ReadDir", func() {
		It("should merge entries from both base and layer", func() {
			baseEntry1 := &testfs.DirEntry{NameFunc: func() string { return "base1.txt" }}
			baseEntry2 := &testfs.DirEntry{NameFunc: func() string { return "common.txt" }}
			layerEntry1 := &testfs.DirEntry{NameFunc: func() string { return "layer1.txt" }}
			layerEntry2 := &testfs.DirEntry{NameFunc: func() string { return "common.txt" }}

			baseFile := &testfs.File{
				ReadDirFunc: func(n int) ([]ihfs.DirEntry, error) {
					return []ihfs.DirEntry{baseEntry1, baseEntry2}, nil
				},
			}
			layerFile := &testfs.File{
				ReadDirFunc: func(n int) ([]ihfs.DirEntry, error) {
					return []ihfs.DirEntry{layerEntry1, layerEntry2}, nil
				},
			}

			file := corfs.NewFile(baseFile, layerFile)
			entries, err := file.ReadDir(-1)

			Expect(err).ToNot(HaveOccurred())
			Expect(entries).To(HaveLen(3))
			names := make([]string, len(entries))
			for i, e := range entries {
				names[i] = e.Name()
			}
			Expect(names).To(ContainElements("layer1.txt", "common.txt", "base1.txt"))
		})

		It("should handle pagination", func() {
			baseEntry := &testfs.DirEntry{NameFunc: func() string { return "base.txt" }}
			layerEntry := &testfs.DirEntry{NameFunc: func() string { return "layer.txt" }}

			baseFile := &testfs.File{
				ReadDirFunc: func(n int) ([]ihfs.DirEntry, error) {
					return []ihfs.DirEntry{baseEntry}, nil
				},
			}
			layerFile := &testfs.File{
				ReadDirFunc: func(n int) ([]ihfs.DirEntry, error) {
					return []ihfs.DirEntry{layerEntry}, nil
				},
			}

			file := corfs.NewFile(baseFile, layerFile)
			
			// First call should return first entry
			entries1, err1 := file.ReadDir(1)
			Expect(err1).ToNot(HaveOccurred())
			Expect(entries1).To(HaveLen(1))
			Expect(entries1[0].Name()).To(Equal("layer.txt"))

			// Second call should return second entry
			entries2, err2 := file.ReadDir(1)
			Expect(err2).ToNot(HaveOccurred())
			Expect(entries2).To(HaveLen(1))
			Expect(entries2[0].Name()).To(Equal("base.txt"))

			// Third call should return EOF
			entries3, err3 := file.ReadDir(1)
			Expect(err3).To(Equal(io.EOF))
			Expect(entries3).To(BeNil())
		})

		It("should return EOF when no entries and n > 0", func() {
			baseFile := &testfs.File{
				ReadDirFunc: func(n int) ([]ihfs.DirEntry, error) {
					return []ihfs.DirEntry{}, nil
				},
			}
			layerFile := &testfs.File{
				ReadDirFunc: func(n int) ([]ihfs.DirEntry, error) {
					return []ihfs.DirEntry{}, nil
				},
			}

			file := corfs.NewFile(baseFile, layerFile)
			entries, err := file.ReadDir(1)

			Expect(err).To(Equal(io.EOF))
			Expect(entries).To(BeNil())
		})
	})
})
