/*
Copyright 2021.

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

package controllers

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	appv1alpha1 "github.com/nullck/kube-operator-s3-demo/api/v1alpha1"
)

const s3setFinalizer = "cache.example.com/finalizer"

// S3SetReconciler reconciles a S3Set object
type S3SetReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=app.example.com,resources=s3sets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=app.example.com,resources=s3sets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=app.example.com,resources=s3sets/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the S3Set object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *S3SetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("s3set", req.NamespacedName)
	svc := r.createBucketInput()
	s3set := &appv1alpha1.S3Set{}
	err := r.Get(ctx, req.NamespacedName, s3set)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("S3Set resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get S3Set")
		return ctrl.Result{}, err
	}

	// Add finalizer for this CR
	if !controllerutil.ContainsFinalizer(s3set, s3setFinalizer) {
		controllerutil.AddFinalizer(s3set, s3setFinalizer)
		log.Info("Adding Finalizer ...")
		err = r.Update(ctx, s3set)
		if err != nil {
			return ctrl.Result{}, err
		}
	}
	// controller logic
	result := r.listBucket(svc, s3set.Spec.BucketName)
	if result != true {
		log.Info("Creating a new s3 Bucket")
		err = r.createBucket(svc, s3set.Spec.BucketName)
		if err != nil {
			log.Error(err, "Error bucket already exists")
		}
	}

	isS3SetMarkedToBeDeleted := s3set.GetDeletionTimestamp() != nil
	if isS3SetMarkedToBeDeleted {
		if controllerutil.ContainsFinalizer(s3set, s3setFinalizer) {
			result = r.listBucket(svc, s3set.Spec.BucketName)
			if result != false {
				log.Info("Deleting bucket ...")
				err = r.deleteBucket(svc, s3set.Spec.BucketName)
				if err != nil {
					log.Error(err, "Error bucket not founded")
				}
			}
			// Remove s3setFinalizer. Once all finalizers have been
			// removed, the object will be deleted.
			controllerutil.RemoveFinalizer(s3set, s3setFinalizer)
			log.Info("Removing Finalizer ...")
			err := r.Update(ctx, s3set)
			if err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
	}
	log.Info("Finish ...")
	return ctrl.Result{}, nil
}

func (r *S3SetReconciler) createBucketInput() *s3.S3 {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(endpoints.UsEast1RegionID),
	}))
	svc := s3.New(sess, &aws.Config{
		Region: aws.String(endpoints.UsEast1RegionID),
	})
	return svc
}

func (r *S3SetReconciler) createBucket(svc *s3.S3, bucketName string) error {
	input := &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	}
	_, err := svc.CreateBucket(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeBucketAlreadyExists:
				fmt.Println(s3.ErrCodeBucketAlreadyExists, aerr.Error())
			case s3.ErrCodeBucketAlreadyOwnedByYou:
				fmt.Println(s3.ErrCodeBucketAlreadyOwnedByYou, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
		return err
	}
	return nil
}

func (r *S3SetReconciler) deleteBucket(svc *s3.S3, bucketName string) error {
	input := &s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	}

	_, err := svc.DeleteBucket(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
		return err
	}
	return nil
}

func (r *S3SetReconciler) listBucket(svc *s3.S3, bucketName string) bool {
	input := &s3.ListBucketsInput{}

	result, err := svc.ListBuckets(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
	}
	found := strings.Contains(result.GoString(), bucketName)
	return found
}

// SetupWithManager sets up the controller with the Manager.
func (r *S3SetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1alpha1.S3Set{}).
		Complete(r)
}
