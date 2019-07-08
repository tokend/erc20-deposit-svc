package query

import (
	"fmt"
	"net/url"
)

type ReviewableRequestFilters struct {
	State              *int
	Reviewer           *string
	Requestor          *string
	PendingTasks       *string
	PendingTasksAnyOf  *string
	PendingTasksNotSet *string
}

type ReviewableRequestIncludes struct {
	RequestDetails bool
}

func (p ReviewableRequestIncludes) Prepare() url.Values {
	result := url.Values{}
	p.prepare(&result)
	return result
}

func (f ReviewableRequestFilters) prepare(values *url.Values) {
	if f.State != nil {
		values.Add("filter[state]", fmt.Sprintf("%d", *f.State))
	}
	if f.Reviewer != nil {
		values.Add("filter[reviewer]", *f.Reviewer)
	}
	if f.Requestor != nil {
		values.Add("filter[requestor]", *f.Requestor)
	}
	if f.PendingTasks != nil {
		values.Add("filter[pending_tasks]", fmt.Sprintf("%s", *f.PendingTasks))
	}
	if f.PendingTasksAnyOf != nil {
		values.Add("filter[pending_tasks_any_of]", fmt.Sprintf("%s", *f.PendingTasksAnyOf))
	}
	if f.PendingTasksNotSet != nil {
		values.Add("filter[pending_tasks_not_set]", fmt.Sprintf("%s",*f.PendingTasksNotSet))
	}

}

func (i ReviewableRequestIncludes) prepare(result *url.Values) {
	if i.RequestDetails {
		result.Add("include", "request_details")
	}
}
