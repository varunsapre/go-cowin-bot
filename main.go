package main

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

type Centers struct {
	Centers []Center `json:"centers"`
}

type Center struct {
	ID           int       `json:"center_id"`
	Name         string    `json:"name"`
	StateName    string    `json:"state_name"`
	DistrictName string    `json:"district_name"`
	BlockName    string    `json:"block_name"`
	Pincode      int       `json:"pincode"`
	Lat          float64   `json:"lat"`
	Lonf         float64   `json:"long"`
	TimeFrom     string    `json:"from"`
	TimeTo       string    `json:"to"`
	FeeType      string    `json:"fee_type"`
	Sessions     []Session `json:"sessions"`
}

type Session struct {
	ID                string   `json:"session_id"`
	Date              string   `json:"date"`
	AvailableCapacity int      `json:"available_capacity"`
	MinAge            int      `json:"min_age_limit"`
	VaccineName       string   `json:"vaccine"`
	Slots             []string `json:"slots"`
}

type OutputInfo struct {
	CenterID          int      `json:"center_id"`
	CenterName        string   `json:"center_name"`
	Pincode           int      `json:"pincode"`
	FeeType           string   `json:"fee"`
	AvailableCapacity int      `json:"available_capacity"`
	MinAge            int      `json:"min_age"`
	VaccineName       string   `json:"vaccine"`
	Slots             []string `json:"slots"`
	Date              string   `json:"date"`
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/{district_id}/{age}", hitURL)
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

func hitURL(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	district_id, ok := vars["district_id"]
	if !ok {
		// default to 265
		district_id = "265"
	}

	date := time.Now().Format(dateLayout)
	url := fmt.Sprintf("https://cdn-api.co-vin.in/api/v2/appointment/sessions/calendarByDistrict?district_id=%v&date=%v", district_id, date)

	centers := Centers{}
	availabilites := []OutputInfo{}

	log.Printf("hitting: %v", url)

	resp, err := http.Get(url)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	// log.Println(string(body))

	err = json.Unmarshal(body, &centers)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	varAge, ok := vars["age"]
	if !ok {
		varAge = "18"
	}

	minAge, err := strconv.Atoi(varAge)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	for _, center := range centers.Centers {
		for _, s := range center.Sessions {
			if filterConditions(s, minAge) {
				availabilites = append(availabilites, createOutput(center, s))
			}
		}
	}

	output, err := json.MarshalIndent(availabilites, " ", " ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	log.Printf("minAge: %v, No. Of Availabilites: %v", minAge, len(availabilites))

	w.Write(output)
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

func createOutput(center Center, s Session) OutputInfo {
	return OutputInfo{
		CenterID:          center.ID,
		CenterName:        center.Name,
		Pincode:           center.Pincode,
		FeeType:           center.FeeType,
		AvailableCapacity: s.AvailableCapacity,
		MinAge:            s.MinAge,
		VaccineName:       s.VaccineName,
		Slots:             s.Slots,
		Date:              s.Date,
	}
}
