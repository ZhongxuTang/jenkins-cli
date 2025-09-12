package api

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/lemonsoul/jenkins-cli/config"
	"github.com/lemonsoul/jenkins-cli/util"
	"github.com/tidwall/gjson"
)

func buildRequest(req *http.Request) error {
	cfg, err := util.GetConfigFile()
	if err != nil {
		return err
	}
	auth := cfg.Username + ":" + cfg.Token
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Add("Authorization", "Basic "+encodedAuth)
	return nil
}

func baseReq(api string) ([]byte, int, *http.Header, error) {
	cfg, err := util.GetConfigFile()
	if err != nil {
		return nil, -1, nil, err
	}

	req, err := http.NewRequest("GET", cfg.BaseApi+api, nil)
	if err != nil {
		return nil, -1, nil, fmt.Errorf("failed to create request: %w", err)
	}

	if err := buildRequest(req); err != nil {
		return nil, -1, nil, err
	}

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		color.Red("request failed: %v", err)
		return nil, -1, nil, fmt.Errorf("request failed: %w", err)
	}

	resBody, ioErr := io.ReadAll(response.Body)
	if ioErr != nil {
		color.Red("io error: %v", ioErr)
		return nil, response.StatusCode, &response.Header, fmt.Errorf("failed to read response: %w", ioErr)
	}

	return resBody, response.StatusCode, &response.Header, nil
}

func GetViews() ([]string, error) {
	resBody, _, _, err := baseReq("/api/json")
	if err != nil {
		return nil, err
	}
	viewNames := gjson.Get(string(resBody), "views.#.name")
	viewRes := make([]string, 0)
	for _, viewName := range viewNames.Array() {
		viewRes = append(viewRes, viewName.String())
	}
	return viewRes, nil
}

func GetViewJob(username string, token string, baseApi string, viewName string) ([]string, error) {
	resBody, _, _, err := baseReq("/view/" + viewName + "/api/json")
	if err != nil {
		return nil, err
	}
	jobNames := gjson.Get(string(resBody), "jobs.#.name")
	jobRes := make([]string, 0)
	for _, jobName := range jobNames.Array() {
		jobRes = append(jobRes, jobName.String())
	}
	return jobRes, nil
}

func GetJobPararm(jobName string) ([]string, []string, error) {
	resBody, _, _, err := baseReq("/job/" + jobName + "/api/json")
	if err != nil {
		return nil, nil, err
	}
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
	return choicesArray, branchArray, nil
}

func GetCrumb() (string, string, error) {
	resBody, _, _, err := baseReq("/crumbIssuer/api/json")
	if err != nil {
		return "", "", err
	}
	resutArray := gjson.GetMany(string(resBody), "crumbRequestField", "crumb")
	return resutArray[0].String(), resutArray[1].String(), nil
}

func BuildWithParameters(jobName string, choices string, branch string) (string, error) {
	crumbRequestField, crumb, err := GetCrumb()
	if err != nil {
		return "", fmt.Errorf("failed to get crumb: %w", err)
	}
	if crumbRequestField == "" || crumb == "" {
		return "", fmt.Errorf("received empty crumb from Jenkins")
	}

	data := url.Values{}
	data.Set("tag", branch)
	data.Set("pro", choices)
	encodedData := data.Encode()

	cfg, err := util.GetConfigFile()
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", cfg.BaseApi+"/job/"+jobName+"/buildWithParameters", strings.NewReader(encodedData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	auth := cfg.Username + ":" + cfg.Token
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Add("Authorization", "Basic "+encodedAuth)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add(crumbRequestField, crumb)

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode == 201 {
		location := response.Header.Get("Location")
		u, err := url.Parse(location)
		if err != nil {
			return "", fmt.Errorf("failed to parse response location: %w", err)
		}
		lastPart := path.Base(u.Path)
		return lastPart, nil
	} else {
		return "", fmt.Errorf("build failed with status code: %d", response.StatusCode)
	}
}

func GetBuildNumber(queueId string) (string, error) {
	resBody, _, _, err := baseReq("/queue/item/" + queueId + "/api/json")
	if err != nil {
		return "", err
	}
	return gjson.Get(string(resBody), "executable.number").String(), nil
}

func GetBuildLog(jobName string, buildNumber string) (string, error) {
	resBody, _, _, err := baseReq("/job/" + jobName + "/" + buildNumber + "/logText/progressiveText/api/json")
	if err != nil {
		return "", err
	}
	log := gjson.Get(string(resBody), "log").String()
	return log, nil
}

func GetQueue() ([]config.Queue, error) {
	resBody, _, _, err := baseReq("/queue/api/json")
	if err != nil {
		return nil, err
	}
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
	return queueArray, nil
}

func GetComputer() ([]config.Computer, error) {
	resBody, _, _, err := baseReq("/computer/api/json?depth=1")
	if err != nil {
		return nil, err
	}
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
	return computerArray, nil
}

func GetBuildStatus(jobName string, buildNumber string) (config.BuildInfo, error) {
	resBody, _, _, err := baseReq("/job/" + jobName + "/" + buildNumber + "/api/json")
	if err != nil {
		return config.BuildInfo{}, err
	}
	res := gjson.GetMany(string(resBody), "queueId", "number", "building", "duration", "fullDisplayName", "changeSets")
	buildStatus := config.BuildInfo{
		QueueId:         res[0].String(),
		BuildNumber:     res[1].String(),
		Building:        res[2].Bool(),
		Duration:        int(res[3].Int()),
		FullDisplayName: res[4].String(),
	}

	changeSets := make([]config.ChangeSet, 0)
	for _, changeSetItem := range res[5].Array() {
		changeSetItem.Get("items").ForEach(func(_, item gjson.Result) bool {
			changeSet := config.ChangeSet{
				CommitId:       item.Get("commitId").String(),
				Timestamp:      item.Get("timestamp").String(),
				AuthorFullName: item.Get("author.fullName").String(),
				Comment:        item.Get("comment").String(),
			}
			changeSets = append(changeSets, changeSet)
			return true
		})
	}
	buildStatus.ChangeSets = changeSets
	return buildStatus, nil
}

func GetTextLog(jobName string, buildNumber string, start *int) (string, bool, int, error) {
	reqUrl := "/job/" + jobName + "/" + buildNumber + "/logText/progressiveText"
	if start != nil {
		reqUrl = reqUrl + "?start=" + strconv.Itoa(*start)
	}
	resBody, _, resHeader, err := baseReq(reqUrl)
	if err != nil {
		return "", false, -1, err
	}
	logText := string(resBody)
	moreDate := false
	textSize := int(-1)
	if resHeader != nil {
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
	}
	return logText, moreDate, textSize, nil
}

func GetPipelineConfig(jobName string) (config.PipelineConfig, error) {
	resBody, _, _, err := baseReq("/job/" + jobName + "/wfapi/runs")
	if err != nil {
		return config.PipelineConfig{}, err
	}
	res := gjson.Get(string(resBody), "#.stages")
	configStages := make([]config.Stage, 0)
	configIds := make([]string, 0)
	for _, stages := range res.Array() {
		for _, stage := range stages.Array() {
			if slices.Contains(configIds, stage.Get("id").String()) {
				continue
			}
			currentStage := config.Stage{
				Id:              stage.Get("id").String(),
				Name:            stage.Get("name").String(),
				Status:          stage.Get("status").String(),
				StartTimeMillis: stage.Get("startTimeMillis").Int(),
				DurationMillis:  stage.Get("durationMillis").Int(),
			}
			configIds = append(configIds, currentStage.Id)
			configStages = append(configStages, currentStage)
		}
	}
	return config.PipelineConfig{
		Stages: configStages,
	}, nil
}

func GetWFDescribe(jobName string, buildNumber string) (config.WFDescribe, error) {
	resBody, _, _, err := baseReq("/job/" + jobName + "/" + buildNumber + "/wfapi/describe")
	if err != nil {
		return config.WFDescribe{}, err
	}
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
	}, nil
}

func Stop(jobName string, buildNumber string) (bool, error) {
	cfg, err := util.GetConfigFile()
	if err != nil {
		return false, err
	}

	req, err := http.NewRequest("POST", cfg.BaseApi+"/job/"+jobName+"/"+buildNumber+"/stop", nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	if err := buildRequest(req); err != nil {
		return false, err
	}

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		color.Red("request failed: %v", err)
		return false, fmt.Errorf("request failed: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode == 200 || response.StatusCode == 302 {
		return true, nil
	}
	return false, fmt.Errorf("stop request failed with status code: %d", response.StatusCode)
}

func CancleItem(queueId string) (bool, error) {
	if queueId == "" {
		return false, fmt.Errorf("queue ID cannot be empty")
	}
	_, statusCode, _, err := baseReq("/queue/cancelItem?id=" + queueId)
	if err != nil {
		return false, err
	}
	if statusCode == 200 || statusCode == 302 {
		return true, nil
	}
	return false, fmt.Errorf("cancel request failed with status code: %d", statusCode)
}
