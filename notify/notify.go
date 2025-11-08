package notify

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/samcarswell/trochilus/config"
	"github.com/samcarswell/trochilus/core"
)

type slackPost struct {
	Channel string `json:"channel"`
	Text    string `json:"text"`
}

type slackResp struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
}

type RunNotifyInfo struct {
	Name    string
	Id      int64
	Status  core.RunStatus
	LogFile string
}

const slackPostMessage = "https://slack.com/api/chat.postMessage"

func NotifyRun(
	notifyConf config.NotifyConfig,
	run RunNotifyInfo,
) (bool, error) {
	slackStr := getNotifyText(run.Name, run.Id, run.Status, run.LogFile, notifyConf.Hostname)
	return notifySlack(notifyConf.Slack, slackStr)
}

// Returns the notification test for a run.
// This is designed to ignore incorrect inputs; ensuring a notification is sent
// is critical; if it's missing some information, that's acceptable.
func getNotifyText(
	cronName string,
	runId int64,
	runStatus core.RunStatus,
	logFile string,
	hostname string,
) string {
	return "*" + cronName + hostnameIfExists(hostname) + "*: run " +
		strconv.FormatInt(runId, 10) + " - " +
		core.FormatStatus(runStatus) +
		tagChannelIfFail(runStatus) +
		logFileIfExists(logFile)
}

func hostnameIfExists(hostname string) string {
	if hostname == "" {
		return ""
	}
	return "@" + hostname
}

func logFileIfExists(logFile string) string {
	if logFile == "" {
		return ""
	}
	return "\nLog: `" + logFile + "`"
}

func tagChannelIfFail(status core.RunStatus) string {
	if status == core.RunStatusFailed {
		return " <!channel>"
	}
	return ""
}

func notifySlack(slackConf config.SlackConfig, text string) (bool, error) {
	postJson, err := json.Marshal(slackPost{
		Channel: slackConf.Channel,
		Text:    text,
	})
	if err != nil {
		return false, err
	}

	r, err := http.NewRequest("POST", slackPostMessage, bytes.NewBuffer(postJson))
	r.Header.Add("Authorization", "Bearer "+slackConf.Token)
	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("charset", "utf-8")

	client := &http.Client{}
	res, err := client.Do(r)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	slackResp := &slackResp{}
	err = json.NewDecoder(res.Body).Decode(slackResp)
	if err != nil {
		return false, err
	}

	if res.StatusCode != 200 || !slackResp.Ok {
		return false, nil
	}

	return true, nil
}
