package api

import (
	"encoding/base64"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/lemonsoul/jenkins-cli/config"
	"github.com/lemonsoul/jenkins-cli/util"
	"github.com/tidwall/gjson"
)

func base_req(api string) ([]byte, *http.Header) {
	cfg := util.GetConfigFile()
	req, _ := http.NewRequest("GET", cfg.BaseApi+api, nil)
	auth := cfg.Username + ":" + cfg.Token
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Add("Authorization", "Basic "+encodedAuth)
	response, _ := http.DefaultClient.Do(req)
	resBody, _ := io.ReadAll(response.Body)
	return resBody, &response.Header
}

func GetViews() []string {
	resBody, _ := base_req("/api/json")
	viewNames := gjson.Get(string(resBody), "views.#.name")
	viewRes := make([]string, 0)
	for _, viewName := range viewNames.Array() {
		viewRes = append(viewRes, viewName.String())
	}
	return viewRes
}

func GetViewJob(username string, token string, baseApi string, viewName string) []string {
	resBody, _ := base_req("/view/" + viewName + "/api/json")
	jobNames := gjson.Get(string(resBody), "jobs.#.name")
	jobRes := make([]string, 0)
	for _, jobName := range jobNames.Array() {
		jobRes = append(jobRes, jobName.String())
	}
	return jobRes
}

func GetJobPararm(jobName string) ([]string, []string) {
	resBody, _ := base_req("/job/" + jobName + "/api/json")
	parameterDefinitions := gjson.Get(string(resBody), `property.#(_class=="hudson.model.ParametersDefinitionProperty").parameterDefinitions`)
	choices := gjson.Get(parameterDefinitions.Raw, "#(_class==\"hudson.model.ChoiceParameterDefinition\").choices")
	choicesArray := make([]string, 0)
	for _, item := range choices.Array() {
		choicesArray = append(choicesArray, item.String())
	}

	branchs := gjson.Get(parameterDefinitions.Raw, "#(_class==\"net.uaznia.lukanus.hudson.plugins.gitparameter.GitParameterDefinition\").allValueItems.values.#.value")
	branchArray := make([]string, 0)
	for _, item := range branchs.Array() {
		branchArray = append(branchArray, item.String())
	}
	return choicesArray, branchArray
}

func GetCrumb() (string, string) {
	resBody, _ := base_req("/crumbIssuer/api/json")
	resutArray := gjson.GetMany(string(resBody), "crumbRequestField", "crumb")
	return resutArray[0].String(), resutArray[1].String()
}

func BuildWithParameters(jobName string, choices string, branch string) string {
	crumbRequestField, crumb := GetCrumb()
	if crumbRequestField == "" || crumb == "" {
		panic("failed to get crumb")
	}
	data := url.Values{}
	data.Set("tag", branch)
	data.Set("pro", choices)
	encodedData := data.Encode()

	cfg := util.GetConfigFile()

	req, _ := http.NewRequest("POST", cfg.BaseApi+"/job/"+jobName+"/buildWithParameters", strings.NewReader(encodedData))

	auth := cfg.Username + ":" + cfg.Token
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Add("Authorization", "Basic "+encodedAuth)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add(crumbRequestField, crumb)

	response, _ := http.DefaultClient.Do(req)
	//resBody, _ := io.ReadAll(response.Body)
	if response.StatusCode == 201 {
		location := response.Header.Get("Location")
		u, err := url.Parse(location)
		if err != nil {
			panic("failed to parse url")
		}
		lastPart := path.Base(u.Path)
		return lastPart
	} else {
		return ""
	}
}

func GetBuildNumber(queueId string) string {
	resBody, _ := base_req("/queue/item/" + queueId + "/api/json")
	return gjson.Get(string(resBody), "executable.number").String()
}

func GetBuildLog(jobName string, buildNumber string) string {
	resBody, _ := base_req("/job/" + jobName + "/" + buildNumber + "/logText/progressiveText/api/json")
	log := gjson.Get(string(resBody), "log").String()
	return log
}

func GetQueue() []config.Queue {
	resBody, _ := base_req("/queue/api/json")
	items := gjson.Get(string(resBody), "items")

	queueArray := make([]config.Queue, 0)
	for _, item := range items.Array() {
		var queueItem config.Queue
		queueItem.Id = item.Get("id").Str
		queueItem.TaskName = item.Get("task.name").Str
		queueItem.Params = item.Get("params").Str
		queueItem.Why = item.Get("why").Str
		queueItem.Blocked = item.Get("blocked").Bool()
		queueItem.Stuck = item.Get("stuck").Bool()
		queueItem.InQueueSince = item.Get("inQueueSince").Str
		queueArray = append(queueArray, queueItem)
	}
	return queueArray
}

func GetComputer() []config.Computer {
	resBody, _ := base_req("/computer/api/json?depth=1")
	computerArray := make([]config.Computer, 0)

	computers := gjson.Get(string(resBody), "computer")
	computers.ForEach(func(_, comp gjson.Result) bool {
		executors := comp.Get("oneOffExecutors")
		executors.ForEach(func(_, exec gjson.Result) bool {
			execInfo := exec.Get("currentExecutable")
			if execInfo.Exists() {
				computerItem := config.Computer{}
				computerItem.BuildNumber = int(execInfo.Get("number").Int())
				re := regexp.MustCompile(`/job/([^/]+)`)
				matches := re.FindStringSubmatch(execInfo.Get("url").String())
				if len(matches) >= 2 {
					computerItem.JobName = matches[1]
				}
				computerArray = append(computerArray, computerItem)
			}
			return true
		})
		return true
	})
	return computerArray
}

func GetBuildStatus(jobName string, buildNumber string) config.BuildStatus {
	resBody, _ := base_req("/job/" + jobName + "/" + buildNumber + "/api/json")
	res := gjson.GetMany(string(resBody), "queueId", "number", "building", "duration", "fullDisplayName")
	return config.BuildStatus{
		QueueId:         res[0].String(),
		BuildNumber:     res[1].String(),
		Building:        res[2].Bool(),
		Duration:        int(res[3].Int()),
		FullDisplayName: res[4].String(),
	}
}

func GetTextLog(jobName string, buildNumber string, start *int) (string, bool, int) {
	reqUrl := "/job/" + jobName + "/" + buildNumber + "/logText/progressiveText"
	if start != nil {
		reqUrl = reqUrl + "?start=" + strconv.Itoa(*start)
	}
	resBody, resHeader := base_req(reqUrl)
	logText := string(resBody)
	moreDate := false
	textSize := int(-1)
	moreDateTemp := resHeader.Get("X-More-Data")
	textSizetemp := resHeader.Get("X-Text-Size")
	if moreDateTemp != "" {
		parsed, err := strconv.ParseBool(moreDateTemp)
		if err == nil {
			moreDate = parsed
		}
	}
	if textSizetemp != "" {
		parsed, err := strconv.ParseInt(textSizetemp, 10, 64)
		if err == nil {
			textSize = int(parsed)
		}
	}
	return logText, moreDate, textSize

}

func GetPipelineConfig(jobName string) config.PipelineConfig {
	resBody, _ := base_req("/job/" + jobName + "/wfapi/runs")
	res := gjson.GetMany(string(resBody), "id", "status", "startTimeMillis", "endTimeMillis", "durationMillis", "stages")
	stages := make([]config.Stage, 0)
	for _, stage := range res[5].Array() {
		stage := config.Stage{
			Id:              stage.Get("id").String(),
			Name:            stage.Get("name").String(),
			Status:          stage.Get("status").String(),
			StartTimeMillis: stage.Get("startTimeMillis").Int(),
			DurationMillis:  stage.Get("durationMillis").Int(),
		}
		stages = append(stages, stage)
	}
	return config.PipelineConfig{
		Stages: stages,
	}
}

func GetWFDescribe(jobName string, buildNumber string) config.WFDescribe {
	resBody, _ := base_req("/job/" + jobName + "/" + buildNumber + "/wfapi/describe")
	res := gjson.GetMany(string(resBody), "id", "status", "startTimeMillis", "endTimeMillis", "durationMillis", "stages")
	stages := make([]config.Stage, 0)
	for _, stage := range res[5].Array() {
		stage := config.Stage{
			Id:              stage.Get("id").String(),
			Name:            stage.Get("name").String(),
			Status:          stage.Get("status").String(),
			StartTimeMillis: stage.Get("startTimeMillis").Int(),
			DurationMillis:  stage.Get("durationMillis").Int(),
		}
		stages = append(stages, stage)
	}
	return config.WFDescribe{
		QueueId:         res[0].String(),
		Status:          res[1].String(),
		StartTimeMillis: res[2].Int(),
		EndTimeMillis:   res[3].Int(),
		DurationMillis:  res[4].Int(),
		Stages:          stages,
	}
}
