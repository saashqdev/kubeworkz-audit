/*
Copyright 2024 KubeWorkz Authors

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
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/saashqdev/kubeworkz/pkg/clog"
	"github.com/saashqdev/kubeworkz/pkg/utils/constants"
	"k8s.io/apiserver/pkg/apis/audit"
)

const eventResourceK8s = "[Kubernetes]"

// receive audit log from K8s
func HandleK8sAuditLog(c *gin.Context) {

	clog.Info("receive k8s audit event list")
	eventList := &audit.EventList{}
	if err := c.ShouldBindJSON(&eventList); err != nil {
		clog.Error("unmarshal k8s event list failed, error: %s", err)
	}

	for _, event := range eventList.Items {
		// transform K8s event to v1.event
		e := &v1.Event{
			EventTime:       event.StageTimestamp.Unix(),
			EventVersion:    "V1",
			SourceIpAddress: event.SourceIPs[0],
			RequestMethod:   event.Verb,
			ResponseStatus:  int(event.ResponseStatus.Code),
			Url:             event.RequestURI,
			UserIdentity:    &v1.UserIdentity{AccountId: event.User.Username},
			UserAgent:       event.UserAgent,
			EventType:       constants.EventTypeUserWrite,
			RequestId:       string(event.AuditID),
		}
		e = getEventName(e)
		if e.ResponseStatus != http.StatusOK {
			e.ErrorCode = strconv.Itoa(e.ResponseStatus)
		}

		clog.Info("audit event from k8s: %+v", e)

		// send event to channel
		ch := backend.GetCacheCh()
		backend.CacheEvent(ch, e)
	}

}

func getEventName(e *v1.Event) *v1.Event {
	var object string
	url := e.Url
	queryUrl := strings.Split(url, "?")[0]
	urlstrs := strings.Split(queryUrl, "/")
	for i, str := range urlstrs {
		if str == constants.K8sResourceNamespace {
			if i+2 < len(urlstrs) {
				object = urlstrs[i+2]
			} else {
				object = constants.K8sResourceNamespace
			}
			break
		}

		if str == constants.K8sResourceVersion && i+1 < len(urlstrs) && urlstrs[i+1] != constants.K8sResourceNamespace {
			object = urlstrs[i+1]
		}
	}
	e.ResourceReports = []v1.Resource{{ResourceType: object}}
	e.EventName = eventResourceK8s + " " + e.RequestMethod + " " + object
	e.Description = e.EventName
	return e
}
