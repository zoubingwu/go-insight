package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
)

var ENV string
var BaseUrl string
var PublicKey string
var PrivateKey string
var OrgId string

type TiInsightSDK struct {
	client *resty.Client
}

func NewInsightSDK() *TiInsightSDK {
	return &TiInsightSDK{
		client: resty.New().
			SetHeader("Content-Type", "application/json").
			SetBaseURL(BaseUrl).
			SetDigestAuth(PublicKey, PrivateKey),
	}
}

type BaseResult struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type TestConnResult struct {
	BaseResult
	Result struct {
		Message string `json:"message"`
		Pass    bool   `json:"pass"`
	} `json:"result"`
}

func (s *TiInsightSDK) TestConnection(conn string) (*TestConnResult, error) {
	var result TestConnResult
	resp, err := s.client.R().
		SetBody(fmt.Sprintf(`{"database_uri":"%s"}`, conn)).
		SetResult(&result).
		Post("/datacontext/connection_check")

	if err != nil {
		return nil, err
	}

	PrettyPrint(resp)
	return &result, nil
}

type CreateContextResult struct {
	BaseResult
	Result struct {
		DataContextId int `json:"data_context_id"`
	} `json:"result"`
}

func (s *TiInsightSDK) CreateContext(conn string) (*CreateContextResult, error) {
	var result CreateContextResult
	resp, err := s.client.R().
		SetBody(fmt.Sprintf(`{"database_uri":"%s","creator":"test","data":{},"org_id":"%s"}`, conn, OrgId)).
		SetResult(&result).
		Post("/datacontext/context")

	if err != nil {
		return nil, err
	}

	PrettyPrint(resp)
	return &result, nil
}

type CreateSessionResult struct {
	BaseResult
	Result struct {
		SessionContextId int `json:"session_context_id"`
	} `json:"result"`
}

func (s *TiInsightSDK) CreateSessionContext(id int) (*CreateSessionResult, error) {
	var result CreateSessionResult
	resp, err := s.client.R().
		SetBody(fmt.Sprintf(`{"data_context_id":"%d","creator":"test","data":{},"org_id":"%s"}`, id, OrgId)).
		SetResult(&result).
		Post("/session/context")

	if err != nil {
		return nil, err
	}

	PrettyPrint(resp)
	return &result, nil
}

type JobResult struct {
	BaseResult
	Result struct {
		JobId string `json:"job_id"`
	} `json:"result"`
}

type TaskResult struct {
	BaseResult
	Result struct {
		EndedAt int    `json:"ended_at"`
		JobId   string `json:"job_id"`
		Reason  string `json:"reason"`
		Result  struct {
			QuestionId       string `json:"question_id"`
			RawQuestion      string `json:"raw_question"`
			SessionContextId int    `json:"session_context_id"`
			TaskTree         map[string]struct {
				Assumptions          []interface{} `json:"assumptions"`
				BreakdownType        string        `json:"breakdown_type"`
				ClarifiedTask        string        `json:"clarified_task"`
				Description          string        `json:"description"`
				Level                int           `json:"level"`
				ParentTask           string        `json:"parent_task"`
				ParentTaskId         string        `json:"parent_task_id"`
				PossibleExplanations string        `json:"possibleExplanations"`
				Reason               string        `json:"reason"`
				SequenceNo           int           `json:"sequence_no"`
				Task                 string        `json:"task"`
				TaskId               string        `json:"task_id"`
			} `json:"task_tree"`
			TimeElapsed float64 `json:"time_elapsed"`
		} `json:"result"`
		Status string `json:"status"`
	} `json:"result"`
}

func (s *TiInsightSDK) BreakdownUserQuestion(prompt string, sessionId int) (*TaskResult, error) {
	var result JobResult
	resp, err := s.client.R().
		SetBody(fmt.Sprintf(`{"raw_question":"%s"}`, prompt)).
		SetResult(&result).
		Post(fmt.Sprintf("/session/%d/actions/question_breakdown", sessionId))

	if err != nil {
		return nil, err
	}
	PrettyPrint(resp)

	var taskResult TaskResult
	for {
		resp, err := s.client.R().SetResult(&taskResult).Get(fmt.Sprintf("/jobs/%s", result.Result.JobId))
		if err != nil {
			return nil, err
		}
		PrettyPrint(resp)
		if taskResult.Result.Status == "done" || taskResult.Result.Status == "failed" {
			return &taskResult, nil
		}

		time.Sleep(2 * time.Second)
	}
}

type Column struct {
	Col string `json:"col"`
}

type ResolvedTaskResult struct {
	BaseResult
	Result struct {
		EndedAt int    `json:"ended_at"`
		JobId   string `json:"job_id"`
		Reason  string `json:"reason"`
		Result  struct {
			Assumptions   []interface{} `json:"assumptions"`
			BreakdownType string        `json:"breakdown_type"`
			ChartOptions  struct {
				ChartName string `json:"chartName"`
				Option    struct {
					Columns []string `json:"columns"`
				} `json:"option"`
				Title string `json:"title"`
			} `json:"chartOptions"`
			ClarifiedTask        string   `json:"clarified_task"`
			Columns              []Column `json:"columns"`
			Description          string   `json:"description"`
			Level                int      `json:"level"`
			ParentTask           string   `json:"parent_task"`
			ParentTaskId         string   `json:"parent_task_id"`
			PossibleExplanations string   `json:"possibleExplanations"`
			Reason               string   `json:"reason"`
			Recommendations      struct {
				Explanation string `json:"explanation"`
				MethodName  string `json:"method_name"`
			} `json:"recommendations"`
			Rows       [][]any `json:"rows"`
			SequenceNo int     `json:"sequence_no"`
			Sql        string  `json:"sql"`
			SqlError   string  `json:"sql_error"`
			Task       string  `json:"task"`
			TaskId     string  `json:"task_id"`
		} `json:"result"`
		Status string `json:"status"`
	} `json:"result"`
}

func (s *TiInsightSDK) FollowupSubTask(sessionId int, taskId string, questionId string) (*ResolvedTaskResult, error) {
	var result JobResult
	resp, err := s.client.R().
		SetBody(fmt.Sprintf(`{"task_id":"%s","question_id":"%s"}`, taskId, questionId)).
		SetResult(&result).
		Post(fmt.Sprintf("/session/%d/actions/text2sql", sessionId))

	if err != nil {
		return nil, err
	}
	PrettyPrint(resp)

	var taskResult ResolvedTaskResult
	for {
		resp, err := s.client.R().SetResult(&taskResult).Get(fmt.Sprintf("/jobs/%s", result.Result.JobId))
		if err != nil {
			return nil, err
		}
		PrettyPrint(resp)
		if taskResult.Result.Status == "done" || taskResult.Result.Status == "failed" {
			return &taskResult, nil
		}

		time.Sleep(2 * time.Second)
	}
}

func PrettyPrint(resp *resty.Response) {
	if ENV == "dev" {
		var result map[string]interface{}
		json.Unmarshal(resp.Body(), &result)
		prettyJSON, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(prettyJSON))
	}
}
