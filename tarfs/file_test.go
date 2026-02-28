package tarfs_test

import (
	"archive/tar"
	"bytes"
	"io"
	"io/fs"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs/tarfs"
)

var _ = Describe("File", func() {
	var (
		tfs  *tarfs.TarFile
		file fs.File
	)

	BeforeEach(func() {
		var err error
		tfs, err = tarfs.Open("../testdata/test.tar")
		Expect(err).NotTo(HaveOccurred())

		file, err = tfs.Open("tartest/test.txt")
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Read", func() {
		It("should read file contents", func() {
			buf := make([]byte, 12)
			n, err := file.Read(buf)

			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(12))
			Expect(string(buf)).To(Equal("test content"))
		})

		It("should return EOF when reading past end", func() {
			_, err := io.ReadAll(file)
			Expect(err).NotTo(HaveOccurred())

			buf := make([]byte, 10)
			n, err := file.Read(buf)

			Expect(err).To(Equal(io.EOF))
			Expect(n).To(Equal(0))
		})

		It("should support multiple reads", func() {
			buf1 := make([]byte, 5)
			n1, err := file.Read(buf1)
			Expect(err).NotTo(HaveOccurred())
			Expect(n1).To(Equal(5))
			Expect(string(buf1)).To(Equal("test "))

			buf2 := make([]byte, 7)
			n2, err := file.Read(buf2)
			Expect(err).NotTo(HaveOccurred())
			Expect(n2).To(Equal(7))
			Expect(string(buf2)).To(Equal("content"))
		})
	})

	Describe("Stat", func() {
		It("should return file info", func() {
			info, err := file.Stat()

			Expect(err).NotTo(HaveOccurred())
			Expect(info).NotTo(BeNil())
			Expect(info.Name()).To(Equal("test.txt"))
			Expect(info.Size()).To(Equal(int64(13)))
			Expect(info.IsDir()).To(BeFalse())
		})
	})

	Describe("Close", func() {
		It("should close without error", func() {
			err := file.Close()

			Expect(err).NotTo(HaveOccurred())
		})

		It("should be idempotent", func() {
			err := file.Close()
			Expect(err).NotTo(HaveOccurred())

			err = file.Close()
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("Name", func() {
		It("should return the file path", func() {
			tf, ok := file.(*tarfs.File)
			Expect(ok).To(BeTrueBecause("file is a *tarfs.File"))
			Expect(tf.Name()).To(Equal("tartest/test.txt"))
		})
	})

	Describe("ReadDir", func() {
		It("should use real cached directory entry when available", func() {
			var buf bytes.Buffer
			tw := tar.NewWriter(&buf)
			// Explicit child dir header (no explicit parent header, so parent is synthetic)
			Expect(tw.WriteHeader(&tar.Header{Name: "parent/child/", Typeflag: tar.TypeDir, Mode: 0755})).To(Succeed())
			Expect(tw.WriteHeader(&tar.Header{Name: "parent/child/file.txt", Mode: 0644, Size: 4})).To(Succeed())
			_, _ = tw.Write([]byte("data"))
			Expect(tw.Close()).To(Succeed())

			tfs := tarfs.FromReader("test.tar", bytes.NewReader(buf.Bytes()))

			// Open "parent" — synthetic dir, but "parent/child" is a real cached entry
			parentFile, err := tfs.Open("parent")
			Expect(err).NotTo(HaveOccurred())
			defer parentFile.Close()

			rdFile, ok := parentFile.(fs.ReadDirFile)
			Expect(ok).To(BeTrue())

			entries, err := rdFile.ReadDir(-1)
			Expect(err).NotTo(HaveOccurred())
			Expect(entries).To(HaveLen(1))
			Expect(entries[0].Name()).To(Equal("child"))
			Expect(entries[0].IsDir()).To(BeTrue())
		})

		// Bug: fileData.file sets File.name = fd.hdr.Name verbatim. When a tar archive
		// uses GNU-tar-style trailing-slash directory headers (e.g. "mydir/"), the returned
		// File has name "mydir/", so ReadDir computes prefix "mydir//" which never matches
		// any child entries, making the directory appear empty.
		It("should list children of a GNU-tar-style directory entry (trailing slash)", func() {
			var buf bytes.Buffer
			tw := tar.NewWriter(&buf)
			// GNU tar / gtar convention: directory headers end with "/"
			Expect(tw.WriteHeader(&tar.Header{Name: "mydir/", Typeflag: tar.TypeDir, Mode: 0755})).To(Succeed())
			Expect(tw.WriteHeader(&tar.Header{Name: "mydir/file.txt", Mode: 0644, Size: 5})).To(Succeed())
			_, _ = tw.Write([]byte("hello"))
			Expect(tw.Close()).To(Succeed())

			tfs := tarfs.FromReader("test.tar", bytes.NewReader(buf.Bytes()))

			// Open root first to fully hydrate the cache (including "mydir/file.txt")
			root, err := tfs.Open(".")
			Expect(err).NotTo(HaveOccurred())
			root.Close()

			// Open "mydir" — hits cache, gets the fileData whose hdr.Name is "mydir/"
			// fileData.file() sets File.name = "mydir/", so ReadDir builds prefix "mydir//"
			dir, err := tfs.Open("mydir")
			Expect(err).NotTo(HaveOccurred())
			defer dir.Close()

			rdFile, ok := dir.(fs.ReadDirFile)
			Expect(ok).To(BeTrue())

			// Bug: returns [] because "mydir/file.txt" does not start with "mydir//"
			entries, err := rdFile.ReadDir(-1)
			Expect(err).NotTo(HaveOccurred())
			Expect(entries).To(HaveLen(1))
			Expect(entries[0].Name()).To(Equal("file.txt"))
		})

		// Bug: ReadDir rebuilds and re-sorts the full entry list from the shared cache on
		// every call. If the cache grows between paginated calls (because another Open adds
		// an entry that sorts before the current readDirCount position), earlier entries are
		// shifted forward and the offset points to the wrong position, producing duplicates
		// and skipping entries.
		It("should not return duplicate entries when cache grows between paginated reads", func() {
			var buf bytes.Buffer
			tw := tar.NewWriter(&buf)
			// Explicit dir header without trailing slash; children in reverse alpha order
			// so that the second child ("a.txt") will sort before the first ("b.txt") once added.
			Expect(tw.WriteHeader(&tar.Header{Name: "dir", Typeflag: tar.TypeDir, Mode: 0755})).To(Succeed())
			Expect(tw.WriteHeader(&tar.Header{Name: "dir/b.txt", Mode: 0644, Size: 1})).To(Succeed())
			_, _ = tw.Write([]byte("b"))
			Expect(tw.WriteHeader(&tar.Header{Name: "dir/a.txt", Mode: 0644, Size: 1})).To(Succeed())
			_, _ = tw.Write([]byte("a"))
			Expect(tw.Close()).To(Succeed())

			tfs := tarfs.FromReader("test.tar", bytes.NewReader(buf.Bytes()))

			// Open "dir/b.txt" — lazily reads the "dir" header and "dir/b.txt" into cache;
			// the tar reader is now positioned after "dir/b.txt", with "dir/a.txt" still unread.
			_, err := tfs.Open("dir/b.txt")
			Expect(err).NotTo(HaveOccurred())

			// Open "dir" from cache — tar reader position is unchanged.
			dirFile, err := tfs.Open("dir")
			Expect(err).NotTo(HaveOccurred())
			defer dirFile.Close()

			rdFile, ok := dirFile.(fs.ReadDirFile)
			Expect(ok).To(BeTrue())

			// First paginated read: only "dir/b.txt" is cached → sorted entries = ["b.txt"]
			// → readDirCount advances to 1.
			entries1, err := rdFile.ReadDir(1)
			Expect(err).NotTo(HaveOccurred())
			Expect(entries1).To(HaveLen(1))

			// Open "dir/a.txt" — reads it from the tar and adds it to the shared cache.
			// "a.txt" sorts BEFORE "b.txt", so the rebuilt sorted list becomes ["a.txt", "b.txt"].
			_, err = tfs.Open("dir/a.txt")
			Expect(err).NotTo(HaveOccurred())

			// Second paginated read: cache now has both entries.
			// Bug: ReadDir rebuilds ["a.txt", "b.txt"], readDirCount=1 still points at index 1
			// → returns "b.txt" again instead of "a.txt".
			entries2, err := rdFile.ReadDir(1)
			Expect(err).NotTo(HaveOccurred())
			Expect(entries2).To(HaveLen(1))

			// Together the two reads should return each entry exactly once.
			allNames := []string{entries1[0].Name(), entries2[0].Name()}
			Expect(allNames).To(ConsistOf("a.txt", "b.txt"))
		})
	})
})
