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

package audit

import (
	"audit/pkg/backend"
	v1 "audit/pkg/backend/v1"
	"audit/pkg/utils/errcode"
	"audit/pkg/utils/response"

	"github.com/gin-gonic/gin"
	"github.com/saashqdev/kubeworkz/pkg/clog"
)

// receive audit log from Kubeworkz
func HandleCubeAuditLog(c *gin.Context) {

	clog.Info("receive kube audit event")
	event := &v1.Event{}
	if err := c.ShouldBindJSON(event); err != nil {
		clog.Error("unmarshal kubeworkz event failed, error: %s", err)
		response.FailReturn(c, errcode.InvalidBodyFormat)
		return
	}
	response.SuccessReturn(c, nil)

	// send event to channel
	ch := backend.GetCacheCh()
	backend.CacheEvent(ch, event)
}
