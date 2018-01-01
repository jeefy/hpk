package hpk

import (
	"bytes"
	"flag"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	apiv1batch "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HPK Job object
// Intermediary object to map command line flags into the Job Spec
type hpkJob struct {
	Command, Image, ImagePullPolicy, JobName, Namespace, MaxCPU, MaxMemory, MaxRuntime string
	NumNodes                                                                           int
}

type hpkJobLog struct {
	JobName string `bson:"jobName"`
	PodName string `bson:"podName"`
	Log     string `bson:"log"`
}

type hpkAllocation struct {
	Name    string `json:"name"`
	Balance string `json:"balance"`
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

func GenerateJobCost(kubeconfig *string, job *apiv1batch.Job) float64 {
	duration := int64(1)
	var err error

	if !job.Status.CompletionTime.IsZero() {
		duration := job.Status.CompletionTime.Time.Sub(job.Status.StartTime.Time).Seconds()
		if duration < 1 {
			duration = 1
		}
	} else {
		duration, err = strconv.ParseInt(job.Annotations["max-runtime"], 10, 64)
		if err != nil {
			log.Info("Invalid default duration set.")
			duration = int64(1)
		}

	}

	cpu_val := float64(job.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu().MilliValue())
	memory_val := float64(job.Spec.Template.Spec.Containers[0].Resources.Limits.Memory().Value())
	//cpu_cost, ok := useData["cpu_cost"].(float64)
	cpu_cost, err := strconv.ParseFloat(GetCPUCost(kubeconfig, job), 64)
	if err != nil {
		log.Info("Invalid CPU Cost. Defaulting to Zero")
		cpu_cost = 0
	}
	//memory_cost, ok := useData["memory_cost"].(float64)
	memory_cost, err := strconv.ParseFloat(GetMemoryCost(kubeconfig, job), 64)
	if err != nil {
		log.Info("Invalid Memory Cost. Defaulting to Zero")
		memory_cost = 0
	}

	total_cost := float64(*job.Spec.Parallelism) * (float64(duration)*((cpu_val/1000)*cpu_cost) + ((memory_val / 1000) * memory_cost))
	if total_cost < 0.01 {
		log.Info("Floor met. Setting total cost to 0.01")
		total_cost = 0.01
	}

	return total_cost
}

// GenerateJobTemplate(): Applys hpkJob object to an *apiv1batch.Job
func GenerateJobTemplate(jobValues hpkJob) *apiv1batch.Job {
	numNodes32 := int32(jobValues.NumNodes)
	maxCPU := resource.MustParse(jobValues.MaxCPU)
	maxMemory := resource.MustParse(jobValues.MaxMemory)
	maxRuntime, _ := strconv.ParseInt(jobValues.MaxRuntime, 10, 64)
	kJob := apiv1batch.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: jobValues.JobName,
			Annotations: map[string]string{
				"max-runtime": jobValues.MaxRuntime,
			},
		},
		Spec: apiv1batch.JobSpec{
			ActiveDeadlineSeconds: &maxRuntime,
			Parallelism:           &numNodes32,
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
