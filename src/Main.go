package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/hugolgst/rich-go/client"
	"github.com/shirou/gopsutil/process"
)

var (
	placeId string
	reset = false
)

type MarketPlaceInfo struct { // https://mholt.github.io/json-to-go/
	Name        string      `json:"Name"`
	Description string      `json:"Description"`
	Creator     struct {
		ID              int    `json:"Id"`
		Name            string `json:"Name"`
		CreatorType     string `json:"CreatorType"`
		CreatorTargetID int    `json:"CreatorTargetId"`
	} `json:"Creator"`
	IconImageAssetID       int64       `json:"IconImageAssetId"`
}



type ThumbnailInfo struct { // https://mholt.github.io/json-to-go/
		Data []struct {
			TargetID int64  `json:"targetId"`
			State    string `json:"state"`
			ImageURL string `json:"imageUrl"`
		} `json:"data"`
}

type Universe struct { // https://mholt.github.io/json-to-go/
	UniverseID int64 `json:"UniverseId"`
}

func GetProcessByName(targetProcessName string) *process.Process {
	processes, _ := process.Processes()

	for _, proc := range processes {
		name, _ := proc.Name()
		cmdLine, _ := proc.Cmdline()
		
		if (name == targetProcessName && strings.Contains(cmdLine, "--play")) {
			return proc
		}
	}

	return nil
}

func GetPlaceInfoByPlaceId(placeId string) *MarketPlaceInfo {
	url := "https://api.roblox.com/marketplace/productinfo?assetId=" + placeId
	resp, _ := http.Get(url)

	defer resp.Body.Close()

	var info *MarketPlaceInfo

	json.NewDecoder(resp.Body).Decode(&info)

	return info
}

func GetUniverseIdByPlaceId(placeId string) *Universe {
	url := "https://api.roblox.com/universes/get-universe-containing-place?placeid=" + placeId
	resp, _ := http.Get(url)

	defer resp.Body.Close()

	var info *Universe

	json.NewDecoder(resp.Body).Decode(&info)

	return info
}


func GetIconByUniverseId(UniverseID string) *ThumbnailInfo {
	url := "https://thumbnails.roblox.com/v1/games/icons?universeIds=" + UniverseID + "&size=512x512&format=Png&isCircular=false"
	resp, _ := http.Get(url)

	defer resp.Body.Close()

	var info *ThumbnailInfo

	json.NewDecoder(resp.Body).Decode(&info)

	return info
}

func UpdateRobloxPresence() {
	roblox := GetProcessByName("RobloxPlayerBeta.exe")

	for (roblox == nil) {
		roblox = GetProcessByName("RobloxPlayerBeta.exe")

		if (reset == false) {
			reset = true
			placeId = ""

			client.Logout()
			fmt.Println("reset client activity")
		}
	}

	err := client.Login("823294557155754005")

	if (err != nil) {
		fmt.Println(err)
	}

	reset = false

	args, _ := roblox.Cmdline()

	placePattern := regexp.MustCompile(`placeId=(\d+)`)
	placeMatch := placePattern.FindStringSubmatch(args)[1]

	// timePattern := regexp.MustCompile(`launchtime=(\d+)`)
	// timeMatch := timePattern.FindStringSubmatch(args)[1]

	// startTime, _ := strconv.ParseInt(timeMatch, 10, 64)

	now := time.Now()

	if (placeMatch != placeId) {
		placeId = placeMatch
		place := GetPlaceInfoByPlaceId(placeId)
		universeId := GetUniverseIdByPlaceId(placeId).UniverseID
		thumbnail := GetIconByUniverseId(strconv.FormatInt(universeId, 10)).Data[0].ImageURL

		client.SetActivity(client.Activity {
			State: "by " + place.Creator.Name,
			Details: place.Name,
			LargeImage: thumbnail,
			LargeText: "Playing Roblox!",
			Buttons: []*client.Button {
				&client.Button {
					Label: "Open Game Page",
					Url: "https://www.roblox.com/games/" + placeId + "/-",
				},
			},
			Timestamps: &client.Timestamps {
				Start: &now,
			},
		})

		fmt.Println("set activity: " + place.Name)
		fmt.Println("by: " + place.Creator.Name)
	}
}

func main() {
	for (true) {
		UpdateRobloxPresence()

		time.Sleep(time.Second * 5)
	}
}
