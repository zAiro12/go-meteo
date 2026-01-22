package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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
func getCityNameFromCoordinates(lat, lon float64) (city, country string) {
	reverseURL := fmt.Sprintf("https://nominatim.openstreetmap.org/reverse?format=json&lat=%.6f&lon=%.6f", lat, lon)

	resp, err := http.Get(reverseURL)
	if err != nil {
		return customLocationLabel, ""
	}
	defer resp.Body.Close()

	var reverseData struct {
		Address struct {
			City    string `json:"city"`
			Town    string `json:"town"`
			Village string `json:"village"`
			Country string `json:"country"`
		} `json:"address"`
	}

	if json.NewDecoder(resp.Body).Decode(&reverseData) != nil {
		return customLocationLabel, ""
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

	return city, reverseData.Address.Country
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
		location.City, location.Country = getCityNameFromCoordinates(location.Lat, location.Lon)
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
	location.City, location.Country = getCityNameFromCoordinates(lat, lon)

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
