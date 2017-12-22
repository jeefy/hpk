package hpk

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	apiv1batch "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	log "github.com/sirupsen/logrus"
)

// HPK Job object
// Intermediary object to map command line flags into the Job Spec
type hpkJob struct {
	Command, Image, ImagePullPolicy, JobName, Namespace string
	NumNodes                                            int
}

// GenerateClientSet(): With kubeconfig, generate and returns a generic Kubernetes client
func GenerateClientSet(kubeconfig string) *kubernetes.Clientset {
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Panic(fmt.Sprintf("%#v", err))
	}
	log.Debug("Kubernetes config generated.")

	// Create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Panic(fmt.Sprintf("%#v", err))
	}
	log.Debug("Kubernetes client generated.")

	return clientset
}

// ParseCmdArgs(): Dirty thing to handle arg parsing properly
func ParseCmdArgs() (hpkJob, *string) {
	var kubeconfig *string

	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	image := flag.String("image", "ubuntu", "Container image to run")
	imgpullpol := flag.String("imagepullpolicy", "Always", "Pull policy for the container. (Default: Always. Others: IfNotPresent,Never")
	loglevelstr := flag.String("loglevel", "warning", "Logs to display (Default: warning. Others: debug,info,error,fatal,panic)")
	kubenamespace := flag.String("namespace", "kube-public", "namespace to run jobs in")
	numnodes := flag.Int("numnodes", 1, "Number of nodes")

	flag.Parse()

	loglevelobj, err := log.ParseLevel(*loglevelstr)
	if err != nil {
		panic(err)
	}

	log.SetLevel(loglevelobj)
	log.Debug("Command line flags parsed")

	commandArgs := []string{}
	if len(flag.Args()) > 0 {
		commandArgs = append([]string{"/bin/bash", "-c"}, flag.Args()...)
	}

	jobcommand, err := json.Marshal(commandArgs)
	if err != nil {
		log.Panic(err)
	}
	log.Debug("Job command generated.")

	jobValues := hpkJob{
		Command:         string(jobcommand[:]),
		Image:           *image,
		ImagePullPolicy: *imgpullpol,
		JobName:         GenerateJobName(),
		NumNodes:        *numnodes,
		Namespace:       *kubenamespace,
	}

	log.Debug("hpkJob object generated.")

	return jobValues, kubeconfig
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
							Name:    jobValues.JobName,
							Image:   jobValues.Image,
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
