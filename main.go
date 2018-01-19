package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
)

const (
	intercomLat = 0.930948639728
	intercomLong = -0.10921684028
	eathRadiusKM = 6371
)

// Location is the interface that a type can implement to provide access to its location
// coordinates.
// The values returned MUST be in RADIAN.
type Location interface {
	GetLatitude() float64
	GetLongitude() float64
}

// User type is used to hold decoded JSON data. This type also implements Location interface.
type User struct {
	UserId int
	Name   string
	Lat    float64
	Long   float64
}

func (u *User) GetLatitude() float64 {
	return u.Lat
}

func (u *User) GetLongitude() float64 {
	return u.Long
}

func (u *User) UnmarshalJSON(b []byte) error {
	var err error
	tu := struct {
		UserId int    `json:"user_id"`
		Name   string `json:"name"`
		Lat    string `json:"latitude"`
		Long   string `json:"longitude"`
	}{}
	if err := json.Unmarshal(b, &tu); err != nil {
		return err
	}
	u.UserId = tu.UserId
	u.Name = tu.Name
	u.Lat, u.Long, err = getLatLongRadian(tu.Lat, tu.Long)
	if err != nil {
		return err
	}
	
	return nil
}

func main() {
	file, err := os.Open("customer.json")
    if err != nil {
        log.Fatal("error opening file: ", err)
    }
    defer file.Close()

	intercom := &User{Lat: intercomLat, Long: intercomLong}

	userDistances, err := distances(file, intercom, float64(100))
	if err != nil {
        log.Fatal("error calculating distance: ", err)
	}
	
	sort.Slice(userDistances, func(i, j int) bool {
		return userDistances[i].UserId < userDistances[j].UserId
	})

	for _, u := range userDistances {
		fmt.Printf("%d %s\n", u.UserId, u.Name)
	}
}

// distances reads the JSON input, unmarshals it and compares the distance between
// input values and 'from'. It returns a slice of *User's who are within 'limit'
// distance from 'from'.
// This is a simple dequential implementation. In case of high data volume, input can
// be split into smaller slices and processed concurrently since the problem is
// embarrassingly parallel.
func distances(r io.Reader, from *User, limit float64) ([]*User, error) {
	uds := []*User{}
    scanner := bufio.NewScanner(r)
    for scanner.Scan() {
		u := &User{}
		if err := json.Unmarshal(scanner.Bytes(), u); err != nil {
			return nil, err
		}
		if distance(from, u) < limit {
			uds = append(uds, u)
		}
    }
    if err := scanner.Err(); err != nil {
        return nil, err
	}

	return uds, nil
}

// distance calculates the distance between two location that can provide their
// latitude and longitude in radians.
func distance(from, to Location) float64 {
	dlong := math.Abs(from.GetLongitude() - to.GetLongitude())
	centralAngle := math.Acos(math.Sin(from.GetLatitude()) * math.Sin(to.GetLatitude()) +
		math.Cos(from.GetLatitude()) * math.Cos(to.GetLatitude()) * math.Cos(dlong))

	return eathRadiusKM * centralAngle
}

// getLatLongRadian is a helper function to parse degree in strings and return
// corresponding radian values.
func getLatLongRadian(lat, long string) (float64, float64, error) {
	latr, err := parseRadianFromDegree(lat)
	if err != nil {
		return 0, 0, err
	}
	longr, err := parseRadianFromDegree(long)
	if err != nil {
		return 0, 0, err
	}
	return latr, longr, nil
}

// parseRadianFromDegree parses a string representation of a float value. Since this
// function is used after call to 
func parseRadianFromDegree(deg string) (float64, error) {
	degf, err := strconv.ParseFloat(deg, 64)
	if err != nil {
		return 0, err
	}
	return radian(degf), nil
}

func radian(deg float64) float64 {
	return deg * math.Pi / 180
}