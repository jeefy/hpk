package hpk

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KBatch(): Primary function for the KBatch utility.
//Submits a batch-style job to kubernetes wrapped in an HPC-style interface
func KBatch() {
	defer func() { //catch or finally
		if err := recover(); err != nil { //catch
			os.Exit(1)
		}
	}()
	log.SetOutput(os.Stdout)

	jobValues, kubeconfig := ParseCmdArgs()
	log.Info("Command Args parsed")

	batchConfigMap := ImportScriptFile(flag.Args()[0], jobValues.JobName)
	log.Info("Script converted to ConfigMap")
	log.Debug(fmt.Sprintf("%v", batchConfigMap))

	kJob := GenerateJobTemplate(jobValues)
	log.Info("Generated batch job object")
	log.Debug(fmt.Sprintf("%v", kJob))

	defaultMode := int32(0744)

	kJob.Spec.Template.Spec.Volumes = append(kJob.Spec.Template.Spec.Volumes, v1.Volume{
		Name: "batch-script",
		VolumeSource: v1.VolumeSource{
			ConfigMap: &v1.ConfigMapVolumeSource{
				LocalObjectReference: v1.LocalObjectReference{
					Name: jobValues.JobName + "-config-map",
				},
				DefaultMode: &defaultMode,
			},
		},
	})

	kJob.Spec.Template.Spec.Containers[0].VolumeMounts = append(kJob.Spec.Template.Spec.Containers[0].VolumeMounts, v1.VolumeMount{
		Name:      "batch-script",
		MountPath: "/root/" + flag.Args()[0],
		SubPath:   flag.Args()[0],
	})

	kJob.Spec.Template.Spec.Containers[0].Command = append([]string{"./" + flag.Args()[0]}, flag.Args()[1:]...)

	log.Info("Updated job template with batch specific values")
	log.Debug(fmt.Sprintf("%v", kJob))

	clientset := GenerateClientSet(*kubeconfig)
	log.Info("Kubernetes Client generated")

	configMapsClient := clientset.CoreV1().ConfigMaps(jobValues.Namespace)
	_, err := configMapsClient.Create(batchConfigMap)
	if err != nil {
		log.Panic(err)
	}
	log.Info("ConfigMap submitted to Kubernetes")

	// Create the job client
	jobsClient := clientset.Batch().Jobs(jobValues.Namespace)
	_, err = jobsClient.Create(kJob)
	if err != nil {
		log.Panic(err)
	}
	log.Info("Job submitted to Kubernetes")

}

// ImportScriptFile(): Reads contents of submitted bash script
// and creates a config map to be inserted in job-pods
func ImportScriptFile(scriptPath, jobName string) *v1.ConfigMap {
	contents, err := ioutil.ReadFile(scriptPath)
	if err != nil {
		log.Panic(fmt.Sprintf("%#v", err))
	}
	log.Info("File has been read.")
	log.Debug(fmt.Sprintf("%v", contents))
	configMapName := jobName + "-config-map"

	configMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: configMapName,
		},
		Data: map[string]string{
			scriptPath: string(contents[:]),
		},
	}

	log.Info("ConfigMap object created")
	log.Debug(fmt.Sprintf("%v", configMap))

	return configMap
}
