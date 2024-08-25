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
	"audit/pkg/utils/auth"
	"audit/pkg/utils/env"
	"audit/pkg/utils/errcode"
	"audit/pkg/utils/response"
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic/v7"
	"github.com/saashqdev/kubeworkz/pkg/authentication/authenticators/token"
	"github.com/saashqdev/kubeworkz/pkg/authorizer/rbac"
	"github.com/saashqdev/kubeworkz/pkg/clog"
	"github.com/saashqdev/kubeworkz/pkg/utils/constants"
)

const (
	exportQueryEventMaxSize = 10000
)

type auditQuery struct {
	UserName        string `form:"userName,omitempty"`
	SourceIpAddress string `form:"sourceIpAddress,omitempty"`
	ResourceName    string `form:"resourceName,omitempty"`
	EventName       string `form:"eventName,omitempty"`
	ResponseStatus  int    `form:"responseStatus,omitempty"`
	StartTime       int64  `form:"startTime,omitempty"`
	EndTime         int64  `form:"endTime,omitempty"`
	Page            int    `form:"page,omitempty"`
	Size            int    `form:"size,omitempty"`
	SortBy          string `form:"sortBy,omitempty"`
	SortAsc         bool   `form:"sortAsc,omitempty"`
}

type EsResult struct {
	Total  int64
	Events []v1.Event
}

// @Summary query audit log
// @Description query audit log from es
// @Tags audit
// @Param	query	query	auditQuery  false  "key and value for query"
// @Success 200 {object} EsResult
// @Failure 500 {object} errcode.ErrorInfo
// @Router /api/v1/kube/audit  [get]
func SearchAuditLog(c *gin.Context) {

	if !backend.SendElasticSearch {
		response.FailReturn(c, errcode.New(&errcode.ErrorInfo{Code: http.StatusBadRequest, Message: "Audit or ElasticSearch is disabled."}))
		return
	}

	// authority check
	user := auth.GetUserFromReq(c)
	if user == "" {
		response.FailReturn(c, errcode.AuthenticateError)
		return
	}
	if !checkIsAdmin(user) {
		response.FailReturn(c, errcode.NoAuthority)
		return
	}

	var query auditQuery
	if err := c.Bind(&query); err != nil {
		clog.Error("parse search audit log param error: %s", err)
		response.FailReturn(c, errcode.InvalidBodyFormat)
		return
	}
	if query.Page < 0 {
		query.Page = 0
	}
	if query.Size <= 0 {
		query.Size = 10
	}

	result, err := searchLog(query)
	if err != nil {
		response.FailReturn(c, err)
		return
	}
	response.SuccessReturn(c, result)
}

// @Summary export audit log
// @Description query and export audit log from es
// @Tags audit
// @Param	query	query	auditQuery  false  "key and value for query"
// @Success 200 {string} string
// @Failure 500 {object} errcode.ErrorInfo
// @Router /api/v1/kube/audit/export  [get]
func ExportAuditLog(c *gin.Context) {

	// authority check
	user, userErr := token.GetUserFromReq(c.Request)
	if userErr != nil {
		response.FailReturn(c, errcode.AuthenticateError)
		return
	}
	if !checkIsAdmin(user.Username) {
		response.FailReturn(c, errcode.NoAuthority)
	}

	var query auditQuery
	if err := c.Bind(&query); err != nil {
		clog.Error("parse search audit log param error: %s", err)
		response.FailReturn(c, errcode.InvalidBodyFormat)
		return
	}
	query.Page = 0
	query.Size = exportQueryEventMaxSize

	result, err := searchLog(query)
	if err != nil {
		response.FailReturn(c, err)
		return
	}
	if result.Total == 0 {
		response.FailReturn(c, errcode.NotFound)
		return
	}
	//for i := 0; i <= int(result.Total)/exportEventMaxSize; i++ {
	//	end := 0
	//	if (i+1)*exportEventMaxSize < int(result.Total) {
	//		end = (i + 1) * exportEventMaxSize
	//	} else {
	//		end = int(result.Total)
	//	}

	dataBytes, err := writeCsv(result.Events)
	if err != nil {
		response.FailReturn(c, err)
		return
	}

	fileName := strconv.FormatInt(time.Now().Unix(), 10)
	c.Writer.Header().Set(constants.HttpHeaderContentType, constants.HttpHeaderContentTypeOctet)
	c.Writer.Header().Set(constants.HttpHeaderContentDisposition, fmt.Sprintf("attachment;filename=%s.csv", fileName))
	c.Data(http.StatusOK, "text/csv", dataBytes.Bytes())
	//}

}

func writeCsv(events []v1.Event) (bytes.Buffer, *errcode.ErrorInfo) {
	data := [][]string{
		{"eventID", "userIdentity", "time", "IPAddress", "eventName", "requestMethod", "requestParams", "statusCode", "Url"},
	}
	for _, event := range events {
		timef := time.Unix(event.EventTime, 0).Format("2006-01-02 15:04:05")
		var accountId string
		if event.UserIdentity == nil {
			accountId = ""
		} else {
			accountId = event.UserIdentity.AccountId
		}
		data = append(data, []string{event.RequestId, accountId, timef, event.SourceIpAddress, event.EventName, event.RequestMethod, event.RequestParameters, strconv.Itoa(event.ResponseStatus), event.Url})
	}

	dataBytes := &bytes.Buffer{}
	dataBytes.WriteString("\xEF\xBB\xBF")
	wr := csv.NewWriter(dataBytes)

	if err := wr.WriteAll(data); err != nil {
		clog.Error("write user template file error: %s", err)
		return *dataBytes, errcode.InternalServerError
	}
	// clear
	wr.Flush()
	return *dataBytes, nil
}

func searchLog(query auditQuery) (EsResult, *errcode.ErrorInfo) {

	var esResult EsResult
	// connect to es
	client, err := elastic.NewClient(elastic.SetSniff(false), elastic.SetURL(env.ElasticSearchHost().Host))
	if err != nil {
		clog.Error("connect to elasticsearch error: %s, url: %s ", err, env.ElasticSearchHost().Host)
		return esResult, errcode.InternalServerError
	}

	// structure filter
	boolQ := elastic.NewBoolQuery()

	// filter username
	if len(strings.TrimSpace(query.UserName)) > 0 {
		boolQ.Filter(elastic.NewTermQuery("UserIdentity.AccountId", query.UserName))
	}

	// filter time
	if query.EndTime > 0 && query.StartTime <= 0 {
		boolQ.Filter(elastic.NewRangeQuery("EventTime").Lte(query.EndTime))
	} else if query.EndTime <= 0 && query.StartTime > 0 {
		boolQ.Filter(elastic.NewRangeQuery("EventTime").Gte(query.StartTime))
	} else if query.EndTime > 0 && query.StartTime > 0 {
		boolQ.Filter(elastic.NewRangeQuery("EventTime").Lte(query.EndTime).Gte(query.StartTime))
	}

	// filter ip
	if len(strings.TrimSpace(query.SourceIpAddress)) > 0 {
		boolQ.Filter(elastic.NewTermQuery("SourceIpAddress", query.SourceIpAddress))
	}

	// fuzzy filter resource name
	if len(strings.TrimSpace(query.ResourceName)) > 0 {
		boolQ.Must(elastic.NewMatchQuery("ResourceReports.ResourceName", query.ResourceName))
	}

	// fuzzy filter event name
	if len(strings.TrimSpace(query.EventName)) > 0 {
		boolQ.Must(elastic.NewMatchQuery("EventName", query.EventName))
	}

	// filter status code
	if query.ResponseStatus > 0 {
		boolQ.Filter(elastic.NewTermQuery("ResponseStatus", query.ResponseStatus))
	}

	if query.SortBy == "" {
		query.SortBy = "EventTime"
	}

	res, err := client.Search().
		Index(env.ElasticSearchHost().Index).
		Query(boolQ).
		From((query.Page-1)*query.Size).
		Size(query.Size).
		Sort(query.SortBy, query.SortAsc).
		Do(context.Background())
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			clog.Debug("search audit log from es error: %s", err)
			return esResult, nil
		}
		clog.Error("search audit log from es error: %s", err)
		return esResult, errcode.InternalServerError
	}

	if res != nil && res.Hits.TotalHits.Value > 0 {
		esResult.Total = res.Hits.TotalHits.Value
		for _, hit := range res.Hits.Hits {
			var event v1.Event
			err = json.Unmarshal(hit.Source, &event)
			if err != nil {
				clog.Error("json unmarshal audit log error: %s", err)
				continue
			}
			esResult.Events = append(esResult.Events, event)
		}
	} else {
		esResult.Total = 0
		clog.Debug("search audit log result from es is 0")
	}
	return esResult, nil
}

func checkIsAdmin(userName string) bool {
	h := rbac.NewDefaultResolver(constants.LocalCluster)
	user, err := h.GetUser(userName)
	if err != nil {
		clog.Error(err.Error())
		return false
	}

	_, clusterRoles, err := h.RolesFor(rbac.User2UserInfo(user.Name), "")
	if err != nil {
		clog.Error(err.Error())
		return false
	}
	for _, clusterRole := range clusterRoles {
		if clusterRole.Name == constants.PlatformAdmin {
			return true
		}
	}
	return false
}

func IsEnabled(c *gin.Context) {
	response.SuccessReturn(c, backend.SendElasticSearch)
}
