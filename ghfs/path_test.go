package ghfs

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unstoppablemango/ihfs"
)

var _ = Describe("normalize", func() {
	DescribeTable("api.github.com scheme (pass-through)",
		func(input, expected string) {
			result, err := normalize(input)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(expected))
		},
		Entry(nil, "https://api.github.com/users/test-user", "/users/test-user"),
		Entry(nil, "https://api.github.com/repos/owner/repo", "/repos/owner/repo"),
		Entry(nil, "https://api.github.com/repos/owner/repo/contents/file.txt?ref=main", "/repos/owner/repo/contents/file.txt?ref=main"),
	)

	DescribeTable("github.com scheme (web-style)",
		func(input, expected string) {
			result, err := normalize(input)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(expected))
		},
		Entry(nil, "https://github.com/test-user", "users/test-user"),
		Entry(nil, "https://github.com/owner/repo", "repos/owner/repo"),
		Entry(nil, "https://github.com/owner/repo/tree/main", "repos/owner/repo/branches/main"),
		Entry(nil, "https://github.com/owner/repo/blob/main/file.txt", "repos/owner/repo/contents/file.txt?ref=main"),
	)

	DescribeTable("raw.githubusercontent.com scheme (raw-style)",
		func(input, expected string) {
			result, err := normalize(input)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(expected))
		},
		Entry(nil, "https://raw.githubusercontent.com/test-user", "users/test-user"),
		Entry(nil, "https://raw.githubusercontent.com/owner/repo", "repos/owner/repo"),
		Entry(nil, "https://raw.githubusercontent.com/owner/repo/main", "repos/owner/repo/branches/main"),
		Entry(nil, "https://raw.githubusercontent.com/owner/repo/main/file.txt", "repos/owner/repo/contents/file.txt?ref=main"),
		Entry(nil, "https://raw.githubusercontent.com/owner/repo/main/nested/file.txt", "repos/owner/repo/contents/nested/file.txt?ref=main"),
	)

	DescribeTable("schemeless github.com prefix (web-style)",
		func(input, expected string) {
			result, err := normalize(input)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(expected))
		},
		Entry(nil, "github.com/test-user", "users/test-user"),
		Entry(nil, "github.com/owner/repo", "repos/owner/repo"),
		Entry(nil, "github.com/owner/repo/tree/main", "repos/owner/repo/branches/main"),
	)

	DescribeTable("schemeless api.github.com prefix (pass-through, host stripped)",
		func(input, expected string) {
			result, err := normalize(input)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(expected))
		},
		Entry(nil, "api.github.com/users/test-user", "users/test-user"),
		Entry(nil, "api.github.com/repos/owner/repo", "repos/owner/repo"),
	)

	DescribeTable("schemeless raw.githubusercontent.com prefix (raw-style)",
		func(input, expected string) {
			result, err := normalize(input)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(expected))
		},
		Entry(nil, "raw.githubusercontent.com/test-user", "users/test-user"),
		Entry(nil, "raw.githubusercontent.com/owner/repo/main/file.txt", "repos/owner/repo/contents/file.txt?ref=main"),
	)

	DescribeTable("no prefix (API pass-through)",
		func(input, expected string) {
			result, err := normalize(input)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(expected))
		},
		Entry(nil, "users/test-user", "users/test-user"),
		Entry(nil, "repos/owner/repo", "repos/owner/repo"),
		Entry(nil, "repos/owner/repo/contents/file.txt?ref=main", "repos/owner/repo/contents/file.txt?ref=main"),
	)

	It("should pass through unknown hostnames unchanged", func() {
		result, err := normalize("https://gitlab.com/owner/repo")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("https://gitlab.com/owner/repo"))
	})
})

var _ = Describe("fromWebURL", func() {
	It("should return user path for 1 segment", func() {
		result, err := fromWebURL("test-user")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("users/test-user"))
	})

	It("should return current user for empty path", func() {
		result, err := fromWebURL("")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("user"))
	})

	It("should return repo path for 2 segments", func() {
		result, err := fromWebURL("owner/repo")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("repos/owner/repo"))
	})

	It("should return branch path for owner/repo/tree/branch", func() {
		result, err := fromWebURL("owner/repo/tree/main")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("repos/owner/repo/branches/main"))
	})

	It("should return content path for owner/repo/blob/branch/file", func() {
		result, err := fromWebURL("owner/repo/blob/main/file.txt")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("repos/owner/repo/contents/file.txt?ref=main"))
	})

	It("should return release path for owner/repo/releases/tag/TAG", func() {
		result, err := fromWebURL("owner/repo/releases/tag/v1.0.0")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("repos/owner/repo/releases/tags/v1.0.0"))
	})

	It("should return release path for owner/repo/releases/download/TAG", func() {
		result, err := fromWebURL("owner/repo/releases/download/v1.0.0")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("repos/owner/repo/releases/tags/v1.0.0"))
	})

	It("should return asset path for owner/repo/releases/tag/TAG/asset", func() {
		result, err := fromWebURL("owner/repo/releases/tag/v1.0.0/asset.tar.gz")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("repos/owner/repo/releases/assets/asset.tar.gz"))
	})

	It("should return nested content path for owner/repo/blob/branch/a/b", func() {
		result, err := fromWebURL("owner/repo/blob/main/nested/file.txt")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("repos/owner/repo/contents/nested/file.txt?ref=main"))
	})

	It("should return ErrNotExist for 3 unrecognized segments", func() {
		_, err := fromWebURL("owner/repo/invalid")
		Expect(err).To(HaveOccurred())
		Expect(err).To(MatchError(ihfs.ErrNotExist))
	})

	It("should return ErrNotExist for 4 segments without tree", func() {
		_, err := fromWebURL("owner/repo/blob/branch")
		Expect(err).To(HaveOccurred())
		Expect(err).To(MatchError(ihfs.ErrNotExist))
	})
})

var _ = Describe("fromRawURL", func() {
	It("should return user path for 1 segment", func() {
		result, err := fromRawURL("test-user")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("users/test-user"))
	})

	It("should return current user for empty path", func() {
		result, err := fromRawURL("")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("user"))
	})

	It("should return repo path for 2 segments", func() {
		result, err := fromRawURL("owner/repo")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("repos/owner/repo"))
	})

	It("should return branch path for 3 segments", func() {
		result, err := fromRawURL("owner/repo/main")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("repos/owner/repo/branches/main"))
	})

	It("should return content path for 4 segments", func() {
		result, err := fromRawURL("owner/repo/main/file.txt")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("repos/owner/repo/contents/file.txt?ref=main"))
	})

	It("should return nested content path for 5+ segments", func() {
		result, err := fromRawURL("owner/repo/main/nested/file.txt")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("repos/owner/repo/contents/nested/file.txt?ref=main"))
	})

	It("should strip raw.githubusercontent.com prefix", func() {
		result, err := fromRawURL("raw.githubusercontent.com/owner/repo/main/file.txt")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("repos/owner/repo/contents/file.txt?ref=main"))
	})
})
