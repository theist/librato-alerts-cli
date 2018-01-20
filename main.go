package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"gopkg.in/resty.v1"
	"log"
	"os"
	"encoding/json"
	"github.com/fatih/color"
)

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

type Response struct {
	Query	string 	`json:"query"`
	Alerts	[]Alert	`json:"alerts"`
}

func print_alerts(){

	resp, err := resty.R().Get("https://metrics-api.librato.com/v1/alerts")
	if err != nil {
		log.Fatal("Error getting alert list > ", err)
	}

	var json_res Response
	json.Unmarshal([]byte(resp.String()), &json_res)

	for _, alert := range json_res.Alerts {
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
	if fi.Size() > 0 {
		switch mode {
		case "list", "status", "recent":
			log.Fatal(mode, " mode can't be called with piped data, please use enable or disable mode")
		case "enable":
			//TODO: implement enable
			log.Fatal("Unimplemented mode ", mode)
		case "disable":
			//TODO: implement disable
			log.Fatal("Unimplemented mode ", mode)
		default:
			log.Fatal("unknown mode ", mode)
		}
	} else {
		switch mode {
		case "enable", "disable":
			log.Fatal("Enable or disable requires a list of alerts piped into comand")
		case "list":
			print_alerts()
		case "help":
			//TODO: implement help
			log.Fatal("Unimplemented mode ", mode)
		case "status":
			//TODO: implement status
			log.Fatal("Unimplemented mode ", mode)
		case "recent":
			//TODO: implement recent
			log.Fatal("Unimplemented mode ", mode)
		default:
			log.Fatal("unknown mode ", mode)
		}
	}
}
