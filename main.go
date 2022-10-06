package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"github.com/mitchellh/go-homedir"
	"gopkg.in/resty.v1"
)

type statusResponse struct {
	Firing  []alertEvent `json:"firing"`
	Cleared []alertEvent `json:"cleared"`
}

type alertEvent struct {
	ID          int `json:"id"`
	TriggeredAt int `json:"triggered_at"`
}

type libratoAlert struct {
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

type alertList []libratoAlert

type queryMeta struct {
	Offset int `json:"offset"`
	Length int `json:"length"`
	Found  int `json:"found"`
	Total  int `json:"total"`
}

type alertListResponse struct {
	Query  queryMeta      `json:"query"`
	Alerts []libratoAlert `json:"alerts"`
}

//TODO: firing and recent can be only one func parametrized

func getAllAlertList() (error, *alertList) {
	offset := 0
	length := 0
	total := 1000

	var alerts alertList

	for offset+length < total {
		offset = offset + length
		resp, err := resty.R().Get("https://metrics-api.librato.com/v1/alerts?offset=" + strconv.Itoa(offset))
		if err != nil {
			return err, nil
		}

		var jsonRes alertListResponse
		err = json.Unmarshal([]byte(resp.String()), &jsonRes)
		if err != nil {
			return err, nil
		}
		length = jsonRes.Query.Length
		offset = jsonRes.Query.Offset
		total = jsonRes.Query.Total
		alerts = append(alerts, jsonRes.Alerts...)
	}

	return nil, &alerts
}

func printAlerts() {
	err, alerts := getAllAlertList()
	if err != nil {
		log.Fatal("Eror getting alert list ", err)
	}
	for _, alert := range *alerts {
		fmt.Print(color.HiYellowString(alert.Name), ": ")
		if alert.Active {
			color.HiGreen("Active")
		} else {
			color.HiRed("Disabled")
		}
	}
}

func getStatus() (error, *statusResponse) {
	resp, err := resty.R().Get("https://metrics-api.librato.com/v1/alerts/status")
	if err != nil {
		return err, nil
	}
	log.Printf("%v\n", resp.String())
	var jsonRes statusResponse
	err = json.Unmarshal([]byte(resp.String()), &jsonRes)
	if err != nil {
		return err, nil
	}
	return nil, &jsonRes
}

func printFiring() {
	err, jsonRes := getStatus()
	if err != nil {
		log.Fatal("Error getting firing status: ", err)
	}

	if len(jsonRes.Firing) > 0 {
		fmt.Println("Alerts firing:")
		for _, alert := range jsonRes.Firing {
			resp, err := resty.R().Get("https://metrics-api.librato.com/v1/alerts/" + strconv.Itoa(alert.ID))
			if err != nil {
				log.Fatal("Error getting alert id > ", err)
			}
			var jsonAlert libratoAlert
			err = json.Unmarshal([]byte(resp.String()), &jsonAlert)
			if err != nil {
				log.Fatal("Error unmarshaling Firing JSON: ", err)
			}
			fmt.Println(jsonAlert.Name)
		}
	} else {
		fmt.Println("There are no alerts firing at this moment")
	}
}

func printRecent() {
	err, jsonRes := getStatus()
	if err != nil {
		log.Fatal("Error getting recent status: ", err)
	}

	if len(jsonRes.Cleared) > 0 {
		fmt.Println("Alerts recently cleared:")
		for _, alert := range jsonRes.Cleared {
			resp, err := resty.R().Get("https://metrics-api.librato.com/v1/alerts/" + strconv.Itoa(alert.ID))
			if err != nil {
				log.Fatal("Error getting alert id > ", err)
			}
			var jsonAlert libratoAlert
			err = json.Unmarshal([]byte(resp.String()), &jsonAlert)
			if err != nil {
				log.Fatal("Error unmarshaling Recent JSON: ", err)
			}
			fmt.Println(jsonAlert.Name)
		}
	} else {
		fmt.Println("There are no alerts recently cleared at this moment")
	}
}

func alertsEnable() {
	err, alerts := getAllAlertList()
	if err != nil {
		log.Fatal("Eror getting alert list ", err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		alertName := line
		if strings.Contains(line, string(':')) {
			arr := strings.Split(line, string(':'))
			alertName = arr[0]
		}
		for _, alert := range *alerts {
			if alert.Name == alertName {
				if alert.Active {
					fmt.Println("alert " + alertName + " already enabled")
				} else {
					fmt.Println("enabling alert " + alertName)
					alert.Active = true
					if alert.Description == "" {
						alert.Description = "-"
					}
					result, updateErr := resty.R().
						SetBody(alert).
						Put("https://metrics-api.librato.com/v1/alerts/" + strconv.Itoa(alert.ID))
					if updateErr != nil {
						log.Fatal("Error updating alert " + alert.Name)
					}
					if result.IsError() {
						log.Fatalf("Error updating alter %v: Return code (%v), Return body %v", alert.Name, result.StatusCode(), string(result.Body()))
					}
					fmt.Println(alert.Name + " enabled")
				}
			}
		}
	}
}

func alertsDisable() {
	err, alerts := getAllAlertList()
	if err != nil {
		log.Fatal("Eror getting alert list ", err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		alertName := line
		if strings.Contains(line, string(':')) {
			arr := strings.Split(line, string(':'))
			alertName = arr[0]
		}
		for _, alert := range *alerts {
			if alert.Name == alertName {
				if alert.Active {
					fmt.Println("disabling alert " + alert.Name)
					alert.Active = false
					if alert.Description == "" {
						alert.Description = "-"
					}
					result, updateErr := resty.R().
						SetBody(alert).
						Put("https://metrics-api.librato.com/v1/alerts/" + strconv.Itoa(alert.ID))
					if updateErr != nil {
						log.Fatal("Error updating alert " + alert.Name)
					}
					if result.IsError() {
						log.Fatalf("Error updating alter %v: Return code (%v), Return body %v", alert.Name, result.StatusCode(), string(result.Body()))
					}
					fmt.Println(alert.Name + " disabled")
				} else {
					fmt.Println("alert " + alertName + " already disabled")
				}
			}
		}
	}
}

func printAlertsStatus() {
	err, alerts := getAllAlertList()
	if err != nil {
		log.Fatal("Eror getting alert list ", err)
	}

	err, statusRes := getStatus()
	if err != nil {
		log.Fatal("Error getting status: ", err)
	}

	for _, alert := range *alerts {
		fmt.Print(color.HiYellowString(alert.Name), ": ")
		if alert.Active {
			status := color.HiGreenString("Active")
			for _, alertStatus := range statusRes.Cleared {
				if alert.ID == alertStatus.ID {
					status = color.GreenString("Recent, Active")
					break
				}
			}
			for _, alertStatus := range statusRes.Firing {
				if alert.ID == alertStatus.ID {
					status = color.HiRedString("Firing, Active")
					break
				}
			}
			fmt.Println(status)
		} else {
			color.Red("Disabled")
		}
	}
}

func printHelp() {

	fmt.Println(`# librato-alerts-cli

Small commandline client to enable and disable alerts in librato legacy
accounts.

Usage: ` + "`" + ` librato-alerts-cli [help | disable | enable | list | status | recent]` + "`" + `

` + "`" + `enable` + "`" + ` and ` + "`" + `disable` + "`" + ` requires a list of alerts to disable passed by standard
input thru a pipe, the output of ` + "`" + `list` + "`" + ` can be used for this purpose like this:
` + "```" + `
   librato-alerts-cli list | grep <pattern> | librato-alerts-cli disable
` + "```" + `

## CONFIGURATION

This requires two environment varables to store the librato credentials,
` + "`" + `LIBRATO_MAIL` + "`" + ` with the librato user's mail and ` + "`" + `LIBRATO_TOKEN` + "`" + `
with a valid librato API token. API token must have read / write access to allow update alarms state.
The environment variables can also be placed in an ` + "`" + `.env` + "`" + ` file or in a
` + "`" + `.librato-alerts-cli` + "`" + ` file in home directory. You can use ` + "`" + `librato-alerts-cli config` + "`" + `
to generate that file.

## MODES

` + "```" + `
   list:       List all alerts, telling if they are enabled or disabled.
   statuslist: List all alerts, telling if they are enabled or disabled and its status Firing / Recent.
   status:     Lists the alert names which are in alarm state.
   recent:     Lists the alert names of alert which were resolved recently.
   enable:     Enable alerts passed by stdin. Alerts must be pased one by line,
               and it will be updated only if they are disabled
   disable:    Disable alerts passed by stdin. Alerts must be pased one by line,
               and it will be updated only if they are enabled
   config:     Prints current config in a valid format to be a proper config file.
   help:       This help.
` + "```" + `

## ALMOST KNOWN BUGS or TODO's:

 * This is tested against an old, no tagged metrics librato account may work
   in the modern ones.`)
}

func checkEnv() bool {
	envNeeded := []string{"LIBRATO_MAIL", "LIBRATO_TOKEN"}
	checkEnv := true
	for _, envVar := range envNeeded {
		_, present := os.LookupEnv(envVar)
		if !present {
			checkEnv = false
			log.Println("Missing needed environment variable ", envVar)
		}
	}
	return checkEnv
}

func printConfig() {
	userConfigFile, _ := homedir.Expand("~/.librato-alerts-cli")
	fmt.Printf("# place and fill if needed these lines in a local file called .env\n")
	fmt.Printf("# or in your home dir as %v\n", userConfigFile)
	fmt.Printf("# or find a way to set it as environment variables\n")
	fmt.Printf("# local .env takes precedence over home file, any of these will override already setted environment variables\n\n")
	fmt.Printf("LIBRATO_MAIL=%v\n", os.Getenv("LIBRATO_MAIL"))
	fmt.Printf("LIBRATO_TOKEN=%v\n", os.Getenv("LIBRATO_TOKEN"))
}

func main() {
	// load dotenv
	godotenv.Load()
	userConfigFile, _ := homedir.Expand("~/.librato-alerts-cli")
	godotenv.Load(userConfigFile)

	// check arg 0
	mode := "list"
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}
	if mode != "config" && mode != "help" && !checkEnv() {
		log.Fatal("Insufficient configuration. Please run librato-alerts-cli config and follow instructions")
	}
	// resty configuration
	resty.SetDebug(false)
	resty.SetBasicAuth(os.Getenv("LIBRATO_MAIL"), os.Getenv("LIBRATO_TOKEN"))
	// check stdin
	fi, err := os.Stdin.Stat()
	if err != nil {
		log.Fatal("Unable to read stdin >", err)
	}

	piped := (fi.Mode() & os.ModeCharDevice) == 0
	if piped && (mode == "list" || mode == "statuslist" || mode == "status" || mode == "recent") {
		log.Fatal(mode, " mode can't be called with piped data, please use enable or disable mode")
	}
	if !piped && (mode == "enabled" || mode == "disable") {
		log.Fatal(mode + " mode requires a list of alerts piped into comand")
	}

	switch mode {
	case "list":
		printAlerts()
	case "statuslist":
		printAlertsStatus()
	case "enable":
		alertsEnable()
	case "disable":
		alertsDisable()
	case "recent":
		printRecent()
	case "status":
		printFiring()
	case "config":
		printConfig()
	default:
		printHelp()
	}
}
