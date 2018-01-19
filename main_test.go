package main

import (
	"math"
	"strings"
	"testing"
)

func TestParseRadian(t *testing.T) {
	cases := []struct{
		in string
		want float64
		errors bool
	}{
		{"0", 0, false},
		{"180", math.Pi, false},
		{"360", 2 * math.Pi, false},
		{"invalid input", 0, true},
	}
	for _, c := range cases {
		// TODO: susceptible to round off error, include error delta.
		if actual, err := parseRadianFromDegree(c.in); err == nil {
			if c.want != actual {
				t.Errorf("parseRadianFromDegree(%s) failure: expected %v, actual %v", 
					c.want, actual)
			}
		} else {
			if !c.errors {
				t.Errorf("parseRadianFromDegree(%s) failure: unexpected error %v", err)
			}
		}
	}
}

func TestDistance(t *testing.T) {
	cases := []struct{
		from *User
		to *User
		errDelta float64
		want float64
	}{
		{
			&User{Lat: intercomLat, Long: intercomLong},
			&User{Lat: intercomLat, Long: intercomLong},
			0,
			0,
		},
		{
			&User{Lat: radian(25.3209013), Long: radian(82.921069)},  // Varanasi, India
			&User{Lat: radian(28.5274229), Long: radian(77.1389452)}, // New Delhi, India
			1,
			675,
		},
		{
			// antipodal land based locations on the equator
			&User{Lat: radian(0.175781), Long: radian(-63.281250)},  // Brasil
			&User{Lat: radian(-0.175781), Long: radian(116.718750)}, // Indonasia
			100, // experimentation with a map to obtain better coordinates can reduce the error.
			20000, // (circumference of earth / 2) approximately
		},
	}
	for _, c := range cases {
		if actual := distance(c.from, c.to); !eqWithErr(actual, c.want, c.errDelta) {
			t.Errorf("distance(%v, %v) failure: expected %f, actual %f", c.from, c.to, c.want, actual)
		}
	}
}

func TestDistances(t *testing.T) {
	data := `{"latitude": "51.92893", "user_id": 1, "name": "Alice Cahill", "longitude": "-10.27699"}
	{"latitude": "53.2451022", "user_id": 2, "name": "Ian Kehoe", "longitude": "-6.238335"}`
	r := strings.NewReader(data)

	from := &User{Lat: intercomLat, Long: intercomLong}
	want := []*User{&User{UserId: 2}}

	ud, err := distances(r, from, float64(100))
	if err != nil {
        t.Errorf("distances failure: unexpected error: ", err)
	}

	if len(ud) != 1 {
		t.Errorf("distances failure: unexpected number of results, expected %d, actual %d",
			len(want), len(ud))
	}

	if ud[0].UserId != want[0].UserId {
		t.Errorf("distances failure: want user id %d, actual %d", want[0].UserId, ud[0].UserId)
	}
}

func TestDistancesError(t *testing.T) {
	data := `{"latitude": "51.92893", "user_id": 1, "name": "Alice Cahill", "longitude": "-10.27699"}
	{"latitude": "FFFFFFFF", "user_id": 2, "name": "Case Should Fail", "longitude": "-6.238335"}`
	r := strings.NewReader(data)

	from := &User{Lat: intercomLat, Long: intercomLong}
	
	_, err := distances(r, from, float64(100))
	if err == nil {
        t.Errorf("distances failure: function should fail")
	}
}

func eqWithErr(val, to, errDelta float64) bool {
	return val - errDelta <= to && val + errDelta >= to
}