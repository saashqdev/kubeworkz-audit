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
	"time"

	"github.com/gin-gonic/gin"
	"github.com/saashqdev/kubeworkz/pkg/clog"
)

const eventResourceWebconsole = "[Webconsole]"

type webconsoleAuditMsg struct {
	SessionID     string    `json:"session_id"`
	CreateTime    time.Time `json:"create_time"`
	PodName       string    `json:"pod_name,omitempty"`
	Namespace     string    `json:"namespace,omitempty"`
	ClusterName   string    `json:"cluster_name,omitempty"`
	Data          string    `json:"data"`
	DataType      string    `json:"data_type"` //stdin, stdout
	RemoteIP      string    `json:"remote_ip,omitempty"`
	UserAgent     string    `json:"user_agent,omitempty"`
	ContainerUser string    `json:"container_user,omitempty"`
	WebUser       string    `json:"web_user,omitempty"`
	Platform      string    `json:"platform,omitempty"`
}

// receive audit log from webconsole
func HandleWebconsoleAuditLog(c *gin.Context) {

	clog.Info("receive webconsole audit event")
	msg := &webconsoleAuditMsg{}
	if err := c.ShouldBindJSON(msg); err != nil {
		clog.Error("unmarshal event from webconsole err: %v", err)
		response.FailReturn(c, errcode.InvalidBodyFormat)
		return
	}
	response.SuccessReturn(c, nil)

	event, err := buildEvent(msg)
	if err != nil {
		clog.Error("build event with audit message err: %v", err)
		return
	}
	// send event to channel
	ch := backend.GetCacheCh()
	backend.CacheEvent(ch, event)
}

func buildEvent(msg *webconsoleAuditMsg) (*v1.Event, error) {
	event := &v1.Event{
		EventTime: msg.CreateTime.Unix(),
		EventName: eventResourceWebconsole + " " + msg.Data,
		Description: "ClusterName: " + msg.ClusterName + ", Namespace: " + msg.Namespace +
			", ContainerUser: " + msg.ContainerUser + ", Platform: " + msg.Platform,
		SourceIpAddress:   msg.RemoteIP,
		UserAgent:         msg.UserAgent,
		RequestId:         msg.SessionID,
		RequestParameters: msg.Data,
		EventType:         msg.DataType,
		ResourceReports:   []v1.Resource{{ResourceType: "Pod", ResourceId: "", ResourceName: msg.PodName}},
		UserIdentity:      &v1.UserIdentity{AccountId: msg.WebUser},
	}
	return event, nil
}
