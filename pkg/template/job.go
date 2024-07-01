package template

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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
	return c.Delete(ctx, instance,
		client.PropagationPolicy(metav1.DeletePropagationForeground))
}

func NewJobRunner(ctx context.Context, c client.Client, instance client.Object, log logr.Logger) *JobRunner {
	return &JobRunner{
		ctx:      ctx,
		client:   c,
		instance: instance,
		log:      log,
	}
}

type JobRunner struct {
	ctx      context.Context
	client   client.Client
	instance client.Object
	log      logr.Logger

	jobs []jobHashField
}

func (r *JobRunner) Add(hashField *string, job *batchv1.Job) {
	r.jobs = append(r.jobs, jobHashField{
		Job:       job,
		HashField: hashField,
	})
}

func (r *JobRunner) Run(ctx context.Context, report ReportFunc) (ctrl.Result, error) {
	for _, jh := range r.jobs {
		if result, err := r.run(ctx, jh.Job, jh.HashField, report); err != nil || !result.IsZero() {
			return result, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *JobRunner) run(ctx context.Context, job *batchv1.Job, hashField *string, report ReportFunc) (ctrl.Result, error) {
	controllerutil.SetControllerReference(r.instance, job, r.client.Scheme())

	jobHash, err := ObjectHash(job)
	if err != nil {
		return ctrl.Result{}, err
	}

	if *hashField == jobHash {
		// matching job has already run
		return ctrl.Result{}, nil
	}

	if err := CreateJob(r.ctx, r.client, job, r.log); err != nil {
		return ctrl.Result{}, err
	} else if job.Status.CompletionTime == nil {
		if err := report(ctx, "Waiting on Job %s condition Complete", job.Name); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	if err := DeleteJob(r.ctx, r.client, job, r.log); err != nil {
		return ctrl.Result{}, err
	}

	// update status field with last run job hash
	*hashField = jobHash

	if err := r.client.Status().Update(r.ctx, r.instance); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

type jobHashField struct {
	Job       *batchv1.Job
	HashField *string
}
