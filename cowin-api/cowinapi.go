package cowinapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

const (
	DateLayout = "02-01-2006"
)

type Centers struct {
	Sessions []Session `json:"sessions"`
}

type Session struct {
	CenterID          int      `json:"center_id"`
	Name              string   `json:"name"`
	StateName         string   `json:"state_name"`
	DistrictName      string   `json:"district_name"`
	BlockName         string   `json:"block_name"`
	Pincode           int      `json:"pincode"`
	Lat               float64  `json:"lat"`
	Lonf              float64  `json:"long"`
	TimeFrom          string   `json:"from"`
	TimeTo            string   `json:"to"`
	FeeType           string   `json:"fee_type"`
	SessionID         string   `json:"session_id"`
	Date              string   `json:"date"`
	AvailableCapacity float64  `json:"available_capacity"`
	MinAge            int      `json:"min_age_limit"`
	VaccineName       string   `json:"vaccine"`
	Slots             []string `json:"slots"`
}

type OutputInfo struct {
	CenterName        string   `json:"center_name"`
	Pincode           int      `json:"pincode"`
	FeeType           string   `json:"fee"`
	AvailableCapacity float64  `json:"available_capacity"`
	MinAge            int      `json:"min_age"`
	VaccineName       string   `json:"vaccine"`
	Slots             []string `json:"slots"`
	Date              string   `json:"date"`
}

func StartCMDOnly(district_id, age string, pollTimer, days int) {
	for {
		time.Sleep(time.Duration(pollTimer) * time.Second)

		output, err := GetBulkAvailability(district_id, age, days)
		if err != nil {
			log.Println("ERROR: ", err)
			continue
		}

		for _, o := range output {
			log.Printf("\t\t%v - %v | capacity: %v | Date: %v", o.CenterName, o.Pincode, o.AvailableCapacity, o.Date)
		}
	}
}

func GetBulkAvailability(district_id, age string, days int) ([]OutputInfo, error) {
	today := time.Now()
	weekAvailability := []OutputInfo{}
	numDays := days

	// stop checking for today if time has crossed 5pm
	if today.Local().Hour() >= 17 {
		log.Println(" -- crossed 17:00hrs, not checking for today anymore -- ")
		today = today.AddDate(0, 0, 1)
	}

	log.Printf("polling for: %v + %v day(s)", today.Format(DateLayout), numDays-1)

	for i := 0; i < numDays; i++ {
		d := today.AddDate(0, 0, i).Format(DateLayout)

		output, err := HitURL(district_id, age, d)
		if err != nil {
			msg := fmt.Errorf("error for date '%v': %v", d, err)
			return nil, msg
		}

		weekAvailability = append(weekAvailability, output...)
	}

	return weekAvailability, nil
}

func HitURL(district_id, age, date string) ([]OutputInfo, error) {
	if date == "" {
		date = time.Now().Format(DateLayout)
	}

	url := fmt.Sprintf("https://cdn-api.co-vin.in/api/v2/appointment/sessions/public/findByDistrict?district_id=%v&date=%v", district_id, date)

	centers := Centers{}
	availabilites := []OutputInfo{}

	log.Println(url)

	client := http.Client{
		Timeout: 3 * time.Second,
	}

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &centers)
	if err != nil {
		return nil, fmt.Errorf("json Error: '%v' | body : '%v'", err, string(body))
	}

	minAge, err := strconv.Atoi(age)
	if err != nil {
		return nil, err
	}

	for _, s := range centers.Sessions {
		if filterConditions(s, minAge) {
			availabilites = append(availabilites, createOutput(s))
		}
	}

	log.Printf("\tDisctrictID : %v | MinimumAge: %v | Centers Available: %v", district_id, minAge, len(availabilites))

	if len(availabilites) == 0 {
		return nil, nil
	}

	return availabilites, nil
}

func filterConditions(s Session, minAge int) bool {
	if s.MinAge > minAge {
		return false
	}

	if s.AvailableCapacity <= 2 {
		return false
	}

	return true
}

func createOutput(s Session) OutputInfo {
	return OutputInfo{
		CenterName:        s.Name,
		Pincode:           s.Pincode,
		FeeType:           s.FeeType,
		AvailableCapacity: s.AvailableCapacity,
		MinAge:            s.MinAge,
		VaccineName:       s.VaccineName,
		Slots:             s.Slots,
		Date:              s.Date,
	}
}
