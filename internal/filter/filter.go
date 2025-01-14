package filter

import (
	"context"

	"github.com/go-logr/logr"
	ootov1alpha1 "github.com/qbarrand/oot-operator/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func hasLabel(label string) predicate.Predicate {
	return predicate.NewPredicateFuncs(func(o client.Object) bool {
		return o.GetLabels()[label] != ""
	})
}

var skipDeletions predicate.Predicate = predicate.Funcs{
	DeleteFunc: func(_ event.DeleteEvent) bool { return false },
}

type Filter struct {
	client client.Client
	logger logr.Logger
}

func New(client client.Client, logger logr.Logger) *Filter {
	return &Filter{
		client: client,
		logger: logger,
	}
}

func (f *Filter) ModuleReconcilerNodePredicate(kernelLabel string) predicate.Predicate {
	return predicate.And(
		skipDeletions,
		hasLabel(kernelLabel),
		predicate.LabelChangedPredicate{},
	)
}

func (f *Filter) NodeKernelReconcilerPredicate(labelName string) predicate.Predicate {
	labelMismatch := predicate.NewPredicateFuncs(func(o client.Object) bool {
		return o.GetLabels()[labelName] != o.(*v1.Node).Status.NodeInfo.KernelVersion
	})

	return predicate.And(skipDeletions, labelMismatch)
}

func (f *Filter) FindModulesForNode(node client.Object) []reconcile.Request {
	logger := f.logger.WithValues("node", node.GetName())

	reqs := make([]reconcile.Request, 0)

	logger.Info("Listing all modules")

	mods := ootov1alpha1.ModuleList{}

	if err := f.client.List(context.Background(), &mods); err != nil {
		logger.Error(err, "could not list modules")
		return reqs
	}

	logger.Info("Listed modules", "count", len(mods.Items))

	nodeLabelsSet := labels.Set(node.GetLabels())

	for _, mod := range mods.Items {
		logger := logger.WithValues("module name", mod.Name)

		logger.V(1).Info("Processing module")

		sel := labels.NewSelector()

		for k, v := range mod.Spec.Selector {
			logger.V(1).Info("Processing selector item", "key", k, "value", v)

			requirement, err := labels.NewRequirement(k, selection.Equals, []string{v})
			if err != nil {
				logger.Error(err, "could not generate requirement: %v", err)
				return reqs
			}

			sel = sel.Add(*requirement)
		}

		if !sel.Matches(nodeLabelsSet) {
			logger.V(1).Info("Node labels do not match the module's selector; skipping")
			continue
		}

		nsn := types.NamespacedName{Name: mod.Name, Namespace: mod.Namespace}

		reqs = append(reqs, reconcile.Request{NamespacedName: nsn})
	}

	logger.Info("Adding reconciliation requests", "count", len(reqs))
	logger.V(1).Info("New requests", "requests", reqs)

	return reqs
}
