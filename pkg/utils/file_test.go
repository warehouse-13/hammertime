package utils_test

import (
	"encoding/json"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/weaveworks-liquidmetal/flintlock/api/types"

	"github.com/warehouse-13/hammertime/pkg/utils"
)

var _ = Describe("ProcessFile", func() {
	var (
		tempFile        *os.File
		uid             *string
		name, namespace string
	)

	JustBeforeEach(func() {
		var err error
		tempFile, err = ioutil.TempFile("", "utils_test")
		Expect(err).NotTo(HaveOccurred())

		spec := &types.MicroVMSpec{
			Id:        name,
			Namespace: namespace,
			Uid:       uid,
		}

		dat, err := json.Marshal(spec)
		Expect(err).NotTo(HaveOccurred())

		Expect(ioutil.WriteFile(tempFile.Name(), dat, 0755)).To(Succeed())
	})

	AfterEach(func() {
		Expect(os.RemoveAll(tempFile.Name())).To(Succeed())
	})

	Context("everything is set in the file", func() {
		BeforeEach(func() {
			name = "foo"
			namespace = "bar"
			uid = utils.PointyString("baz")
		})

		It("returns uid, name and namespace", func() {
			uid, name, namespace, err := utils.ProcessFile(tempFile.Name())
			Expect(err).NotTo(HaveOccurred())
			Expect(name).To(Equal("foo"))
			Expect(namespace).To(Equal("bar"))
			Expect(uid).To(Equal("baz"))
		})
	})

	Context("uid is not found in file", func() {
		BeforeEach(func() {
			name = "foo"
			namespace = "bar"
			uid = nil
		})

		It("returns name and namespace", func() {
			uid, name, namespace, err := utils.ProcessFile(tempFile.Name())
			Expect(err).NotTo(HaveOccurred())
			Expect(name).To(Equal("foo"))
			Expect(namespace).To(Equal("bar"))
			Expect(uid).To(Equal(""))
		})
	})

	Context("loading the file fails", func() {
		It("returns an error", func() {
			uid, name, namespace, err := utils.ProcessFile("noexist")
			Expect(err).To(HaveOccurred())
			Expect(name).To(Equal(""))
			Expect(namespace).To(Equal(""))
			Expect(uid).To(Equal(""))
		})
	})

	Context("no values are found in the file", func() {
		BeforeEach(func() {
			name = ""
			namespace = ""
			uid = nil
		})

		It("returns an error", func() {
			uid, name, namespace, err := utils.ProcessFile(tempFile.Name())
			Expect(err).To(HaveOccurred())
			Expect(name).To(Equal(""))
			Expect(namespace).To(Equal(""))
			Expect(uid).To(Equal(""))
		})
	})
})
