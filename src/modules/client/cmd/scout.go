package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/troopstack/troop/src/modules/client/utils"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

type ScoutOpeRequest struct {
	Hostname string `json:"hostname"`
}

var Scout = &cobra.Command{
	Use:   "scout",
	Short: "For The Scout's Operation",
	Args:  cobra.MaximumNArgs(2),
	RunE:  scout,
}
var ScoutList bool
var ScoutAccept string
var ScoutReject string
var ScoutDelete string

func init() {
	Scout.PersistentFlags().BoolVarP(&ScoutList, "list", "l", false,
		"get scout list")
	Scout.PersistentFlags().StringVarP(&ScoutAccept, "accept", "a", "",
		"accept scout")
	Scout.PersistentFlags().StringVarP(&ScoutReject, "reject", "r", "",
		"reject scout")
	Scout.PersistentFlags().StringVarP(&ScoutDelete, "delete", "d", "",
		"delete scout")
}

func scout(c *cobra.Command, args []string) error {
	var err error
	if ScoutList {
		err = scoutList()
	} else if ScoutAccept != "" {
		err = scoutAccept()
	} else if ScoutReject != "" {
		err = scoutReject()
	} else if ScoutDelete != "" {
		err = scoutDelete()
	} else {
		err = scoutList()
	}
	return err
}

type Host struct {
	Hostname     string
	Ip           string
	ScoutVersion string
	Plugins      string
	Os           string
}

type Scouts struct {
	Accepted   []*Host
	Unaccepted []string
	Denied     []string
}

type ScoutResult struct {
	Result Scouts `json:"result"`
}

func scoutList() error {
	url := utils.Config().General.Addresses + "/hosts"
	req, _ := http.NewRequest("GET", url, nil)
	req = utils.HttpHandler(req)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		close(sendTaskChan)
		fmt.Println("Error: Can not connection general.")
		os.Exit(1)
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	ScoutResult := ScoutResult{}
	err = json.Unmarshal(body, &ScoutResult)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println(string(body))
	} else {
		scoutKeys := ScoutResult.Result
		color.HiGreen("[Accepted Scouts (%d)]", len(scoutKeys.Accepted))
		for i := range scoutKeys.Accepted {
			//color.Cyan(fmt.Sprintf("%s [Version: %s] [OS: %s] [IP: %s] [Plugins: %s]",
			//	scoutKeys.Accepted[i].Hostname, scoutKeys.Accepted[i].ScoutVersion,
			//	scoutKeys.Accepted[i].Os, scoutKeys.Accepted[i].Ip,
			//	scoutKeys.Accepted[i].Plugins))
			a := color.CyanString(scoutKeys.Accepted[i].Hostname)
			b := color.GreenString(fmt.Sprintf("[Version: %s] [OS: %s] [IP: %s] [Plugins: %s]",
				scoutKeys.Accepted[i].ScoutVersion,
				scoutKeys.Accepted[i].Os, scoutKeys.Accepted[i].Ip,
				scoutKeys.Accepted[i].Plugins))
			fmt.Fprintf(color.Output, "%s ", a)
			fmt.Fprintln(color.Output, b)
		}
		color.HiRed("[Unaccepted Scouts (%d)]", len(scoutKeys.Unaccepted))
		for i := range scoutKeys.Unaccepted {
			color.Red(scoutKeys.Unaccepted[i])
		}
		color.HiBlue("[Denied Scouts (%d)]", len(scoutKeys.Denied))
		for i := range scoutKeys.Denied {
			color.Blue(scoutKeys.Denied[i])
		}
	}
	return nil
}

type OpeResponse struct {
	Error  string `json:"error"`
	Result string `json:"result"`
}

func scoutOperateResponse(res *http.Response) {
	body, _ := ioutil.ReadAll(res.Body)
	OpeResponse := OpeResponse{}
	err := json.Unmarshal(body, &OpeResponse)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println(string(body))
	} else {
		if res.StatusCode < 400 {
			fmt.Println(OpeResponse.Result)
		} else {
			fmt.Println("error:", OpeResponse.Error)
		}
	}
}

func scoutAccept() error {
	acceptUrl := utils.Config().General.Addresses + "/host/accept"
	acceptAllUrl := utils.Config().General.Addresses + "/host/accept/all"
	var req *http.Request
	if ScoutAccept != "*" {
		data := make(map[string]interface{})
		data["hostname"] = ScoutAccept
		bytesData, _ := json.Marshal(data)
		payload := bytes.NewReader(bytesData)
		req, _ = http.NewRequest("POST", acceptUrl, payload)
	} else {
		req, _ = http.NewRequest("POST", acceptAllUrl, nil)
	}
	req = utils.HttpHandler(req)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error: Can not connection general.")
		os.Exit(1)
	}
	defer res.Body.Close()
	scoutOperateResponse(res)
	return nil
}

func scoutReject() error {
	rejectUrl := utils.Config().General.Addresses + "/host/reject"
	rejectAllUrl := utils.Config().General.Addresses + "/host/reject/all"
	var req *http.Request
	if ScoutReject != "*" {
		data := make(map[string]interface{})
		data["hostname"] = ScoutReject
		bytesData, _ := json.Marshal(data)
		payload := bytes.NewReader(bytesData)
		req, _ = http.NewRequest("POST", rejectUrl, payload)
	} else {
		req, _ = http.NewRequest("POST", rejectAllUrl, nil)
	}
	req = utils.HttpHandler(req)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error: Can not connection general.")
		os.Exit(1)
	}
	defer res.Body.Close()
	scoutOperateResponse(res)
	return nil
}

func scoutDelete() error {
	deleteUrl := utils.Config().General.Addresses + "/host/delete"
	deleteAllUrl := utils.Config().General.Addresses + "/host/delete/all"
	var req *http.Request
	if ScoutDelete != "*" {
		data := make(map[string]interface{})
		data["hostname"] = ScoutDelete
		bytesData, _ := json.Marshal(data)
		payload := bytes.NewReader(bytesData)
		req, _ = http.NewRequest("POST", deleteUrl, payload)
	} else {
		req, _ = http.NewRequest("POST", deleteAllUrl, nil)
	}
	req = utils.HttpHandler(req)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error: Can not connection general.")
		os.Exit(1)
	}
	defer res.Body.Close()
	scoutOperateResponse(res)
	return nil
}
