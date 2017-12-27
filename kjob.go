package hpk

import (
	"bytes"
	"flag"
	"strconv"
	"time"

	apiv1batch "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HPK Job object
// Intermediary object to map command line flags into the Job Spec
type hpkJob struct {
	Command, Image, ImagePullPolicy, JobName, Namespace, MaxCPU, MaxMemory string
	NumNodes                                                               int
}

type hpkJobLog struct {
	JobName string `bson:"jobName"`
	PodName string `bson:"podName"`
	Log     string `bson:"log"`
}

// GenerateJobName(): Generate/concat JobName
func GenerateJobName() string {
	var bufferName bytes.Buffer
	bufferName.WriteString("hkp-")
	bufferName.WriteString(strconv.FormatInt(time.Now().UnixNano(), 10))
	return bufferName.String()
}

// GenerateJobNameFilter(): Generate/concat the LabelSelector for use later
func GenerateJobNameFilter(jobName string) string {
	var bufferLabel bytes.Buffer
	bufferLabel.WriteString("job-name=")
	bufferLabel.WriteString(jobName)
	return bufferLabel.String()
}

// GenerateJobTemplate(): Applys hpkJob object to an *apiv1batch.Job
func GenerateJobTemplate(jobValues hpkJob) *apiv1batch.Job {
	numNodes32 := int32(jobValues.NumNodes)
	maxCPU := resource.MustParse(jobValues.MaxCPU)
	maxMemory := resource.MustParse(jobValues.MaxMemory)
	kJob := apiv1batch.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: jobValues.JobName,
		},
		Spec: apiv1batch.JobSpec{
			Parallelism: &numNodes32,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      jobValues.JobName,
					Namespace: jobValues.Namespace,
				},
				Spec: v1.PodSpec{
					RestartPolicy: "Never",
					Containers: []v1.Container{
						{
							Name:  jobValues.JobName,
							Image: jobValues.Image,
							Resources: v1.ResourceRequirements{
								Limits: v1.ResourceList{
									v1.ResourceCPU:    maxCPU,
									v1.ResourceMemory: maxMemory,
								},
							},
							Command: flag.Args(),
							//Command:    []string{"ls", "-la", "/root/"},
							WorkingDir: "/root/",
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      "scratch",
									MountPath: "/scratch/",
								},
							},
						},
					},
					Volumes: []v1.Volume{
						{
							Name: "scratch",
						},
					},
				},
			},
		},
	}
	return &kJob
}
