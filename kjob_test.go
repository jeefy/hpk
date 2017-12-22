package hpk

import "testing"
import "github.com/stretchr/testify/assert"

func testGenerateJobName(t *testing.T) {
	name1 := GenerateJobName()
	name2 := GenerateJobName()
	assert.NotEqual(t, name1, name2)
}
func testParseCmdArgs(t *testing.T) {
	kJob, _ := ParseCmdArgs()
	assert.Equal(t, kJob.Image, "ubuntu")
	assert.Equal(t, kJob.ImagePullPolicy, "never")
	assert.Equal(t, kJob.Namespace, "kube-public")
	assert.Equal(t, kJob.NumNodes, 1)

}
