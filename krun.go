// Submit a batch style job to Kubernetes in an HPC-style way.
// Similar to `srun` from Slurm.
package hpk

import (
	"fmt"
	"os"
	"time"

	apiv1batch "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"

	"github.com/kylelemons/godebug/pretty"
	log "github.com/sirupsen/logrus"
)

// krun(): Primary function for the krun utility.
// Lets users submit an interactive HPC-style command
// to a Kubernetes cluster
func KRun() {
	defer func() { //catch or finally
		if err := recover(); err != nil { //catch
			os.Exit(1)
		}
	}()
	log.SetOutput(os.Stdout)

	jobValues, kubeconfig := ParseCmdArgs()

	clientset := GenerateClientSet(*kubeconfig)

	// Create the job client
	jobsClient := clientset.Batch().Jobs(jobValues.Namespace)
	log.Debug("Kuberntes job client generated.")

	newJob := GenerateJobTemplate(jobValues)

	if !JobResourceCheck(kubeconfig, newJob) {
		log.Panic("Resources unavailable for your allocation.")
	}

	_, jobErr := jobsClient.Create(newJob)
	if jobErr != nil {
		log.Panic(jobErr)
	}
	log.Info(fmt.Sprintf("Submitted job %q.\n", jobValues.JobName))

	// Create the Informer to watch Kubernetes for Job updates
	jobFilter := GenerateJobNameFilter(jobValues.JobName)

	lw := &cache.ListWatch{
		ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
			return jobsClient.List(metav1.ListOptions{LabelSelector: jobFilter})
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return jobsClient.Watch(metav1.ListOptions{LabelSelector: jobFilter})
		},
	}
	_, ctrl := cache.NewInformer(lw,
		&apiv1batch.Job{},
		time.Duration(1000*time.Millisecond),
		cache.ResourceEventHandlerFuncs{
			DeleteFunc: func(obj interface{}) {
				if resource, ok := obj.(*apiv1batch.Job); ok {
					log.Info(fmt.Sprintf("Job %s@%s deleted\n", resource.Name, resource.Namespace))
				}
			},
			AddFunc: func(obj interface{}) {
				if resource, ok := obj.(*apiv1batch.Job); ok {
					log.Info(fmt.Sprintf("Job %s@%s created\n", resource.Name, resource.Namespace))

				}
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				if oldObj != newObj {
					if oldJob, ok := oldObj.(*apiv1batch.Job); ok {
						if newJob, ok := newObj.(*apiv1batch.Job); ok {
							log.Info(fmt.Sprintf("Job %s@%s updated\n", newJob.Name, newJob.Namespace))
							if diff := pretty.Compare(oldJob, newJob); diff != "" {
								log.Debug(fmt.Sprintf("%s: Job diff: (-old +new)\n%s", newJob.Name, diff))
								if oldJob.Status.CompletionTime == nil && newJob.Status.CompletionTime != nil {
									log.Info(fmt.Sprintf("Job %s@%s completed\n", newJob.Name, newJob.Namespace))
									podsClient := clientset.CoreV1().Pods(newJob.Namespace)
									podList, err := podsClient.List(metav1.ListOptions{LabelSelector: jobFilter})
									if err != nil {
										log.Panic(err)
									}
									for _, pod := range podList.Items {
										log.Debug(fmt.Sprintf("Pod name: %s\n", pod.GetObjectMeta().GetName()))
										//pretty.Print(pod)
										logs, err := podsClient.GetLogs(pod.GetObjectMeta().GetName(), &v1.PodLogOptions{Container: newJob.Name}).Do().Raw()
										if err != nil {
											log.Panic(err)
										}
										fmt.Printf(string(logs[:]))
										//podsClient.Delete(pod.GetObjectMeta().GetName(), &metav1.DeleteOptions{})
									}
									//jobsClient.Delete(newJob.Name, &metav1.DeleteOptions{})
									os.Exit(0)
								}
							}
						}
					}
				}
			},
		},
	)
	// Can't stop won't stop.
	ctrl.Run(wait.NeverStop)

}

// homeDir() returns the user HOME environment whether *nix or win
func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
