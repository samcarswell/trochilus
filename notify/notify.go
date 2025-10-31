package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

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

const slackPostMessage = "https://slack.com/api/chat.postMessage"

func NotifyRunSlack(token string, channel string, run data.GetRunRow) (bool, error) {
	slackStr := "*" + run.Cron.Name + "*: run " +
		strconv.FormatInt(run.Run.ID, 10) + " " +
		core.FormatSucceeded(run.Run.Succeeded) + "\n" +
		"Log: `" + run.Run.LogFile + "`\nSystem Log: `" + run.Run.ExecLogFile + "`"
	return NotifySlack(token, channel, slackStr)
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
