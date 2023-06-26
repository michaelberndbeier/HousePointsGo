package main

import (
	"bufio"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var housePointsCSV = "./pointEvents.csv"

type PointEvent struct {
	From        int
	FromName    string
	To          int
	ToName      string
	Why         string
	NumOfPoints int
	When        time.Time
}

type House int

const (
	Gryffindor House = 0
	Slytherin  House = 1
	Ravenclaw  House = 2
	Hufflepuff House = 3
)

type HousePoints struct {
	PointsSum int
	Events    []PointEvent
	HouseName string
}

type PointsSiteData struct {
	Gryffindor HousePoints
	Slytherin  HousePoints
	Hufflepuff HousePoints
	Ravenclaw  HousePoints
}

func pointsSite(w http.ResponseWriter, req *http.Request) {
	var data = getPointsSiteData()

	tmpl, _ := template.ParseFiles("./index.html")
	tmpl.Execute(w, data)
	fmt.Println("reload")
}

func getHouseNameFromHouse(house House) string {
	switch house {
	case Gryffindor:
		return "Gryffendor"
	case Slytherin:
		return "Syltherin"
	case Ravenclaw:
		return "Ravenclaw"
	case Hufflepuff:
		return "Hufflepuff"
	default:
		return "Hagrids Hütte"
	}
}

func getPointsSiteData() PointsSiteData {
	var housePoints [4]HousePoints

	for _, e := range readPointEvents() {
		var housePointsInstance = housePoints[House(e.To)]

		housePointsInstance.Events = append(housePointsInstance.Events, e)
		housePointsInstance.PointsSum = housePointsInstance.PointsSum + e.NumOfPoints
		housePointsInstance.HouseName = getHouseNameFromHouse(House(e.To))

		housePoints[House(e.To)] = housePointsInstance
	}

	var data = PointsSiteData{
		Gryffindor: housePoints[Gryffindor],
		Slytherin:  housePoints[Slytherin],
		Hufflepuff: housePoints[Hufflepuff],
		Ravenclaw:  housePoints[Ravenclaw],
	}

	return data
}

func readPointEvents() []PointEvent {
	file, _ := os.Open(housePointsCSV)

	defer file.Close()

	scanner := bufio.NewScanner(file)

	var pointEvents = []PointEvent{}

	for scanner.Scan() {
		line := scanner.Text()
		var splitted = strings.Split(line, "|")

		if len(splitted) == 5 {

			var FromString = splitted[0]
			var ToString = splitted[1]
			var NumOfPointsString = splitted[2]
			var Why = splitted[3]
			var WhenString = splitted[4]

			var NumOfPoints, err = strconv.ParseInt(NumOfPointsString, 10, 64)
			if err != nil {
				fmt.Println(err)
			}

			var From, err1 = strconv.ParseInt(FromString, 10, 8)
			if err1 != nil {
				fmt.Println(err)
			}

			var To, err2 = strconv.ParseInt(ToString, 10, 8)
			if err2 != nil {
				fmt.Println(err)
			}

			var WhenInt, err3 = strconv.ParseInt(WhenString, 10, 64)
			if err3 != nil {
				fmt.Println(err)
			}
			var When = time.Unix(WhenInt, 0)

			var pointEvent = PointEvent{
				From:        int(From),
				FromName:    getHouseNameFromHouse(House(From)),
				To:          int(To),
				ToName:      getHouseNameFromHouse(House(To)),
				Why:         Why,
				NumOfPoints: int(NumOfPoints),
				When:        When,
			}
			pointEvents = append(pointEvents, pointEvent)
		}
	}

	return pointEvents
}
func handlePointsForm(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("forms.html"))
	if r.Method != http.MethodPost {
		tmpl.Execute(w, nil)
		return
	}

	var fromString = r.FormValue("from")
	var toString = r.FormValue("to")

	var from, fromName, err1 = getHouseIdAndNameFromString(fromString)
	if err1 == true {
		tmpl.Execute(w, struct{ Success bool }{false})
		return
	}
	var to, toName, err2 = getHouseIdAndNameFromString(toString)
	if err2 == true {
		tmpl.Execute(w, struct{ Success bool }{false})
		return
	}

	var numOfPointsString = r.FormValue("numOfPoints")
	var numOfPoints, err3 = strconv.ParseInt(numOfPointsString, 10, 8)

	if err3 != nil {
		tmpl.Execute(w, struct{ Success bool }{false})
		return
	}

	pointEvent := PointEvent{
		From:        from,
		FromName:    fromName,
		To:          to,
		ToName:      toName,
		Why:         r.FormValue("why"),
		NumOfPoints: int(numOfPoints),
		When:        time.Now(),
	}

	// do something with details
	addHousePoints(pointEvent)

	tmpl.Execute(w, struct{ Success bool }{true})
}

func addHousePoints(event PointEvent) {

	file, err := os.OpenFile(housePointsCSV,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		log.Fatal("cannot open file: %s", err)
	}

	dataWriter := bufio.NewWriter(file)

	dataWriter.WriteString(getHousePointsString(event) + "\n")

	dataWriter.Flush()
	file.Close()

}

func getHousePointsString(event PointEvent) string {

	/*
		var FromString = splitted[0]
		var ToString = splitted[1]
		var NumOfPointsString = splitted[2]
		var Why = splitted[3]
		var WhenString = splitted[4]
	*/
	var whenUnix = event.When.Unix()

	var csvLine = fmt.Sprintf("%d|%d|%d|%s|%d",
		event.From,
		event.To,
		event.NumOfPoints,
		event.Why,
		whenUnix)

	return csvLine
}

func getHouseIdAndNameFromString(fromString string) (houseId int, houseName string, hadError bool) {
	var fromInt, err = strconv.ParseInt(fromString, 10, 8)

	if err != nil {
		return 99, "Error", true
	}

	if fromInt < 0 || fromInt > 3 {
		return 99, "Error", true
	}

	var house = House(fromInt)
	var name = getHouseNameFromHouse(house)

	return int(fromInt), name, false
}
func main() {

	http.HandleFunc("/", handlePointsForm)

	http.HandleFunc("/points", pointsSite)

	http.ListenAndServe(":8090", nil)
}
