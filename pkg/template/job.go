package template

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateJob(ctx context.Context, c client.Client, instance *batchv1.Job, log logr.Logger) error {
	if err := c.Get(ctx, client.ObjectKeyFromObject(instance), instance); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		log.Info("Creating Job", "Name", instance.Name)
		return c.Create(ctx, instance)
	}
	return nil
}

func DeleteJob(ctx context.Context, c client.Client, instance *batchv1.Job, log logr.Logger) error {
	log.Info("Deleting Job", "Name", instance.Name)
	return c.Delete(ctx, instance)
}

func NewJobRunner(ctx context.Context, c client.Client, instance *batchv1.Job, log logr.Logger) *JobRunner {
	return &JobRunner{
		ctx:      ctx,
		client:   c,
		instance: instance,
		log:      log,
	}
}

type CheckStatusFunc func(hash string) bool
type UpdateStatusFunc func(hash string)

type JobRunner struct {
	ctx      context.Context
	client   client.Client
	instance *batchv1.Job
	log      logr.Logger

	checkStatus  CheckStatusFunc
	updateStatus UpdateStatusFunc
}

func (r *JobRunner) CheckStatus(fn CheckStatusFunc) *JobRunner {
	r.checkStatus = fn
	return r
}

func (r *JobRunner) UpdateStatus(fn UpdateStatusFunc) *JobRunner {
	r.updateStatus = fn
	return r
}

func (r *JobRunner) Run(owner client.Object) (ctrl.Result, error) {
	jobHash, err := ObjectHash(r.instance)
	if err != nil {
		return ctrl.Result{}, err
	}

	if r.checkStatus(jobHash) {
		return ctrl.Result{}, nil
	}

	if err := CreateJob(r.ctx, r.client, r.instance, r.log); err != nil {
		return ctrl.Result{}, err
	} else if r.instance.Status.CompletionTime == nil {
		r.log.Info("Waiting on job completion", "name", r.instance.Name)
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	if err := DeleteJob(r.ctx, r.client, r.instance, r.log); err != nil {
		return ctrl.Result{}, err
	}

	r.updateStatus(jobHash)
	if err := r.client.Status().Update(r.ctx, owner); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}
