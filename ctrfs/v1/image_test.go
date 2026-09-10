package v1_test

import (
	"archive/tar"
	"io"
	"io/fs"
	"testing/fstest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	ctrfsv1 "github.com/unstoppablemango/ihfs/ctrfs/v1"
	"github.com/unstoppablemango/ihfs/memfs"
	"github.com/unstoppablemango/ihfs/testfs"
)

var _ = Describe("FS", Label("image"), func() {
	Describe("FromImage", func() {
		It("should open a file from an image", func() {
			layer, err := makeLayer([]tarEntry{
				{hdr: &tar.Header{Name: "hello.txt", Typeflag: tar.TypeReg, Size: 11, Mode: 0644}, data: "hello world"},
			})
			Expect(err).NotTo(HaveOccurred())

			img, err := mutate.AppendLayers(empty.Image, layer)
			Expect(err).NotTo(HaveOccurred())

			fsys := ctrfsv1.FromImage(img)
			defer fsys.Close()

			f, err := fsys.Open("hello.txt")
			Expect(err).NotTo(HaveOccurred())
			defer f.Close()

			data, err := io.ReadAll(f)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(data)).To(Equal("hello world"))
		})

		It("should stat a file from an image", func() {
			layer, err := makeLayer([]tarEntry{
				{hdr: &tar.Header{Name: "readme.md", Typeflag: tar.TypeReg, Size: 8, Mode: 0644}, data: "# readme"},
			})
			Expect(err).NotTo(HaveOccurred())

			img, err := mutate.AppendLayers(empty.Image, layer)
			Expect(err).NotTo(HaveOccurred())

			fsys := ctrfsv1.FromImage(img)
			defer fsys.Close()

			info, err := fs.Stat(fsys, "readme.md")
			Expect(err).NotTo(HaveOccurred())
			Expect(info.Name()).To(Equal("readme.md"))
			Expect(info.IsDir()).To(BeFalse())
		})

		It("should read directory entries from an image", func() {
			layer, err := makeLayer([]tarEntry{
				{hdr: &tar.Header{Name: "dir/file.txt", Typeflag: tar.TypeReg, Size: 4, Mode: 0644}, data: "data"},
			})
			Expect(err).NotTo(HaveOccurred())

			img, err := mutate.AppendLayers(empty.Image, layer)
			Expect(err).NotTo(HaveOccurred())

			fsys := ctrfsv1.FromImage(img)
			defer fsys.Close()

			entries, err := fs.ReadDir(fsys, "dir")
			Expect(err).NotTo(HaveOccurred())
			Expect(entries).To(HaveLen(1))
			Expect(entries[0].Name()).To(Equal("file.txt"))
		})

		Describe("fstest", func() {
			It("should pass fstest.TestFS", func() {
				layer, err := makeLayer([]tarEntry{
					{hdr: &tar.Header{Name: "dir/", Typeflag: tar.TypeDir, Mode: 0755}},
					{hdr: &tar.Header{Name: "dir/hello.txt", Typeflag: tar.TypeReg, Size: 5, Mode: 0644}, data: "hello"},
					{hdr: &tar.Header{Name: "readme.md", Typeflag: tar.TypeReg, Size: 6, Mode: 0644}, data: "readme"},
				})
				Expect(err).NotTo(HaveOccurred())

				img, err := mutate.AppendLayers(empty.Image, layer)
				Expect(err).NotTo(HaveOccurred())

				fsys := ctrfsv1.FromImage(img)
				defer fsys.Close()

				err = fstest.TestFS(fsys, "dir/hello.txt", "readme.md")
				Expect(err).NotTo(HaveOccurred())
			})
		})

		It("should merge multiple layers", func() {
			layer1, err := makeLayer([]tarEntry{
				{hdr: &tar.Header{Name: "base.txt", Typeflag: tar.TypeReg, Size: 4, Mode: 0644}, data: "base"},
			})
			Expect(err).NotTo(HaveOccurred())

			layer2, err := makeLayer([]tarEntry{
				{hdr: &tar.Header{Name: "overlay.txt", Typeflag: tar.TypeReg, Size: 7, Mode: 0644}, data: "overlay"},
			})
			Expect(err).NotTo(HaveOccurred())

			img, err := mutate.AppendLayers(empty.Image, layer1, layer2)
			Expect(err).NotTo(HaveOccurred())

			fsys := ctrfsv1.FromImage(img)
			defer fsys.Close()

			data1, err := fs.ReadFile(fsys, "base.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(string(data1)).To(Equal("base"))

			data2, err := fs.ReadFile(fsys, "overlay.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(string(data2)).To(Equal("overlay"))
		})
	})
})

var _ = Describe("AppendFS", func() {
	It("should append a layer from an fs.FS onto a base image", func() {
		m := memfs.New()
		f, err := m.Create("app.bin")
		Expect(err).NotTo(HaveOccurred())
		Expect(f.Close()).To(Succeed())

		img, err := ctrfsv1.AppendFS(empty.Image, m, ".")
		Expect(err).NotTo(HaveOccurred())

		layers, err := img.Layers()
		Expect(err).NotTo(HaveOccurred())
		Expect(layers).To(HaveLen(1))
	})

	It("should propagate ToLayer errors", func() {
		_, err := ctrfsv1.AppendFS(empty.Image, testfs.BoringFs{}, "nonexistent")

		Expect(err).To(HaveOccurred())
	})
})
