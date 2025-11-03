package notify

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"

	"carswellpress.com/trochilus/config"
	"carswellpress.com/trochilus/core"
	"carswellpress.com/trochilus/data"
)

type slackPost struct {
	Channel string `json:"channel"`
	Text    string `json:"text"`
}

type slackResp struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
}

const slackPostMessage = "https://slack.com/api/chat.postMessage"

func NotifyRunSlack(slackConf config.SlackConfig, run data.GetRunRow) (bool, error) {
	slackStr := "*" + run.Cron.Name + "*: run " +
		strconv.FormatInt(run.Run.ID, 10) + tagChannelIfFail(run.Run.Status) +
		"Status: " + core.FormatStatus(core.RunStatus(run.Run.Status)) + "\n" +
		printLogIfExists(run.Run.LogFile)
	return NotifySlack(slackConf, slackStr)
}

func printLogIfExists(logFile string) string {
	if logFile == "" {
		return ""
	}
	return "Log: `" + logFile + "`"
}

func tagChannelIfFail(status string) string {
	if status == string(core.RunStatusFailed) {
		return " <!channel> "
	}
	return " "
}

func NotifySlack(slackConf config.SlackConfig, text string) (bool, error) {
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
