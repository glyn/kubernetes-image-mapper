/*
 * Copyright (c) 2019-Present Pivotal Software, Inc. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"flag"
	"os"
	"time"

	mapperv1alpha1 "github.com/pivotal/kubernetes-image-mapper/api/v1alpha1"
	"github.com/pivotal/kubernetes-image-mapper/controllers"
	"github.com/pivotal/kubernetes-image-mapper/pkg/unimap"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = mapperv1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

const defaultResyncPeriod = time.Hour * 10

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var debug bool
	var resyncPeriod time.Duration
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&debug, "debug", false, "Enable debug logging.")
	flag.DurationVar(&resyncPeriod, "resync-period", defaultResyncPeriod, "The controller resync period.")
	flag.Parse()

	ctrl.SetLogger(zap.Logger(debug))
	setupLog.V(1).Info("debug logging enabled")

	setupLog.Info("creating manager")
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		LeaderElection:     enableLeaderElection,
		SyncPeriod:         &resyncPeriod,
	})
	if err != nil {
		setupLog.Error(err, "failed to create manager")
		os.Exit(1)
	}

	stopCh := ctrl.SetupSignalHandler()
	comp := unimap.New(stopCh)

	setupLog.Info("creating controller", "controller", "ImageMap")
	if err = (&controllers.ImageMapReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("ImageMap"),
		Map:    comp,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "failed to create controller", "controller", "ImageMap")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(stopCh); err != nil {
		setupLog.Error(err, "failed to start manager")
		os.Exit(1)
	}
}
