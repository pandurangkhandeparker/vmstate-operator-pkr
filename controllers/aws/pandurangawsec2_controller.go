/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package aws

import (
	"context"
	"encoding/json"
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	// "knative.dev/pkg/configmap"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	awsv1 "github.com/pandurangkhandeparker/vmstate-operator-pkr/apis/aws/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// PandurangAWSEC2Reconciler reconciles a PandurangAWSEC2 object
type PandurangAWSEC2Reconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=aws.pandurang.com,resources=pandurangawsec2s,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=aws.pandurang.com,resources=pandurangawsec2s/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=aws.pandurang.com,resources=pandurangawsec2s/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the PandurangAWSEC2 object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile

const AWSEC2Finalizer = "aws.pandurang.com/finalizer"

func (r *PandurangAWSEC2Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// _ = log.FromContext(ctx)

	// // TODO(user): your logic here

	// return ctrl.Result{}, nil
	//_ = log.FromContext(ctx)
	log := ctrllog.FromContext(ctx)
	//log := r.Log.WithValues("AWSEC2", req.NamespacedName)
	log.Info("Reconciling AWSEC2s CRs")

	// Fetch the AWSEC2 CR
	//awsEC2, err := services.FetchAWSEC2CR(req.Name, req.Namespace)

	// Fetch the AWSEC2 instance
	awsEC2 := &awsv1.PandurangAWSEC2{}
	//ctrl.SetControllerReference(awsEC2, awsEC2, r.Scheme)
	log.Info(req.NamespacedName.Name)

	err := r.Client.Get(ctx, req.NamespacedName, awsEC2)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("awsEC2 resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get awsEC2.")
		return ctrl.Result{}, err
	}
	log.Info(awsEC2.Name)
	// Add const values for mandatory specs ( if left blank)
	// log.Info("Adding awsEC2 mandatory specs")
	// utils.AddBackupMandatorySpecs(awsEC2)
	// Check if the Job already exists, if not create a new one

	awsEC2.Status.EC2Status = "Starting a Job"
	fmt.Println(awsEC2.Status.EC2Status)

	r.recorder.Event(awsEC2, corev1.EventTypeNormal, "Starting", fmt.Sprintf("Starting a Job"))

	found := &batchv1.Job{}
	err = r.Client.Get(ctx, types.NamespacedName{Name: awsEC2.Name + "create", Namespace: awsEC2.Namespace}, found)
	//log.Info(*found.)
	if err != nil && errors.IsNotFound(err) {
		// Define a new Job
		job := r.JobForAWSEC2(ctx, awsEC2, "create")
		log.Info("Creating a new Job", "job.Namespace", job.Namespace, "job.Name", job.Name)

		awsEC2.Status.EC2Status = "In Progress: Creating EC2 instance"
		fmt.Println(awsEC2.Status.EC2Status)

		r.recorder.Event(awsEC2, corev1.EventTypeNormal, "In Progress", fmt.Sprintf("In Progress: Creating EC2 instance  %s/%s", awsEC2.Namespace, awsEC2.Name))

		err = r.Client.Create(ctx, job)
		if err != nil {
			log.Error(err, "Failed to create new Job", "job.Namespace", job.Namespace, "job.Name", job.Name)

			awsEC2.Status.EC2Status = "Failed"
			fmt.Println(awsEC2.Status.EC2Status)

			r.recorder.Event(awsEC2, corev1.EventTypeNormal, "Failed", fmt.Sprintf("Failed"))

			return ctrl.Result{}, err
		}
		// job created successfully - return and requeue

		awsEC2.Status.EC2Status = "Completed"
		fmt.Println(awsEC2.Status.EC2Status)

		r.recorder.Event(awsEC2, corev1.EventTypeNormal, "Completed", fmt.Sprintf("Completed Job"))

		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get job")
		return ctrl.Result{}, err
	}

	// Check for any updates for redeployment
	/*applyChange := false

	// Ensure image name is correct, update image if required
	newInstanceIds := awsEC2.Spec.InstanceIds
	log.Info(newInstanceIds)

	newStartSchedule := awsEC2.Spec.StartSchedule
	log.Info(newStartSchedule)

	newImage := awsEC2.Spec.Image
	log.Info(newImage)

	var currentImage string = ""
	var currentStartSchedule string = ""
	var currentInstanceIds string = ""

	// Check existing schedule
	if found.Spec.Schedule != "" {
		currentStartSchedule = found.Spec.Schedule
	}

	if newStartSchedule != currentStartSchedule {
		found.Spec.Schedule = newStartSchedule
		applyChange = true
	}

	// Check existing image
	if found.Spec.JobTemplate.Spec.Template.Spec.Containers != nil {
		currentImage = found.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Image
	}

	if newImage != currentImage {
		found.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Image = newImage
		applyChange = true
	}

	// Check instanceIds
	if found.Spec.JobTemplate.Spec.Template.Spec.Containers != nil {
		currentInstanceIds = found.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Env[0].Value
		log.Info(currentInstanceIds)
	}

	if newInstanceIds != currentInstanceIds {
		found.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Env[0].Value = newInstanceIds
		applyChange = true
	}

	log.Info(currentInstanceIds)
	log.Info(currentImage)
	log.Info(currentStartSchedule)

	log.Info(strconv.FormatBool(applyChange))

	if applyChange {
		log.Info(strconv.FormatBool(applyChange))
		err = r.Client.Update(ctx, found)
		if err != nil {
			log.Error(err, "Failed to update jobJob", "jobJob.Namespace", found.Namespace, "jobJob.Name", found.Name)
			return ctrl.Result{}, err
		}
		// Spec updated - return and requeue
		return ctrl.Result{Requeue: true}, nil
	}*/

	// Update the AWSEC2 status
	// TODO: Define what needs to be added in status. Currently adding just instanceIds
	/*if !reflect.DeepEqual(currentInstanceIds, awsEC2.Status.VMStartStatus) ||
		!reflect.DeepEqual(currentInstanceIds, awsEC2.Status.VMStopStatus) {
		awsEC2.Status.VMStartStatus = currentInstanceIds
		awsEC2.Status.VMStopStatus = currentInstanceIds
		err := r.Client.Status().Update(ctx, awsEC2)
		if err != nil {
			log.Error(err, "Failed to update awsEC2 status")
			return ctrl.Result{}, err
		}
	}*/
	// Check if the AWSEC2 instance is marked to be deleted, which is
	// indicated by the deletion timestamp being set.
	isAWSEC2MarkedToBeDeleted := awsEC2.GetDeletionTimestamp() != nil
	if isAWSEC2MarkedToBeDeleted {
		if controllerutil.ContainsFinalizer(awsEC2, AWSEC2Finalizer) {
			// Run finalization logic for AWSEC2Finalizer. If the
			// finalization logic fails, don't remove the finalizer so
			// that we can retry during the next reconciliation.
			log.Info(awsEC2.Name)
			log.Info("CR is marked for deletion")
			if err := r.finalizeAWSEC2(ctx, awsEC2); err != nil {
				return ctrl.Result{}, err
			}

			// Remove AWSEC2Finalizer. Once all finalizers have been
			// removed, the object will be deleted.
			controllerutil.RemoveFinalizer(awsEC2, AWSEC2Finalizer)
			err := r.Client.Update(ctx, awsEC2)
			if err != nil {
				return ctrl.Result{}, err
			}
			log.Info("Finalizer removed")
			log.Info(awsEC2.Name)
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer for this CR
	if !controllerutil.ContainsFinalizer(awsEC2, AWSEC2Finalizer) {
		log.Info("Finalizer added again")
		log.Info(awsEC2.Name)
		controllerutil.AddFinalizer(awsEC2, AWSEC2Finalizer)
		err = r.Client.Update(ctx, awsEC2)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *PandurangAWSEC2Reconciler) finalizeAWSEC2(ctx context.Context, awsEC2 *awsv1.PandurangAWSEC2) error {
	// TODO(user): Add the cleanup steps that the operator
	// needs to do before the CR can be deleted. Examples
	// of finalizers include performing backups and deleting
	// resources that are not owned by this CR, like a PVC.
	log := ctrllog.FromContext(ctx)
	log.Info("Successfully finalized AWSEC2")
	found := &batchv1.Job{}
	err := r.Client.Get(ctx, types.NamespacedName{Name: awsEC2.Name + "delete", Namespace: awsEC2.Namespace}, found)
	//log.Info(*found.)
	if err != nil && errors.IsNotFound(err) {
		// Define a new job
		awsEC2.Status.EC2Status = "In Progress"
		fmt.Println(awsEC2.Status.EC2Status)
		job := r.JobForAWSEC2(ctx, awsEC2, "delete")
		log.Info("Creating a new Job", "job.Namespace", job.Namespace, "job.Name", job.Name)
		err = r.Client.Create(ctx, job)
		if err != nil {
			log.Error(err, "Failed to create new Job", "job.Namespace", job.Namespace, "job.Name", job.Name)
			awsEC2.Status.EC2Status = "Failed"
			fmt.Println(awsEC2.Status.EC2Status)
			return err
		}
		// job created successfully - return and requeue
		awsEC2.Status.EC2Status = "Completed"
		fmt.Println(awsEC2.Status.EC2Status)
		return nil
	} else if err != nil {
		log.Error(err, "Failed to get job")
		return err
	}
	return nil
}

type AWScm struct {
	InstanceType string `json:"instanceType"`
	ImageId      string `json:"imageId"`
}

// Job Spec
func (r *PandurangAWSEC2Reconciler) JobForAWSEC2(ctx context.Context, awsEC2 *awsv1.PandurangAWSEC2, command string) *batchv1.Job {

	var awscm AWScm
	// cfgData, _ := configmap.Load("aws-configmap")
	// json.Unmarshal([]byte(cfgData["aws.json"]), &cm)

	cfg, _ := rest.InClusterConfig()
	clientset, _ := kubernetes.NewForConfig(cfg)

	cm, _ := clientset.CoreV1().ConfigMaps(awsEC2.Namespace).Get(ctx, awsEC2.Spec.CfgMap, metav1.GetOptions{})

	cfgData := cm.Data["aws.json"]

	err := json.Unmarshal([]byte(cfgData), &awscm)
	if err != nil {
		panic(err)
	}

	jobName := awsEC2.Name + command
	job := &batchv1.Job{
		ObjectMeta: v1.ObjectMeta{
			Name:      jobName,
			Namespace: awsEC2.Namespace,
			Labels:    AWSEC2Labels(awsEC2, "awsEC2"),
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{{
						Name: "configmap-volume",
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "aws-configmap",
								},
							},
						},
					}},
					Containers: []corev1.Container{{
						Name:  awsEC2.Name,
						Image: awsEC2.Spec.Image,
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "configmap-volume",
							MountPath: "/opt/aws.json",
						}},
						Env: []corev1.EnvVar{
							{
								Name:  "ec2_command",
								Value: awsEC2.Spec.Command,
							},
							{
								Name:  "ec2_tag_key",
								Value: awsEC2.Spec.TagKey,
							},
							{
								Name:  "ec2_tag_value",
								Value: awsEC2.Spec.TagValue,
							},
							{
								Name:  "ec2_instance_type",
								Value: awscm.InstanceType,
							},
							{
								Name:  "ec2_image_id",
								Value: awscm.ImageId,
							},
							{
								Name: "AWS_ACCESS_KEY_ID",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: "aws-secret",
										},
										Key: "aws-access-key-id",
									},
								},
							},
							{
								Name: "AWS_SECRET_ACCESS_KEY",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: "aws-secret",
										},
										Key: "aws-secret-access-key",
									},
								},
							},
							{
								Name: "AWS_DEFAULT_REGION",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: "aws-secret",
										},
										Key: "aws-default-region",
									},
								},
							}},
						ImagePullPolicy: awsEC2.Spec.ImagePullPolicy,
					}},
					RestartPolicy: awsEC2.Spec.RestartPolicy,
				},
			},
		},
	}
	// Set awsEC2 instance as the owner and controller
	ctrl.SetControllerReference(awsEC2, job, r.Scheme)
	return job
}

func AWSEC2Labels(v *awsv1.PandurangAWSEC2, tier string) map[string]string {
	return map[string]string{
		"app":       "AWSEC2",
		"AWSEC2_cr": v.Name,
		"tier":      tier,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *PandurangAWSEC2Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.recorder = mgr.GetEventRecorderFor("PandurangAWSEC2")

	return ctrl.NewControllerManagedBy(mgr).
		For(&awsv1.PandurangAWSEC2{}).
		Owns(&batchv1.Job{}).
		Complete(r)
}
