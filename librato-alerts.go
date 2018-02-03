package main

import (
	"fmt"
	"bufio"
	"github.com/joho/godotenv"
	"gopkg.in/resty.v1"
	"log"
	"os"
	"encoding/json"
	"github.com/fatih/color"
	"strconv"
	"strings"
)

type StatusResponse struct {
	Firing []AlertEvent `json:"firing"`
	Cleared []AlertEvent `json:"cleared"`
}

type AlertEvent struct {
	ID          int `json:"id"`
	TriggeredAt int `json:"triggered_at"`
}

type Alert struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Conditions  []struct {
		ID              int     `json:"id"`
		Type            string  `json:"type"`
		MetricName      string  `json:"metric_name"`
		Source          string  `json:"source"`
		Threshold       float64 `json:"threshold"`
		Duration        int     `json:"duration"`
		SummaryFunction string  `json:"summary_function"`
	} `json:"conditions"`
	Services []struct {
		ID       int    `json:"id"`
		Type     string `json:"type"`
		Settings struct {
			URL string `json:"url"`
		} `json:"settings"`
		Title string `json:"title"`
	} `json:"services"`
	Attributes struct {
	} `json:"attributes"`
	Active         bool `json:"active"`
	CreatedAt      int  `json:"created_at"`
	UpdatedAt      int  `json:"updated_at"`
	Version        int  `json:"version"`
	RearmSeconds   int  `json:"rearm_seconds"`
	RearmPerSignal bool `json:"rearm_per_signal"`
	Md             bool `json:"md"`
}

type AlertListResponse struct {
	Query	string 	`json:"query"`
	Alerts	[]Alert	`json:"alerts"`
}

//TODO: firing and recent can be only one func parametrized
func printFiring(){
	resp, err := resty.R().Get("https://metrics-api.librato.com/v1/alerts/status")
	if err != nil {
		log.Fatal("Error getting alert status > ", err)
	}
	var jsonRes StatusResponse
	json.Unmarshal([]byte(resp.String()), &jsonRes)

	if len(jsonRes.Firing) > 0 {
		fmt.Println("Alerts firing:")
		for _, alert := range jsonRes.Firing {
			resp, err := resty.R().Get("https://metrics-api.librato.com/v1/alerts/" + strconv.Itoa(alert.ID))
			if err != nil {
				log.Fatal("Error getting alert id > ", err)
			}
			var jsonAlert Alert
			json.Unmarshal([]byte(resp.String()), &jsonAlert)
			fmt.Println(jsonAlert.Name)
		}
	} else {
		fmt.Println("There are no alerts firing at this moment")
	}
}

func printRecent() {
	resp, err := resty.R().Get("https://metrics-api.librato.com/v1/alerts/status")
	if err != nil {
		log.Fatal("Error getting alert status > ", err)
	}
	var jsonRes StatusResponse
	json.Unmarshal([]byte(resp.String()), &jsonRes)

	if len(jsonRes.Cleared) > 0 {
		fmt.Println("Alerts recently cleared:")
		for _, alert := range jsonRes.Cleared {
			resp, err := resty.R().Get("https://metrics-api.librato.com/v1/alerts/" + strconv.Itoa(alert.ID))
			if err != nil {
				log.Fatal("Error getting alert id > ", err)
			}
			var jsonAlert Alert
			json.Unmarshal([]byte(resp.String()), &jsonAlert)
			fmt.Println(jsonAlert.Name)
		}
	} else {
		fmt.Println("There are no alerts recently cleared at this moment")
	}
}

func alertsEnable() {
	resp, err := resty.R().Get("https://metrics-api.librato.com/v1/alerts")
	if err != nil {
		log.Fatal("Error getting alert list > ", err)
	}

	var jsonRes AlertListResponse
	json.Unmarshal([]byte(resp.String()), &jsonRes)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := string(scanner.Text())
		alertName := line
		if strings.Contains(line,string(':')) {
			arr := strings.Split(line,string(':'))
			alertName = arr[0]
		}
		for _,alert := range jsonRes.Alerts {
			if alert.Name == alertName {
				if alert.Active {
					fmt.Println("alert " + alertName + " already enabled")
				} else {
					fmt.Println("enabling alert " + alertName)
					alert.Active = true
					_, updateErr := resty.R().SetBody(alert).Put("https://metrics-api.librato.com/v1/alerts/" + strconv.Itoa(alert.ID))
					if updateErr != nil {
						log.Fatal("Error updating alert " + alert.Name )
					}
					fmt.Println(alert.Name + " enabled")
				}
			}
		}
	}
}

func alertsDisable() {
	resp, err := resty.R().Get("https://metrics-api.librato.com/v1/alerts")
	if err != nil {
		log.Fatal("Error getting alert list > ", err)
	}

	var jsonRes AlertListResponse
	json.Unmarshal([]byte(resp.String()), &jsonRes)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := string(scanner.Text())
		alertName := line
		if strings.Contains(line,string(':')) {
			arr := strings.Split(line,string(':'))
			alertName = arr[0]
		}
		for _,alert := range jsonRes.Alerts {
			if alert.Name == alertName {
				if alert.Active {
					fmt.Println("disabling alert " + alert.Name)
					alert.Active = false
					_, updateErr := resty.R().SetBody(alert).Put("https://metrics-api.librato.com/v1/alerts/" + strconv.Itoa(alert.ID))
					if updateErr != nil {
						log.Fatal("Error updating alert " + alert.Name )
					}
					fmt.Println(alert.Name + " disabled")
				} else {
					fmt.Println("alert " + alertName + " already disabled")
				}
			}
		}
	}
}


func printAlerts() {

	resp, err := resty.R().Get("https://metrics-api.librato.com/v1/alerts")
	if err != nil {
		log.Fatal("Error getting alert list > ", err)
	}

	var jsonRes AlertListResponse
	json.Unmarshal([]byte(resp.String()), &jsonRes)

	for _, alert := range jsonRes.Alerts {
		fmt.Print(color.HiYellowString(alert.Name), ": ")
		if alert.Active {
			color.HiGreen("Active")
		} else {
			color.HiRed("Disabled")
		}
	}
}

func main() {
	// load dotenv
	err := godotenv.Load()
	if err != nil {
		log.Println("No extra vars loaded >", err)
	}
	// resty configuration
	resty.SetDebug(false)
	resty.SetBasicAuth(os.Getenv("MAIL"), os.Getenv("TOKEN"))
	// check arg 0
	mode := "list"
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}

	switch mode {
	case "list", "enable", "disable", "help", "status", "recent":
	default:
		mode = "help"
	}

	// check stdin
	fi, err := os.Stdin.Stat()
	if err != nil {
		log.Fatal("Unable to read stdin >", err)
	}
	if (fi.Mode() & os.ModeCharDevice) == 0 {
		switch mode {
		case "list", "status", "recent":
			log.Fatal(mode, " mode can't be called with piped data, please use enable or disable mode")
		case "enable":
			alertsEnable()
		case "disable":
			alertsDisable()
		default:
			log.Fatal("unknown mode ", mode)
		}
	} else {
		switch mode {
		case "enable", "disable":
			log.Fatal(mode +" mode requires a list of alerts piped into comand")
		case "list":
			printAlerts()
		case "help":
			//TODO: implement help
			log.Fatal("Unimplemented mode ", mode)
		case "status":
			printFiring()
		case "recent":
			printRecent()
		default:
			log.Fatal("unknown mode ", mode)
		}
	}
}
