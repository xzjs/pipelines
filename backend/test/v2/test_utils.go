// Copyright 2018-2023 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package test

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/cenkalti/backoff"
	experimentparams "github.com/kubeflow/pipelines/backend/api/v2beta1/go_http_client/experiment_client/experiment_service"
	"github.com/kubeflow/pipelines/backend/api/v2beta1/go_http_client/experiment_model"
	api_server "github.com/kubeflow/pipelines/backend/src/common/client/api_server/v2"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func WaitForReady(namespace string, initializeTimeout time.Duration) error {
	operation := func() error {
		response, err := http.Get(fmt.Sprintf("http://ml-pipeline.%s.svc.cluster.local:8888/apis/v2beta1/healthz", namespace))
		if err != nil {
			return err
		}

		// If we get a 503 service unavailable, it's a non-retriable error.
		if response.StatusCode == 503 {
			return backoff.Permanent(errors.Wrapf(
				err, "Waiting for ml pipeline API server failed with non retriable error."))
		}

		return nil
	}

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = initializeTimeout
	err := backoff.Retry(operation, b)
	return errors.Wrapf(err, "Waiting for ml pipeline API server failed after all attempts.")
}

func GetClientConfig(namespace string) clientcmd.ClientConfig {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	overrides := clientcmd.ConfigOverrides{Context: clientcmdapi.Context{Namespace: namespace}}
	return clientcmd.NewInteractiveDeferredLoadingClientConfig(loadingRules,
		&overrides, os.Stdin)
}

func ListAllExperiment(client *api_server.ExperimentClient, namespace string) ([]*experiment_model.V2beta1Experiment, int, string, error) {
	return ListExperiment(client, &experimentparams.ListExperimentsParams{}, namespace)
}

func ListExperiment(client *api_server.ExperimentClient, parameters *experimentparams.ListExperimentsParams, namespace string) ([]*experiment_model.V2beta1Experiment, int, string, error) {
	if namespace != "" {
		parameters.Namespace = &namespace
	}
	return client.List(parameters)
}

func DeleteAllExperiments(client *api_server.ExperimentClient, namespace string, t *testing.T) {
	experiments, _, _, err := ListAllExperiment(client, namespace)
	assert.Nil(t, err)
	for _, e := range experiments {
		if e.DisplayName != "Default" {
			assert.Nil(t, client.Delete(&experimentparams.DeleteExperimentParams{ExperimentID: e.ExperimentID}))
		}
	}
}
