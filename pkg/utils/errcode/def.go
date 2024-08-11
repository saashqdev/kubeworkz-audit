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

package errcode

import (
	"fmt"
	"net/http"
)

type ErrorInfo struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func New(errorInfo *ErrorInfo, params ...interface{}) *ErrorInfo {
	return &ErrorInfo{
		Code:    errorInfo.Code,
		Message: fmt.Sprintf(errorInfo.Message, params...),
	}
}

var (
	// common
	internalServerError = &ErrorInfo{http.StatusInternalServerError, "Server is busy, please try again."}
	invalidBodyFormat   = &ErrorInfo{http.StatusBadRequest, "Body format invalid."}

	// auth
	noAuthority       = &ErrorInfo{http.StatusForbidden, "No Authority"}
	authenticateError = &ErrorInfo{http.StatusUnauthorized, "Authenticate failed."}

	notFound = &ErrorInfo{http.StatusNotFound, "No result found."}
)
