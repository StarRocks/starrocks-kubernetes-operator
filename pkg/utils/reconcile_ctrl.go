/*
Copyright 2022 StarRocks.

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
package utils

import (
	"errors"
	"time"

	k8s_errors "k8s.io/apimachinery/pkg/api/errors"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var (
	ErrFinalizerUnfinished = errors.New("unfinished")
)

// synced completed
func OK() (reconcile.Result, error) {
	return reconcile.Result{}, nil
}

// requeue
func Retry(after int, err error) (reconcile.Result, error) {
	return reconcile.Result{Requeue: true, RequeueAfter: time.Second * time.Duration(after)}, err
}

// sync failed, need to reque
func Failed(err error) (reconcile.Result, error) {
	// ignore some errors which we don not want to log
	if k8s_errors.IsConflict(err) {
		err = nil
	}
	if err == ErrFinalizerUnfinished {
		err = nil
	}
	return Retry(2, err)
}
