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

package main

import (
	"flag"

	"github.com/gin-gonic/gin"
	"github.com/saashqdev/kubeworkz/pkg/clients"
	"github.com/saashqdev/kubeworkz/pkg/clog"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"

	"audit/pkg/audit"
	"audit/pkg/backend"
	"audit/pkg/healthz"
	"audit/pkg/listener"
	"audit/pkg/utils/env"
)

const apiPathAuditRoot = "/api/v1/kube/audit"

// @title Swagger Kubeworkz-Audit API
// @version 1.0
// @description This is Kubeworkz-Audit api documentation.
func main() {

	clients.InitCubeClientSetWithOpts(nil)
	logLevel := flag.String("log-level", "info", "log level")
	flag.Parse()
	clog.InitCubeLoggerWithOpts(&clog.Config{
		LogLevel:        *logLevel,
		StacktraceLevel: "error",
	})

	go listener.Listener()

	router := gin.Default()
	router.GET("/healthz", healthz.HealthyCheck)

	url := ginSwagger.URL("/swagger/doc.json") // The url pointing to API definition
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))

	router.GET(apiPathAuditRoot+"/enabled", audit.IsEnabled)
	router.POST(apiPathAuditRoot+"/k8s", audit.HandleK8sAuditLog)
	router.POST(apiPathAuditRoot+"/kube", audit.HandleCubeAuditLog)
	router.POST(apiPathAuditRoot+"/webconsole", audit.HandleWebconsoleAuditLog)
	router.POST(apiPathAuditRoot+"/generic", audit.HandleGenericAuditLog)

	router.GET(apiPathAuditRoot, audit.SearchAuditLog)
	router.GET(apiPathAuditRoot+"/export", audit.ExportAuditLog)

	b := backend.NewBackend()
	go b.Run()

	err := router.Run(":" + env.Port())
	if err != nil {
		clog.Error("%s", err)
	}
}
