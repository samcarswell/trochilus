package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"carswellpress.com/cron-cowboy/core"
	"carswellpress.com/cron-cowboy/data"
)

type slackPost struct {
	Channel string `json:"channel"`
	Text    string `json:"text"`
}

type slackResp struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
}

type runNotifyRow struct {
	Cron      string `json:"cron"`
	Run       int64  `json:"run"`
	SysLog    string `json:"sysLog"`
	Log       string `json:"stdout"`
	Succeeded string `json:"succeeded"`
}

const slackPostMessage = "https://slack.com/api/chat.postMessage"

func NotifyRunSlack(token string, channel string, run data.GetRunRow) (bool, error) {
	runNotifyRow := runNotifyRow{
		Cron:      run.Cron.Name,
		Run:       run.Run.ID,
		SysLog:    run.Run.ExecLogFile,
		Log:       run.Run.LogFile,
		Succeeded: core.FormatSucceeded(run.Run.Succeeded),
	}
	runNotifyRowJson, err := json.Marshal(runNotifyRow)
	if err != nil {
		return false, err
	}
	return NotifySlack(token, channel, string(runNotifyRowJson))
}

func NotifySlack(token string, channel string, text string) (bool, error) {
	postJson, err := json.Marshal(slackPost{
		Channel: channel,
		Text:    text,
	})
	if err != nil {
		return false, err
	}

	r, err := http.NewRequest("POST", slackPostMessage, bytes.NewBuffer(postJson))
	r.Header.Add("Authorization", "Bearer "+token)
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

	fmt.Println(slackResp.Error)
	if res.StatusCode != 200 || !slackResp.Ok {
		return false, nil
	}

	return true, nil
}
