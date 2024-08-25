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

package env

import "os"

const (
	defaultEsHost  = "http://elasticsearch-master.elasticsearch:9200"
	defaultEsIndex = "audit"
	defaultEsType  = "logs"
	defaultPort    = "8888"
)

type EsWebhook struct {
	Host  string
	Index string
	Type  string
}

func Webhook() *EsWebhook {
	host := os.Getenv("AUDIT_WEBHOOK_HOST")
	index := os.Getenv("AUDIT_WEBHOOK_INDEX")
	types := os.Getenv("AUDIT_WEBHOOK_TYPE")
	if host == "" || index == "" || types == "" {
		return nil
	}
	return &EsWebhook{host, index, types}
}

func ElasticSearchHost() *EsWebhook {
	if Webhook() == nil {
		return &EsWebhook{defaultEsHost, defaultEsIndex, defaultEsType}
	} else {
		return Webhook()
	}
}

func JwtSecret() string {
	return os.Getenv("JWT_SECRET")
}

func Port() string {
	p := os.Getenv("PORT")
	if p == "" {
		return defaultPort
	}
	return p
}
