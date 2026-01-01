package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ip-api response structure
type IpAPI struct {
	Query   string `json:"query"`
	ISP     string `json:"isp"`
	Country string `json:"country"`
	Status  string `json:"status"`
}

// ipmyp response structure
type IpMyP struct {
	Query string `json:"query"`
	Ip    string `json:"ip"`
}

// fetchJSON fetches JSON from a URL and parses it
// IMPORTANT: it parses the body EVEN if HTTP status != 200
func fetchJSON(url string, target any) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, target)
}

// fetchIPAPI fetches enriched IP info from ip-api
func fetchIPAPI(ip string) (*IpAPI, error) {
	url := "http://ip-api.com/json/" + ip
	var info IpAPI
	if err := fetchJSON(url, &info); err != nil {
		return nil, err
	}
	if info.Status != "success" {
		return nil, fmt.Errorf("ip-api lookup failed")
	}
	return &info, nil
}

func main() {
	// 1) Primary IP info (current IP)
	var primary IpAPI
	if err := fetchJSON("http://ip-api.com/json/", &primary); err != nil {
		fmt.Println("ip-api unreachable")
		return
	}

	// 2) Secondary IP (ipmyp — may return JSON or plain text)
	var ipmyp IpMyP
	resp, err := http.Get("https://api.ipmyp.ir/")
	if err != nil {
		fmt.Println("ipmyp lookup failed:", err)
	} else {
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("ipmyp read failed:", err)
		} else {
			// try to parse JSON first
			if err := json.Unmarshal(body, &ipmyp); err != nil {
				// fallback: treat response as plain-text IP
				ip := strings.TrimSpace(string(body))
				if ip == "" {
					fmt.Println("ipmyp: empty response")
				} else {
					ipmyp.Query = ip
				}
			}
		}
	}

	// 3) Enrich ipmyp IP if available (accept either `query` or `ip`)
	var ipmypInfo *IpAPI
	ip := ipmyp.Query
	if ip == "" {
		ip = ipmyp.Ip
	}
	if ip != "" {
		if info, err := fetchIPAPI(ip); err != nil {
			fmt.Println("ip-api lookup for ipmyp failed:", err)
		} else {
			ipmypInfo = info
		}
	}

	// 4) Output
	fmt.Println("================ PUBLIC IP INFO ================")
	fmt.Printf("IP (ip-api)   : %s\n", primary.Query)
	fmt.Printf("ISP           : %s\n", primary.ISP)
	fmt.Printf("Country       : %s\n", primary.Country)

	// show ipmyp raw result if available
	ipDisplay := ipmyp.Query
	if ipDisplay == "" {
		ipDisplay = ipmyp.Ip
	}
	if ipDisplay != "" {
		fmt.Printf("IP (ipmyp)    : %s\n", ipDisplay)
	}

	if ipmypInfo != nil {
		fmt.Println("--------------- ipmyp enriched ----------------")
		fmt.Printf("IP            : %s\n", ipmypInfo.Query)
		fmt.Printf("ISP           : %s\n", ipmypInfo.ISP)
		fmt.Printf("Country       : %s\n", ipmypInfo.Country)
	}

	fmt.Println("================================================")
}
