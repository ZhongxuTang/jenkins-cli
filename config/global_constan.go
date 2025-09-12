package config

const BASE_NAME = "jenkins-cli"
const WORKSPACE_INFO = "workspace"

const BASE_CONFIG_DIR = "/.config/" + BASE_NAME + "/" + BASE_NAME + ".yaml"
const WORKSPACE_INFO_DIR = "/.config/" + BASE_NAME + "/" + WORKSPACE_INFO + ".yaml"

type JenkinsConfig struct {
	Username string `yaml:"username"`
	Token    string `yaml:"token"`
	BaseApi  string `yaml:"base_api"`
}

type Workspace struct {
	Views []View `yaml:"views"`
}

type View struct {
	Name string `yaml:"name"`
	Job  []Job  `yaml:"job"`
}

type Job struct {
	Name     string   `yaml:"name"`
	JobParam JobParam `yaml:"job_param"`
}

type JobParam struct {
	Choices []string `yaml:"choices"`
	Branch  []string `yaml:"branch"`
}

type Queue struct {
	Id           string
	TaskName     string
	Params       string
	Why          string
	Blocked      bool
	Stuck        bool
	InQueueSince string
}

type Computer struct {
	BuildNumber int
	JobName     string
}

type BuildInfo struct {
	QueueId         string `json:"queueId"`
	BuildNumber     string `json:"buildNumber"`
	Building        bool   `json:"building"`
	Duration        int    `json:"duration"`
	FullDisplayName string `json:"fullDisplayName"`
	ChangeSets      []ChangeSet
}

type ChangeSet struct {
	CommitId       string
	Timestamp      string
	AuthorFullName string
	Comment        string
}

type PipelineConfig struct {
	Stages []Stage `json:"stages"`
}

type WFDescribe struct {
	QueueId         string  `json:"queueId"`
	Status          string  `json:"status"`
	StartTimeMillis int64   `json:"startTimeMillis"`
	EndTimeMillis   int64   `json:"endTimeMillis"`
	DurationMillis  int64   `json:"durationMillis"`
	Stages          []Stage `json:"stages"`
}

type Stage struct {
	Id                  string `json:"id"`
	Name                string `json:"name"`
	Status              string `json:"status"`
	StartTimeMillis     int64  `json:"startTimeMillis"`
	DurationMillis      int64  `json:"durationMillis"`
	PauseDurationMillis int64  `json:"pauseDurationMillis"`
}
