// Copyright 2026 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package event

import (
	"fmt"
	"math"
	"net/url"
	"strconv"
	"time"

	"github.com/chaos-mesh/chaos-mesh/pkg/config"
	u "github.com/chaos-mesh/chaos-mesh/pkg/dashboard/apiserver/utils"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
)

type listRequest struct {
	Query  core.EventQuery
	Legacy *core.Filter
	Page   int
	Limit  int
}

type PageResponse struct {
	Events     []*core.Event `json:"events" validate:"required"`
	Page       int           `json:"page" validate:"required"`
	Limit      int           `json:"limit" validate:"required"`
	NextPage   *int          `json:"next_page" validate:"required" extensions:"x-nullable"`
	Total      *int64        `json:"total,omitempty"`
	TotalPages *int64        `json:"total_pages,omitempty"`
} // @name EventPageResponse

func parseListRequest(values url.Values, conf *config.ChaosDashboardConfig, paged bool) (listRequest, error) {
	request := listRequest{Query: core.EventQuery{
		Namespaces: u.EffectiveNamespaces(values["namespace"], conf),
		Names:      u.NormalizeQueryValues(values["name"]),
		Kinds:      u.NormalizeQueryValues(values["kind"]),
		ObjectIDs:  u.NormalizeQueryValues(values["object_id"]),
		Sort:       core.EventSortID,
		Order:      core.EventOrderDesc,
	}}
	for _, key := range []string{"page", "limit", "include_total", "sort", "order", "start", "end"} {
		if len(values[key]) > 1 {
			return listRequest{}, fmt.Errorf("parameter %s must occur once", key)
		}
	}
	if !paged && (values.Has("page") || values.Has("include_total")) {
		return listRequest{}, fmt.Errorf("pagination parameters require /events/page")
	}
	enhanced := paged || values.Has("sort") || values.Has("order")
	for _, key := range []string{"namespace", "name", "kind", "object_id"} {
		enhanced = enhanced || len(values[key]) > 1
	}
	if !enhanced {
		start, _ := time.Parse(time.RFC3339, values.Get("start"))
		end, _ := time.Parse(time.RFC3339, values.Get("end"))
		request.Legacy = &core.Filter{
			Namespace: firstValue(request.Query.Namespaces),
			Name:      firstValue(request.Query.Names),
			Kind:      firstValue(request.Query.Kinds),
			ObjectID:  firstValue(request.Query.ObjectIDs),
			Start:     start.UTC().Format(layout),
			End:       end.UTC().Format(layout),
			Limit:     values.Get("limit"),
		}
		return request, nil
	}
	if values.Has("sort") {
		request.Query.Sort = core.EventSort(values.Get("sort"))
		switch request.Query.Sort {
		case core.EventSortID, core.EventSortObjectID, core.EventSortNamespace, core.EventSortName,
			core.EventSortKind, core.EventSortCreatedAt, core.EventSortMessage:
		default:
			return listRequest{}, fmt.Errorf("unsupported event sort")
		}
	}
	if values.Has("order") {
		request.Query.Order = core.EventOrder(values.Get("order"))
		if request.Query.Order != core.EventOrderAsc && request.Query.Order != core.EventOrderDesc {
			return listRequest{}, fmt.Errorf("parameter order must be asc or desc")
		}
	}
	for _, bound := range []struct {
		key    string
		target **time.Time
	}{
		{"start", &request.Query.Start}, {"end", &request.Query.End},
	} {
		if values.Has(bound.key) {
			instant, err := time.Parse(time.RFC3339Nano, values.Get(bound.key))
			if err != nil {
				return listRequest{}, fmt.Errorf("parameter %s must be RFC3339: %w", bound.key, err)
			}
			*bound.target = &instant
		}
	}
	if paged {
		page, err := strconv.Atoi(values.Get("page"))
		if err != nil || page < 1 {
			return listRequest{}, fmt.Errorf("parameter page must be a positive integer")
		}
		request.Page = page
		request.Limit = 12
	}
	if values.Has("limit") {
		limit, err := strconv.Atoi(values.Get("limit"))
		if err != nil {
			return listRequest{}, fmt.Errorf("parameter limit must be an integer")
		}
		request.Limit = limit
		if !paged && limit >= 0 {
			if limit == math.MaxInt {
				return listRequest{}, fmt.Errorf("parameter limit is too large")
			}
			request.Query.Window = &core.EventWindow{Limit: limit}
		}
	}
	if paged {
		if request.Limit < 1 || request.Limit > 1000 {
			return listRequest{}, fmt.Errorf("parameter limit must be between 1 and 1000")
		}
		if request.Page-1 > (math.MaxInt-request.Limit-1)/request.Limit {
			return listRequest{}, fmt.Errorf("page offset is too large")
		}
		request.Query.Window = &core.EventWindow{
			Offset: (request.Page - 1) * request.Limit,
			Limit:  request.Limit,
		}
		if values.Has("include_total") {
			includeTotal, err := strconv.ParseBool(values.Get("include_total"))
			if err != nil {
				return listRequest{}, fmt.Errorf("parameter include_total must be a boolean")
			}
			request.Query.IncludeTotal = includeTotal
		}
	}
	return request, nil
}

func firstValue(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
