package hpk

import (
	"encoding/json"
	"flag"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	apiv1batch "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	v1beta1 "k8s.io/client-go/kubernetes/typed/extensions/v1beta1"
	"k8s.io/client-go/tools/clientcmd"
)

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

func ParseKubeconfig() *string {
	var kubeconfig *string

	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	return kubeconfig
}

// ParseCmdArgs(): Dirty thing to handle arg parsing properly
/*
hpk-num-nodes
hpk-base-image
hpk-pull-policy
hpk-default-namespace
hpk-max-cpu
hpk-max-memory
*/
func ParseCmdArgs() (hpkJob, *string) {
	kubeconfig := ParseKubeconfig()
	numnodeconv, _ := strconv.ParseInt(GetConfigKey(kubeconfig, "hpk-num-nodes"), 10, 16)
	image := flag.String("image", GetConfigKey(kubeconfig, "hpk-base-image"), "Container image to run")
	imgpullpol := flag.String("imagepullpolicy", GetConfigKey(kubeconfig, "hpk-pull-policy"), "Pull policy for the container. (Default: Always. Others: IfNotPresent,Never")
	loglevelstr := flag.String("loglevel", "warning", "Logs to display (Default: warning. Others: debug,info,error,fatal,panic)")
	kubenamespace := flag.String("namespace", GetConfigKey(kubeconfig, "hpk-default-namespace"), "namespace to run jobs in")
	numnodes := flag.Int("numnodes", int(numnodeconv), "Number of nodes")
	maxcpu := flag.String("maxcpu", GetConfigKey(kubeconfig, "hpk-max-cpu"), "Max CPU per node")
	maxmemory := flag.String("maxmemory", GetConfigKey(kubeconfig, "hpk-max-memory"), "Max memory per node")

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
		MaxCPU:          *maxcpu,
		MaxMemory:       *maxmemory,
		Namespace:       *kubenamespace,
	}

	log.Debug("hpkJob object generated.")

	return jobValues, kubeconfig
}

func GetDeploymentClient(kubeconfig *string) v1beta1.DeploymentInterface {
	clientset := GenerateClientSet(*kubeconfig)

	deploymentClient := clientset.Extensions().Deployments("kube-system")

	return deploymentClient
}

func GetCPUCost(kubeconfig *string, job *apiv1batch.Job) string {
	log.Info("Checking allocation for " + job.GetName())
	return GetConfigKey(kubeconfig, "hpk-cost-cpu")
}

func GetMemoryCost(kubeconfig *string, job *apiv1batch.Job) string {
	log.Info("Checking allocation for " + job.GetName())
	return GetConfigKey(kubeconfig, "hpk-cost-memory")
}

func GetConfigKey(kubeconfig *string, key string) string {
	config := GetFullConfig(kubeconfig)
	if val, ok := config[key]; ok {
		return val
	}
	return ""
}

func GetFullConfig(kubeconfig *string) map[string]string {
	deploymentClient := GetDeploymentClient(kubeconfig)
	deployment, err := deploymentClient.Get("kube-dns", v1.GetOptions{})
	if err != nil {
		log.Info(err)
	}

	data := deployment.GetAnnotations()

	annotations := make(map[string]string)

	for k, v := range data {
		if strings.Contains(k, "hpk-") {
			annotations[k] = v
		}
	}

	return annotations
}

func UpdateConfig(kubeconfig *string, annotations map[string]string) map[string]string {
	deploymentClient := GetDeploymentClient(kubeconfig)
	deployment, err := deploymentClient.Get("kube-dns", v1.GetOptions{})

	deployment.SetAnnotations(annotations)

	deployment, err = deploymentClient.Update(deployment)
	if err != nil {
		log.Info(err)
	}
	return deployment.GetAnnotations()
}

func UpdateConfigKey(kubeconfig *string, key string, val string) map[string]string {
	annotations := GetFullConfig(kubeconfig)
	annotations[key] = val
	return UpdateConfig(kubeconfig, annotations)
}

func RemoveConfigKey(kubeconfig *string, key string) map[string]string {
	annotations := GetFullConfig(kubeconfig)
	delete(annotations, key)
	return UpdateConfig(kubeconfig, annotations)
}
