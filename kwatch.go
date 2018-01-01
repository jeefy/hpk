package hpk

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/kylelemons/godebug/pretty"
	apiv1batch "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"

	"gopkg.in/mgo.v2/bson"
)

//func KWatch(clientset clientset.Clientset) {
func KWatch() {
	defer func() { //catch or finally
		if err := recover(); err != nil { //catch
			os.Exit(1)
		}
	}()
	log.SetOutput(os.Stdout)

	session := MongoConnect()

	//jobValues, kubeconfig := ParseCmdArgs()
	_, kubeconfig := ParseCmdArgs()
	clientset := GenerateClientSet(*kubeconfig)
	// Create the job client
	jobsClient := clientset.Batch().Jobs("")
	log.Debug("Kuberntes job client generated.")

	lw := &cache.ListWatch{
		ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
			return jobsClient.List(metav1.ListOptions{})
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return jobsClient.Watch(metav1.ListOptions{})
		},
	}

	_, ctrl := cache.NewInformer(lw,
		&apiv1batch.Job{},
		time.Duration(1000*time.Millisecond),
		cache.ResourceEventHandlerFuncs{
			DeleteFunc: func(obj interface{}) {
				if resource, ok := obj.(*apiv1batch.Job); ok {
					c := session.DB("hpk").C("jobs")

					bsonjson := JSONtoBSON(resource)

					err := c.Update(bson.M{"name": resource.GetName()}, bson.M{"$push": bson.M{"changelog": bsonjson}})
					if err != nil {
						log.Panic(err)
					}
					log.Info(fmt.Sprintf("Job %s@%s deleted\n", resource.Name, resource.Namespace))
				}
			},
			AddFunc: func(obj interface{}) {
				if resource, ok := obj.(*apiv1batch.Job); ok {
					c := session.DB("hpk").C("jobs")
					count, err := c.Find(bson.M{"name": resource.GetName()}).Count()
					if err != nil {
						log.Info(err)
					}
					if count > 0 {
						log.Info(fmt.Sprintf("Duplicate job found: %s", resource.GetName()))
					} else {
						bsonjson := JSONtoBSON(resource)
						err := c.Insert(bson.M{"name": resource.GetName(), "changelog": []interface{}{bsonjson}})
						if err != nil {
							log.Panic(err)
						}
						log.Info(fmt.Sprintf("Job %s@%s created\n", resource.Name, resource.Namespace))
					}
				}
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				if oldObj != newObj {
					if oldJob, ok := oldObj.(*apiv1batch.Job); ok {
						if newJob, ok := newObj.(*apiv1batch.Job); ok {
							if diff := pretty.Compare(oldJob, newJob); diff != "" {
								log.Info(fmt.Sprintf("Job %s@%s updated\n", newJob.Name, newJob.Namespace))
								log.Debug(fmt.Sprintf("%s: Job diff: (-old +new)\n%s", newJob.Name, diff))
								c := session.DB("hpk").C("jobs")
								bsonjson := JSONtoBSON(newJob)
								err := c.Update(bson.M{"name": newJob.GetName()}, bson.M{"$push": bson.M{"changelog": bsonjson}})
								if err != nil {
									log.Panic(err)
								}
								if oldJob.Status.CompletionTime == nil && newJob.Status.CompletionTime != nil {
									c := session.DB("hpk").C("jobs_usage")
									duration := newJob.Status.CompletionTime.Time.Sub(newJob.Status.StartTime.Time).Seconds()
									if duration < 1 {
										duration = 1
									}
									useData := bson.M{
										"name":        newJob.GetName(),
										"start":       newJob.Status.StartTime.Time,
										"end":         newJob.Status.CompletionTime.Time,
										"duration":    duration,
										"cpu_str":     newJob.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu().String(),
										"memory_str":  newJob.Spec.Template.Spec.Containers[0].Resources.Limits.Memory().String(),
										"cpu_val":     newJob.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu().MilliValue(),
										"memory_val":  newJob.Spec.Template.Spec.Containers[0].Resources.Limits.Memory().Value(),
										"cpu_cost":    GetCPUCost(kubeconfig, newJob),
										"memory_cost": GetMemoryCost(kubeconfig, newJob),
										"parallelism": newJob.Spec.Parallelism,
									}

									total_cost := GenerateJobCost(kubeconfig, newJob)
									useData["total_cost"] = fmt.Sprintf("%f", total_cost)

									c.Insert(useData)

									c = session.DB("hpk").C("allocations")
									var m []bson.M
									c.Find(bson.M{"name": newJob.Spec.Template.GetNamespace()}).All(&m)

									balance, err := strconv.ParseFloat(m[0]["balance"].(string), 64)
									if err != nil {
										log.Info("Balance not parsed to float64")
									}

									m[0]["balance"] = fmt.Sprintf("%f", balance-total_cost)

									c.Upsert(bson.M{"name": newJob.Spec.Template.GetNamespace()}, m[0])

									log.Info(fmt.Sprintf("Job %s@%s completed\n", newJob.Name, newJob.Namespace))
									podsClient := clientset.CoreV1().Pods(newJob.Namespace)
									podList, err := podsClient.List(metav1.ListOptions{
										LabelSelector: "job-name = " + newJob.GetName(),
									})
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
										log.Info(fmt.Sprintf("Pod logs scraped for  %s\n", pod.GetObjectMeta().GetName()))
										c := session.DB("hpk").C("jobs_logs")
										err = c.Insert(&hpkJobLog{
											JobName: newJob.Name,
											PodName: pod.GetObjectMeta().GetName(),
											Log:     string(logs[:]),
										})
										if err != nil {
											log.Panic(err)
										}
										//fmt.Printf(string(logs[:]))
									}
									log.Info("Deleting Job: " + newJob.GetName())
									jobsClient.Delete(newJob.GetName(), &metav1.DeleteOptions{})
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

func JSONtoBSON(object interface{}) interface{} {
	jsonObj, err := json.Marshal(object)
	var bsonjson interface{}
	err = bson.UnmarshalJSON(jsonObj, &bsonjson)
	if err != nil {
		panic(err)
	}
	return bsonjson
}
