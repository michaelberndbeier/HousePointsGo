package main

import (
	"bufio"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

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
	var speedsCSV = "./pointEvents.csv"
	file, _ := os.Open(speedsCSV)

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
func main() {

	http.HandleFunc("/points", pointsSite)

	http.ListenAndServe(":8090", nil)
}
