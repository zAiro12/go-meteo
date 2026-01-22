package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hectormalot/omgo"
)

// getWeatherDescription restituisce la descrizione testuale del codice meteo
func getWeatherDescription(code int) string {
	descriptions := map[int]string{
		0:  "☀️ Sereno",
		1:  "🌤️ Prevalentemente sereno",
		2:  "⛅ Parzialmente nuvoloso",
		3:  "☁️ Nuvoloso",
		45: "🌫️ Nebbia",
		48: "🌫️ Nebbia con brina",
		51: "🌦️ Pioviggine leggera",
		53: "🌦️ Pioviggine moderata",
		55: "🌧️ Pioviggine intensa",
		61: "🌧️ Pioggia leggera",
		63: "🌧️ Pioggia moderata",
		65: "🌧️ Pioggia intensa",
		71: "❄️ Neve leggera",
		73: "❄️ Neve moderata",
		75: "❄️ Neve intensa",
		77: "❄️ Granelli di neve",
		80: "🌧️ Rovesci di pioggia leggeri",
		81: "⛈️ Rovesci di pioggia moderati",
		82: "⛈️ Rovesci di pioggia violenti",
		85: "🌨️ Rovesci di neve leggeri",
		86: "🌨️ Rovesci di neve intensi",
		95: "⛈️ Temporale",
		96: "⛈️ Temporale con grandine leggera",
		99: "⛈️ Temporale con grandine intensa",
	}
	if desc, ok := descriptions[code]; ok {
		return desc
	}
	return "❓ Condizione sconosciuta"
}

// getCityNameFromCoordinates usa reverse geocoding per ottenere il nome della città
func getCityNameFromCoordinates(lat, lon float64) (city, province, country string) {
	reverseURL := fmt.Sprintf("https://nominatim.openstreetmap.org/reverse?format=json&lat=%.6f&lon=%.6f", lat, lon)

	resp, err := http.Get(reverseURL)
	if err != nil {
		return customLocationLabel, "", ""
	}
	defer resp.Body.Close()

	var reverseData struct {
		Address struct {
			City     string `json:"city"`
			Town     string `json:"town"`
			Village  string `json:"village"`
			County   string `json:"county"`
			State    string `json:"state"`
			Region   string `json:"region"`
			Province string `json:"province"`
			Country  string `json:"country"`
		} `json:"address"`
	}

	if json.NewDecoder(resp.Body).Decode(&reverseData) != nil {
		return customLocationLabel, "", ""
	}

	city = reverseData.Address.City
	if city == "" {
		city = reverseData.Address.Town
	}
	if city == "" {
		city = reverseData.Address.Village
	}
	if city == "" {
		city = customLocationLabel
	}

	// determine province from available fields
	if reverseData.Address.County != "" {
		province = reverseData.Address.County
	} else if reverseData.Address.Province != "" {
		province = reverseData.Address.Province
	} else if reverseData.Address.State != "" {
		province = reverseData.Address.State
	} else if reverseData.Address.Region != "" {
		province = reverseData.Address.Region
	} else {
		province = ""
	}

	country = reverseData.Address.Country
	return city, province, country
}

// provinceCodeFromName attempts to map an Italian province/state name into a two-letter code
func provinceCodeFromName(name string) string {
	if name == "" {
		return ""
	}
	n := strings.ToLower(name)
	m := map[string]string{
		"milano": "MI", "milan": "MI", "milano città": "MI",
		"roma": "RM", "torino": "TO", "napoli": "NA", "genova": "GE",
		"venezia": "VE", "venice": "VE", "verona": "VR", "bari": "BA",
		"palermo": "PA", "catania": "CT", "firenze": "FI", "florence": "FI",
		"bologna": "BO", "cagliari": "CA", "trento": "TN", "bolzano": "BZ",
		"brescia": "BS", "bergamo": "BG", "monza": "MB", "modena": "MO",
		"padova": "PD", "vicenza": "VI", "como": "CO", "pavia": "PV",
		"messina": "ME", "taranto": "TA", "perugia": "PG", "ancona": "AN",
		"siena": "SI", "arezzo": "AR", "lecce": "LE", "salerno": "SA",
		"reggio calabria": "RC", "cosenza": "CS",
	}
	for k, v := range m {
		if strings.Contains(n, k) {
			return v
		}
	}
	// fallback: take first two letters
	cleaned := ""
	for _, r := range n {
		if r >= 'a' && r <= 'z' {
			cleaned += string(r)
			if len(cleaned) >= 2 {
				break
			}
		}
	}
	if len(cleaned) >= 2 {
		return strings.ToUpper(cleaned[:2])
	}
	return strings.ToUpper(cleaned)
}

// getWeather recupera i dati meteo per la posizione attuale o personalizzata
func getWeather() (*WeatherData, error) {
	var location GeoLocation

	// Usa coordinate personalizzate se impostate
	locationMutex.RLock()
	if useCustom {
		location.Lat = customLat
		location.Lon = customLon
		locationMutex.RUnlock()
		city, province, country := getCityNameFromCoordinates(location.Lat, location.Lon)
		// Format city with province code when available: "City (PR), Country"
		pcode := provinceCodeFromName(province)
		if pcode != "" {
			location.City = fmt.Sprintf("%s (%s), %s", city, pcode, country)
		} else if city != "" && country != "" {
			location.City = fmt.Sprintf("%s, %s", city, country)
		} else {
			location.City = city
		}
		location.Country = country
	} else {
		locationMutex.RUnlock()
		// Geolocalizzazione automatica
		resp, err := http.Get("http://ip-api.com/json/")
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if err := json.NewDecoder(resp.Body).Decode(&location); err != nil {
			return nil, err
		}
	}

	client := omgo.NewClient()
	req, err := omgo.NewForecastRequest(location.Lat, location.Lon)
	if err != nil {
		return nil, err
	}

	req.WithHourly(
		omgo.HourlyTemperature2m,
		omgo.HourlyWeatherCode,
		omgo.HourlyPrecipitation,
		omgo.HourlyWindSpeed10m,
		omgo.HourlyRelativeHumidity2m,
		omgo.HourlyVisibility,
	).WithDaily(
		omgo.DailyTemperature2mMax,
		omgo.DailyTemperature2mMin,
		omgo.DailyWeatherCode,
	).WithTimezone("Europe/Rome")

	weather, err := client.Forecast(context.Background(), req)
	if err != nil {
		return nil, err
	}

	notificationsMutex.RLock()
	enabled := notificationsEnabled
	notificationsMutex.RUnlock()

	configMutex.RLock()
	interval := int(notificationInterval / time.Minute)
	start := notificationStartHour
	end := notificationEndHour
	configMutex.RUnlock()

	data := &WeatherData{
		City:                 location.City,
		Country:              location.Country,
		Lat:                  location.Lat,
		Lon:                  location.Lon,
		Time:                 time.Now().Format("15:04 - 02/01/2006"),
		CurrentCondition:     getWeatherDescription(int(weather.Hourly.WeatherCode[0])),
		CurrentTemp:          weather.Hourly.Temperature2m[0],
		Humidity:             weather.Hourly.RelativeHumidity2m[0],
		WindSpeed:            weather.Hourly.WindSpeed10m[0],
		Visibility:           weather.Hourly.Visibility[0] / 1000,
		Precipitation:        weather.Hourly.Precipitation[0],
		TodayMax:             weather.Daily.Temperature2mMax[0],
		TodayMin:             weather.Daily.Temperature2mMin[0],
		TodayCondition:       getWeatherDescription(int(weather.Daily.WeatherCode[0])),
		TomorrowMax:          weather.Daily.Temperature2mMax[1],
		TomorrowMin:          weather.Daily.Temperature2mMin[1],
		TomorrowCondition:    getWeatherDescription(int(weather.Daily.WeatherCode[1])),
		NotificationsEnabled: enabled,
		IntervalMinutes:      interval,
		StartHour:            start,
		EndHour:              end,
		Version:              AppVersion,
	}

	return data, nil
}

// getWeatherFor recupera i dati meteo per coordinate specifiche (usato per anteprima)
func getWeatherFor(lat, lon float64) (*WeatherData, error) {
	// crea location da coordinate fornite
	var location GeoLocation
	location.Lat = lat
	location.Lon = lon
	city, province, country := getCityNameFromCoordinates(lat, lon)
	pcode := provinceCodeFromName(province)
	if pcode != "" {
		location.City = fmt.Sprintf("%s (%s), %s", city, pcode, country)
	} else if city != "" && country != "" {
		location.City = fmt.Sprintf("%s, %s", city, country)
	} else {
		location.City = city
	}
	location.Country = country

	client := omgo.NewClient()
	req, err := omgo.NewForecastRequest(location.Lat, location.Lon)
	if err != nil {
		return nil, err
	}

	req.WithHourly(
		omgo.HourlyTemperature2m,
		omgo.HourlyWeatherCode,
		omgo.HourlyPrecipitation,
		omgo.HourlyWindSpeed10m,
		omgo.HourlyRelativeHumidity2m,
		omgo.HourlyVisibility,
	).WithDaily(
		omgo.DailyTemperature2mMax,
		omgo.DailyTemperature2mMin,
		omgo.DailyWeatherCode,
	).WithTimezone("Europe/Rome")

	weather, err := client.Forecast(context.Background(), req)
	if err != nil {
		return nil, err
	}

	notificationsMutex.RLock()
	enabled := notificationsEnabled
	notificationsMutex.RUnlock()

	configMutex.RLock()
	interval := int(notificationInterval / time.Minute)
	start := notificationStartHour
	end := notificationEndHour
	configMutex.RUnlock()

	data := &WeatherData{
		City:                 location.City,
		Country:              location.Country,
		Lat:                  location.Lat,
		Lon:                  location.Lon,
		Time:                 time.Now().Format("15:04 - 02/01/2006"),
		CurrentCondition:     getWeatherDescription(int(weather.Hourly.WeatherCode[0])),
		CurrentTemp:          weather.Hourly.Temperature2m[0],
		Humidity:             weather.Hourly.RelativeHumidity2m[0],
		WindSpeed:            weather.Hourly.WindSpeed10m[0],
		Visibility:           weather.Hourly.Visibility[0] / 1000,
		Precipitation:        weather.Hourly.Precipitation[0],
		TodayMax:             weather.Daily.Temperature2mMax[0],
		TodayMin:             weather.Daily.Temperature2mMin[0],
		TodayCondition:       getWeatherDescription(int(weather.Daily.WeatherCode[0])),
		TomorrowMax:          weather.Daily.Temperature2mMax[1],
		TomorrowMin:          weather.Daily.Temperature2mMin[1],
		TomorrowCondition:    getWeatherDescription(int(weather.Daily.WeatherCode[1])),
		NotificationsEnabled: enabled,
		IntervalMinutes:      interval,
		StartHour:            start,
		EndHour:              end,
		Version:              AppVersion,
	}

	return data, nil
}
