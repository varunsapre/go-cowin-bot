package cowinapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const (
	DateLayout = "02-01-2006"
)

var (
	ist *time.Location
)

func init() {
	ist, _ = time.LoadLocation("Asia/Kolkata")
	log.Println("Start Time: ", time.Now().In(ist))
}

type Centers struct {
	Sessions []Session `json:"sessions"`
}

type Session struct {
	CenterID               int      `json:"center_id"`
	Name                   string   `json:"name"`
	StateName              string   `json:"state_name"`
	DistrictName           string   `json:"district_name"`
	BlockName              string   `json:"block_name"`
	Pincode                int      `json:"pincode"`
	Lat                    float64  `json:"lat"`
	Lonf                   float64  `json:"long"`
	TimeFrom               string   `json:"from"`
	TimeTo                 string   `json:"to"`
	FeeType                string   `json:"fee_type"`
	SessionID              string   `json:"session_id"`
	Date                   string   `json:"date"`
	AvailableCapacity      float64  `json:"available_capacity"`
	AvailableCapacityDose1 float64  `json:"available_capacity_dose1"`
	AvailableCapacityDose2 float64  `json:"available_capacity_dose2"`
	MinAge                 int      `json:"min_age_limit"`
	VaccineName            string   `json:"vaccine"`
	Slots                  []string `json:"slots"`
}

type OutputInfo struct {
	CenterName             string   `json:"center_name"`
	Pincode                int      `json:"pincode"`
	FeeType                string   `json:"fee"`
	AvailableCapacity      float64  `json:"available_capacity"`
	AvailableCapacityDose1 float64  `json:"available_capacity_dose1"`
	AvailableCapacityDose2 float64  `json:"available_capacity_dose2"`
	MinAge                 int      `json:"min_age"`
	VaccineName            string   `json:"vaccine"`
	Slots                  []string `json:"slots"`
	Date                   string   `json:"date"`
}

type Options struct {
	DistrictID  string
	VaccineName string

	Age               int
	AvailableCapacity int
	PollTimer         int
	Days              int
	DoseNum           int

	RunAtNight bool
}

func StartCMDOnly(op *Options) {
	for {
		time.Sleep(time.Duration(op.PollTimer) * time.Second)

		output, err := GetBulkAvailability(op)
		if err != nil {
			log.Println("ERROR: ", err)
			continue
		}

		for _, o := range output {
			log.Printf("\t\t%v - %v | capacity: %v | Date: %v", o.CenterName, o.Pincode, o.AvailableCapacity, o.Date)
		}
	}
}

func GetBulkAvailability(op *Options) ([]OutputInfo, error) {
	today := time.Now().In(ist)
	weekAvailability := []OutputInfo{}
	numDays := op.Days

	// return if bot shouldn't run at night
	if !op.RunAtNight {
		if today.Hour() >= 0 && today.Hour() <= 6 {
			return nil, nil
		}
	}

	// stop checking for today if time has crossed 5pm
	if today.Hour() >= 17 {
		today = today.AddDate(0, 0, 1)
	}

	log.Printf("polling for: %v + %v day(s)", today.Format(DateLayout), numDays-1)

	for i := 0; i < numDays; i++ {
		d := today.AddDate(0, 0, i).Format(DateLayout)

		output, err := HitURL(op, d)
		if err != nil {
			msg := fmt.Errorf("error for date '%v': %v", d, err)
			return nil, msg
		}

		weekAvailability = append(weekAvailability, output...)
	}

	return weekAvailability, nil
}

func HitURL(op *Options, date string) ([]OutputInfo, error) {
	if date == "" {
		date = time.Now().Format(DateLayout)
	}

	url := fmt.Sprintf("https://cdn-api.co-vin.in/api/v2/appointment/sessions/public/findByDistrict?district_id=%v&date=%v", op.DistrictID, date)

	centers := Centers{}
	availabilites := []OutputInfo{}

	log.Println(url)

	client := http.Client{
		Timeout: 5 * time.Second,
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

	for _, s := range centers.Sessions {
		if filterConditions(s, op) {
			availabilites = append(availabilites, createOutput(s))
		}
	}

	log.Printf("\tDisctrictID : %v | MinimumAge: %v | Dose: %v | VaccineName : %v | Centers Available: %v", op.DistrictID, op.Age, op.DoseNum, op.VaccineName, len(availabilites))

	if len(availabilites) == 0 {
		return nil, nil
	}

	return availabilites, nil
}

func filterConditions(s Session, op *Options) bool {
	if s.MinAge != op.Age {
		return false
	}

	if op.VaccineName != "" && s.VaccineName != op.VaccineName {
		return false
	}

	// by default check against dose 1 capacity
	switch op.DoseNum {
	case 2:
		if s.AvailableCapacityDose2 <= float64(op.AvailableCapacity) {
			return false
		}
	default:
		if s.AvailableCapacityDose1 <= float64(op.AvailableCapacity) {
			return false
		}
	}

	return true
}

func createOutput(s Session) OutputInfo {
	return OutputInfo{
		CenterName:             s.Name,
		Pincode:                s.Pincode,
		FeeType:                s.FeeType,
		AvailableCapacity:      s.AvailableCapacity,
		AvailableCapacityDose1: s.AvailableCapacityDose1,
		AvailableCapacityDose2: s.AvailableCapacityDose2,
		MinAge:                 s.MinAge,
		VaccineName:            s.VaccineName,
		Slots:                  s.Slots,
		Date:                   s.Date,
	}
}
