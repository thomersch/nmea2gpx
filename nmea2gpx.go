package main

import (
	"bufio"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	in := os.Stdin
	out := os.Stdout

	scanner := bufio.NewScanner(in)

	xmlHeader := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?><gpx version="1.0" creator="nmea2gpx by Thomas Skowron"><time>%v</time><trk><trkseg>`, time.Now().Format(time.RFC3339))

	_, err := out.Write([]byte(xmlHeader))
	if err != nil {
		log.Fatal(err)
	}

	var (
		gprmc trkpt
		gpgga trkpt
	)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "$GPRMC") {
			gprmc, err = parseGPRMC(line)
			if err != nil {
				continue
			}
		}

		if strings.HasPrefix(line, "$GPGGA") {
			gpgga, err = parseGPGGA(line)
			if err != nil {
				continue
			}
		}

		merged, err := merge(gprmc, gpgga)
		if err != nil {
			continue
		}
		x, err := xml.Marshal(merged)
		if err != nil {
			log.Fatal(err)
		}
		_, err = out.Write(x)
		if err != nil {
			log.Fatal(err)
		}
	}

	xmlTrail := fmt.Sprintf(`</trkseg></trk></gpx>
`)

	_, err = out.Write([]byte(xmlTrail))
	if err != nil {
		log.Fatal(err)
	}
}

type trkpt struct {
	Date time.Time `xml:"time"`
	Lat  float64   `xml:"lat,attr"`
	Lon  float64   `xml:"lon,attr"`
	Ele  float64   `xml:"ele"`
	Sats int       `xml:"sat"`
}

func parseGPRMC(ln string) (trkpt, error) {
	parts := strings.Split(ln, ",")

	time, err := time.Parse("150405.000-020106", parts[1]+"-"+parts[9])
	if err != nil {
		return trkpt{}, err
	}

	lat := latToDecimal(parts[3])
	lon := lonToDecimal(parts[5])

	return trkpt{
		Date: time,
		Lat:  lat,
		Lon:  lon,
	}, nil
}

func parseGPGGA(ln string) (trkpt, error) {
	parts := strings.Split(ln, ",")

	time, err := time.Parse("150405.000", parts[1])
	if err != nil {
		return trkpt{}, err
	}

	lat := latToDecimal(parts[2])
	lon := lonToDecimal(parts[4])

	ele, err := strconv.ParseFloat(parts[9], 64)
	if err != nil {
		return trkpt{}, err
	}

	nSats, err := strconv.Atoi(parts[7])
	if err != nil {
		return trkpt{}, err
	}

	return trkpt{
		Date: time,
		Lat:  lat,
		Lon:  lon,
		Ele:  ele,
		Sats: nSats,
	}, nil
}

func merge(gprmc, gpgga trkpt) (trkpt, error) {
	if !(gpgga.Date.Hour() == gprmc.Date.Hour() &&
		gpgga.Date.Minute() == gprmc.Date.Minute() &&
		gpgga.Date.Second() == gprmc.Date.Second()) {
		return trkpt{}, errors.New("skip")
	}

	const precision = 1000000
	lat := (math.Round(gpgga.Lat * precision)) / precision
	lon := (math.Round(gpgga.Lon * precision)) / precision

	if lat == 0 && lon == 0 {
		return trkpt{}, errors.New("skip")
	}

	return trkpt{
		Date: gprmc.Date,
		Lat:  lat,
		Lon:  lon,
		Ele:  gpgga.Ele,
		Sats: gpgga.Sats,
	}, nil
}

func latToDecimal(s string) float64 {
	return toDecimal(splitDegsMins(s, 2))
}

func lonToDecimal(s string) float64 {
	return toDecimal(splitDegsMins(s, 3))
}

func splitDegsMins(s string, degDigits int) (int, float64) {
	degs := s[:degDigits]
	mins := s[degDigits:]

	deg, err := strconv.Atoi(degs)
	if err != nil {
		deg = 0
	}
	min, err := strconv.ParseFloat(mins, 64)
	if err != nil {
		min = 0
	}
	return deg, min
}

func toDecimal(deg int, min float64) float64 {
	const secondsInMinute = 60
	return float64(deg) + (min / secondsInMinute)
}
