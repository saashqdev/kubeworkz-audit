/*
Copyright 2024 Kubeworkz Authors

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

package listener

import (
	"audit/pkg/backend"
	"audit/pkg/utils/env"
	"context"

	"github.com/saashqdev/kubeworkz/pkg/apis"
	hotplugv1 "github.com/saashqdev/kubeworkz/pkg/apis/hotplug/v1"
	"github.com/saashqdev/kubeworkz/pkg/clog"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	toolcache "k8s.io/client-go/tools/cache"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
)

const (
	hotPlugNameCommon                 = "common"
	hotPlugComponentNameAudit         = "audit"
	hotPlugComponentEnabled           = "enabled"
	hotPlugComponentNameElasticsearch = "elasticsearch"
)

var (
	auditEnable                 = false
	internalElasticSearchEnable = false
	elasticSearchEnable         = false
)

func Listener() {

	config, err := ctrl.GetConfig()
	if err != nil {
		clog.Error("get ctrl config error: %s", err)
		return
	}

	scheme := runtime.NewScheme()
	utilruntime.Must(apis.AddToScheme(scheme))

	c, err := cache.New(config, cache.Options{Scheme: scheme})
	if err != nil {
		clog.Error("new cache error: %s", err)
		return
	}

	ctx := context.Background()
	hotPlug := &hotplugv1.Hotplug{}
	hotPlug.Name = hotPlugNameCommon
	informer, err := c.GetInformer(ctx, hotPlug)
	if err != nil {
		clog.Error("get informer error: %s", err)
		return
	}

	procFunc := func(newObj interface{}) {
		hotplug, ok := newObj.(*hotplugv1.Hotplug)
		if !ok {
			clog.Error("watch an error obj: %+v", newObj)
			return
		}
		components := hotplug.Spec.Component
		if hotplug.Spec.Component == nil {
			backend.SendElasticSearch = false
			return
		}
		for _, component := range components {
			if component.Name == hotPlugComponentNameAudit {
				if component.Status == hotPlugComponentEnabled {
					auditEnable = true
				} else {
					backend.SendElasticSearch = false
					return
				}
			}
			if component.Name == hotPlugComponentNameElasticsearch {
				if component.Status == hotPlugComponentEnabled {
					internalElasticSearchEnable = true
				}
			}
		}
		if env.Webhook() != nil {
			elasticSearchEnable = true
		} else if internalElasticSearchEnable {
			elasticSearchEnable = true
		}
		if auditEnable && elasticSearchEnable {
			backend.SendElasticSearch = true
		}
	}

	informer.AddEventHandler(toolcache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			procFunc(newObj)
		},
		AddFunc: func(obj interface{}) {
			procFunc(obj)
		},
	})

	err = c.Start(ctx)
	if err != nil {
		clog.Error("start cache error: %s", err)
		return
	}
}
