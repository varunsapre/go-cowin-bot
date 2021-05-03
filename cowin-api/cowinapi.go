package cowinapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

const (
	dateLayout = "02-01-2006"
)

// {
// 	"center_id": 249096,
// 	"name": "Sulikere PHC",
// 	"state_name": "Karnataka",
// 	"district_name": "Bangalore Urban",
// 	"block_name": "Bengaluru South",
// 	"pincode": 560060,
// 	"from": "10:00:00",
// 	"to": "16:00:00",
// 	"lat": 12,
// 	"long": 77,
// 	"fee_type": "Free",
// 	"session_id": "c4860857-26d9-4cd1-a08a-f0042ee2866c",
// 	"date": "03-05-2021",
// 	"available_capacity": 1,
// 	"fee": "0",
// 	"min_age_limit": 45,
// 	"vaccine": "COVISHIELD",
// 	"slots": [
// 	  "10:00AM-11:00AM",
// 	  "11:00AM-12:00PM",
// 	  "12:00PM-01:00PM",
// 	  "01:00PM-04:00PM"
// 	]
//   },

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
	AvailableCapacity int      `json:"available_capacity"`
	MinAge            int      `json:"min_age_limit"`
	VaccineName       string   `json:"vaccine"`
	Slots             []string `json:"slots"`
}

type OutputInfo struct {
	CenterName        string   `json:"center_name"`
	Pincode           int      `json:"pincode"`
	FeeType           string   `json:"fee"`
	AvailableCapacity int      `json:"available_capacity"`
	MinAge            int      `json:"min_age"`
	VaccineName       string   `json:"vaccine"`
	Slots             []string `json:"slots"`
	Date              string   `json:"date"`
}

func Serve() {
	r := mux.NewRouter()
	r.HandleFunc("/{district_id}/{age}", getAvailabilites)
	http.Handle("/", r)

	srv := &http.Server{
		Handler: r,
		Addr:    "127.0.0.1:8000",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

func getAvailabilites(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	district_id, ok := vars["district_id"]
	if !ok {
		// default districit to 265 (bengaluru urban)
		district_id = "265"
	}

	varAge, ok := vars["age"]
	if !ok {
		// default minimun age to 18
		varAge = "18"
	}

	output, err := HitURL(district_id, varAge, "") //empty date to default to today
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	if output == nil {
		msg := fmt.Sprintf("disctrictID : %v | minimumAge: %v | NO VACANCIES", district_id, varAge)
		w.Write([]byte(msg))
		return
	}

	strOutput, err := json.MarshalIndent(output, " ", " ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Write(strOutput)
}

func GetWeekAvailability(district_id, age string) ([]OutputInfo, error) {
	today := time.Now()
	weekAvailability := []OutputInfo{}

	log.Println("fetching 1 week availabilites")

	for i := 0; i < 7; i++ {
		d := today.AddDate(0, 0, i).Format(dateLayout)

		output, err := HitURL(district_id, age, d)
		if err != nil {
			log.Printf("Error for date '%v': %v", d, err)
		}

		weekAvailability = append(weekAvailability, output...)
	}

	return weekAvailability, nil
}

func HitURL(district_id, age, date string) ([]OutputInfo, error) {
	if date == "" {
		date = time.Now().Format(dateLayout)
	}

	url := fmt.Sprintf("https://cdn-api.co-vin.in/api/v2/appointment/sessions/public/findByDistrict?district_id=%v&date=%v", district_id, date)

	centers := Centers{}
	availabilites := []OutputInfo{}

	log.Println(url)

	resp, err := http.Get(url)
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

	if s.AvailableCapacity <= 0 {
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
