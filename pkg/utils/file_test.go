package utils_test

import (
	"encoding/json"
	"io/fs"
	"io/ioutil"
	"os"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/weaveworks-liquidmetal/flintlock/api/types"

	"github.com/warehouse-13/hammertime/pkg/utils"
)

func Test_ProcessFile(t *testing.T) {
	g := NewWithT(t)

	tempFile, err := ioutil.TempFile("", "utils_test")
	g.Expect(err).NotTo(HaveOccurred())

	t.Cleanup(func() {
		g.Expect(os.RemoveAll(tempFile.Name())).To(Succeed())
	})

	var (
		testName = "foo"
		testNs   = "bar"
		testUid  = "baz"
	)

	type testData struct {
		name string
		ns   string
		uid  *string
	}

	tt := []struct {
		test     string
		filename string
		input    testData
		expected func(*WithT, testData, testData, error)
	}{
		{
			test:     "when all values are set in the file, returns all",
			filename: tempFile.Name(),
			input: testData{
				name: testName,
				ns:   testNs,
				uid:  &testUid,
			},
			expected: func(g *WithT, in, out testData, err error) {
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(out.name).To(Equal(in.name))
				g.Expect(out.ns).To(Equal(in.ns))
				g.Expect(out.uid).To(Equal(in.uid))
			},
		},
		{
			test:     "when just name and ns are set in the file, returns name and ns",
			filename: tempFile.Name(),
			input: testData{
				name: testName,
				ns:   testNs,
				uid:  utils.PointyString(""),
			},
			expected: func(g *WithT, in, out testData, err error) {
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(out.name).To(Equal(in.name))
				g.Expect(out.ns).To(Equal(in.ns))
				g.Expect(out.uid).To(Equal(in.uid))
			},
		},
		{
			test:     "when no values are set in the file, returns an error",
			filename: tempFile.Name(),
			input:    testData{},
			expected: func(g *WithT, in, out testData, err error) {
				g.Expect(err).To(MatchError(ContainSubstring("required")))
			},
		},
		{
			test:     "when the file fails to load, returns an error",
			filename: "noexist",
			input:    testData{},
			expected: func(g *WithT, in, out testData, err error) {
				g.Expect(err).To(BeAssignableToTypeOf(&fs.PathError{}))
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.test, func(t *testing.T) {
			spec := &types.MicroVMSpec{
				Id:        tc.input.name,
				Namespace: tc.input.ns,
				Uid:       tc.input.uid,
			}

			dat, err := json.Marshal(spec)
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(ioutil.WriteFile(tempFile.Name(), dat, 0755)).To(Succeed())

			outUid, outN, outNs, err := utils.ProcessFile(tc.filename)

			out := testData{outN, outNs, &outUid}

			tc.expected(g, tc.input, out, err)
		})
	}
}

func Test_LoadSpecFromFile(t *testing.T) {
	g := NewWithT(t)

	tempFile, err := ioutil.TempFile("", "utils_test")
	g.Expect(err).NotTo(HaveOccurred())

	t.Cleanup(func() {
		g.Expect(os.RemoveAll(tempFile.Name())).To(Succeed())
	})

	type testData struct {
		name string
		ns   string
		uid  *string
	}

	tt := []struct {
		test     string
		filename string
		input    string
		expected func(*WithT, *types.MicroVMSpec, error)
	}{
		{
			test:     "happy path",
			filename: tempFile.Name(),
			input:    `{"name": "bar"}`,
			expected: func(g *WithT, out *types.MicroVMSpec, err error) {
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(out).To(BeAssignableToTypeOf(&types.MicroVMSpec{}))
			},
		},
		{
			test:     "if the file contains non-json data, returns an error",
			filename: tempFile.Name(),
			input:    `foo`,
			expected: func(g *WithT, out *types.MicroVMSpec, err error) {
				g.Expect(err).To(HaveOccurred())
				g.Expect(out).To(BeNil())
			},
		},
		{
			test:     "if the file cannot be loaded, returns an error",
			filename: "noexist",
			expected: func(g *WithT, out *types.MicroVMSpec, err error) {
				g.Expect(err).To(HaveOccurred())
				g.Expect(out).To(BeNil())
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.test, func(t *testing.T) {
			g.Expect(ioutil.WriteFile(tempFile.Name(), []byte(tc.input), 0755)).To(Succeed())

			out, err := utils.LoadSpecFromFile(tc.filename)
			tc.expected(g, out, err)
		})
	}
}
