package microvm_test

import (
	"encoding/base64"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/warehouse-13/hammertime/pkg/microvm"
	"github.com/weaveworks-liquidmetal/flintlock/client/cloudinit/userdata"
	"gopkg.in/yaml.v2"
)

func Test_CreateUserData(t *testing.T) {
	g := NewWithT(t)

	var (
		testName = "foo"
		testPath = ""
	)

	out, err := microvm.CreateUserData(testName, testPath)
	g.Expect(err).NotTo(HaveOccurred())

	dat, err := base64.StdEncoding.DecodeString(out)
	g.Expect(err).NotTo(HaveOccurred())
	generated := &userdata.UserData{}
	g.Expect(yaml.Unmarshal(dat, generated)).To(Succeed())

	g.Expect(generated.HostName).To(Equal(testName))
	g.Expect(generated.Users).To(HaveLen(1))
	g.Expect(generated.Users[0].Name).To(Equal("root"))
	g.Expect(generated.Users[0].SSHAuthorizedKeys).To(HaveLen(0))
}

func Test_CreateUserData_withSSHKey(t *testing.T) {
	g := NewWithT(t)

	var (
		testName = "foo"
		testPath = "userdata_test"
		testKey  = "this is a key"
	)

	tempFile, err := ioutil.TempFile("", testPath)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(ioutil.WriteFile(tempFile.Name(), []byte(testKey), 0755)).To(Succeed())

	t.Cleanup(func() {
		g.Expect(os.RemoveAll(tempFile.Name())).To(Succeed())
	})

	out, err := microvm.CreateUserData(testName, tempFile.Name())
	g.Expect(err).NotTo(HaveOccurred())

	dat, err := base64.StdEncoding.DecodeString(out)
	g.Expect(err).NotTo(HaveOccurred())
	generated := &userdata.UserData{}
	g.Expect(yaml.Unmarshal(dat, generated)).To(Succeed())

	g.Expect(generated.HostName).To(Equal(testName))
	g.Expect(generated.Users).To(HaveLen(1))
	g.Expect(generated.Users[0].Name).To(Equal("root"))
	g.Expect(generated.Users[0].SSHAuthorizedKeys).To(HaveLen(1))
	g.Expect(generated.Users[0].SSHAuthorizedKeys[0]).To(Equal(testKey))
}

func Test_CreateUserData_withSSHKey_readFileFails(t *testing.T) {
	g := NewWithT(t)

	var (
		testName = "foo"
		testPath = "noexist"
	)

	_, err := microvm.CreateUserData(testName, testPath)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err).To(BeAssignableToTypeOf(&fs.PathError{}))
}

func Test_CreateMetadata(t *testing.T) {
	g := NewWithT(t)

	var (
		testName      = "foo"
		testNamespace = "bar"
	)

	out, err := microvm.CreateMetadata(testName, testNamespace)
	g.Expect(err).NotTo(HaveOccurred())

	dat, err := base64.StdEncoding.DecodeString(out)
	g.Expect(err).NotTo(HaveOccurred())
	generated := map[string]string{}
	g.Expect(yaml.Unmarshal(dat, generated)).To(Succeed())

	g.Expect(generated).To(HaveKeyWithValue("instance_id", fmt.Sprintf("%s/%s", testNamespace, testName)))
	g.Expect(generated).To(HaveKeyWithValue("local_hostname", testName))
	g.Expect(generated).To(HaveKeyWithValue("platform", "liquid_metal"))
}
