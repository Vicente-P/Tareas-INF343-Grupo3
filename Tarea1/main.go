package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/olekukonko/tablewriter"
)

type Departure struct {
	IataCode string `json:"iataCode"`
	At       string `json:"at"`
}

type Arrival struct {
	IataCode string `json:"iataCode"`
	At       string `json:"at"`
}

type Aircraft struct {
	Code string `json:"code"`
}

type Segment struct {
	Departure       Departure `json:"departure"`
	Arrival         Arrival   `json:"arrival"`
	CarrierCode     string    `json:"carrierCode"`
	Number          string    `json:"number"`
	Aircraft        Aircraft  `json:"aircraft"`
	Duration        string    `json:"duration"`
	Id              string    `json:"id"`
	NumberOfStops   int       `json:"numberOfStops"`
	BlacklistedInEU bool      `json:"blacklistedInEU"`
}

type Itinerary struct {
	Duration string    `json:"duration"`
	Segments []Segment `json:"segments"`
}

type Fee struct {
	Amount string `json:"amount"`
	Type   string `json:"type"`
}

type Price struct {
	Currency   string `json:"currency"`
	Total      string `json:"total"`
	Base       string `json:"base"`
	Fees       []Fee  `json:"fees"`
	GrandTotal string `json:"grandTotal"`
}

type PricingOptions struct {
	FareType                []string `json:"fareType"`
	IncludedCheckedBagsOnly bool     `json:"includedCheckedBagsOnly"`
}

type AdditionalService struct {
	Amount string `json:"amount"`
	Type   string `json:"type"`
}

type FareDetailsBySegment struct {
	SegmentId           string `json:"segmentId"`
	Cabin               string `json:"cabin"`
	FareBasis           string `json:"fareBasis"`
	BrandedFare         string `json:"brandedFare"`
	Class               string `json:"class"`
	IncludedCheckedBags struct {
		Quantity int `json:"quantity"`
	} `json:"includedCheckedBags"`
}

type TravelerPricing struct {
	TravelerId   string `json:"travelerId"`
	FareOption   string `json:"fareOption"`
	TravelerType string `json:"travelerType"`
	Price        struct {
		Currency   string `json:"currency"`
		Total      string `json:"total"`
		Base       string `json:"base"`
		Fees       []Fee  `json:"fees"`
		GrandTotal string `json:"grandTotal"`
	} `json:"price"`
	FareDetailsBySegment []FareDetailsBySegment `json:"fareDetailsBySegment"`
	AdditionalServices   []AdditionalService    `json:"additionalServices"`
}
type FlightOffer struct {
	Type                     string            `json:"type"`
	Id                       string            `json:"id"`
	Source                   string            `json:"source"`
	InstantTicketingRequired bool              `json:"instantTicketingRequired"`
	NonHomogeneous           bool              `json:"nonHomogeneous"`
	OneWay                   bool              `json:"oneWay"`
	LastTicketingDate        string            `json:"lastTicketingDate"`
	NumberOfBookableSeats    int               `json:"numberOfBookableSeats"`
	Itineraries              []Itinerary       `json:"itineraries"`
	Price                    Price             `json:"price"`
	PricingOptions           PricingOptions    `json:"pricingOptions"`
	ValidatingAirlineCodes   []string          `json:"validatingAirlineCodes"`
	TravelerPricings         []TravelerPricing `json:"travelerPricings"`
}

type FlightOffersPricing struct {
	Data struct {
		Type         string        `json:"type"`
		FlightOffers []FlightOffer `json:"flightOffers"`
	} `json:"data"`
}

type AssociatedRecords struct {
	Reference        string `json:"reference"`
	CreationDate     string `json:"creationDate"`
	OriginSystemCode string `json:"originSystemCode"`
	FlightOfferId    string `json:"flightOfferId"`
}

type Phones struct {
	DeviceType         string `json:"deviceType"`
	CountryCallingCode string `json:"countryCallingCode"`
	Number             string `json:"number"`
}

type Travelers struct {
	Id          string `json:"id"`
	DateOfBirth string `json:"dateOfBirth"`
	Gender      string `json:"gender"`
	Name        struct {
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
	} `json:"name"`
	Contact struct {
		Phones       []Phones `json:"phones"`
		EmailAddress string   `json:"emailAddress"`
	} `json:"contact"`
}

type FlightBooking struct {
	Data struct {
		Type         string        `json:"type"`
		FlightOffers []FlightOffer `json:"flightOffers"`
		Travelers    []Travelers   `json:"travelers"`
	} `json:"data"`
}

type Booking struct {
	Data struct {
		Type              string              `json:"type"`
		Id                string              `json:"id"`
		QueuingOfficeId   string              `json:"queuingOfficeId"`
		AssociatedRecords []AssociatedRecords `json:"AssociatedRecords"`
		FlightOffers      []FlightOffer       `json:"flightOffers"`
		Travelers         []Travelers         `json:"travelers"`
	} `json:"data"`
}

type Reserva struct {
	Type              string `json:"type"`
	ID                string `json:"id"`
	QueuingOfficeID   string `json:"queuingOfficeId"`
	AssociatedRecords []struct {
		Reference        string `json:"reference"`
		CreationDate     string `json:"creationDate"`
		OriginSystemCode string `json:"originSystemCode"`
		FlightOfferId    string `json:"flightOfferId"`
	} `json:"AssociatedRecords"`
	FlightOffers []FlightOffer `json:"flightOffers"`
	Travelers    []Travelers   `json:"travelers"`
}

func main() {
	fmt.Println("Bienvenido a goTravel!")
	for {
		fmt.Println("\nMenú:")
		fmt.Println("1. Realizar búsqueda.")
		fmt.Println("2. Obtener reserva.")
		fmt.Println("3. Salir")
		fmt.Print("Ingrese una opción: ")
		var opcion string
		fmt.Scanln(&opcion)

		switch opcion {
		case "1":
			respuesta, adultos := realizarBusqueda()
			var vuelo int
			fmt.Print("Seleccione un vuelo (ingrese 0 para realizar nueva búsqueda): ")
			fmt.Scanln(&vuelo)

			if vuelo == 0 {
				// El cliente eligió realizar una nueva búsqueda
				continue // Vuelve al inicio del bucle
			}
			precios := obtenerPrecio(string(respuesta), vuelo)
			reserva := RealizarReserva(precios, vuelo, adultos)
			fmt.Println("Reserva creada con éxito: ", reserva)
		case "2":
			ObtenerReserva()
		case "3":
			fmt.Println("¡Hasta luego!")
			os.Exit(0)
		default:
			fmt.Println("Opción no válida. Intente de nuevo.")
		}
	}
}

func realizarBusqueda() ([]byte, string) {
	var origen, destino, fecha, adultos string

	fmt.Print("Aeropuerto de origen: ")
	fmt.Scanln(&origen)

	fmt.Print("Aeropuerto de destino: ")
	fmt.Scanln(&destino)

	fmt.Print("Fecha de salida (AAAA-MM-DD): ")
	fmt.Scanln(&fecha)

	fmt.Print("Cantidad de adultos: ")
	fmt.Scanln(&adultos)

	// Realiza una solicitud HTTP a servidor en "server.go" para buscar vuelos
	url := fmt.Sprintf("http://localhost:5000/search?originLocationCode=%s&destinationLocationCode=%s&departureDate=%s&adults=%s", origen, destino, fecha, adultos)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error al hacer la solicitud HTTP:", err)
		return nil, ""
	}
	defer resp.Body.Close()

	var flightOffers []FlightOffer
	// Lee el cuerpo de la respuesta HTTP
	respuestaHTTP, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error al leer la respuesta HTTP:", err)
		return nil, ""
	}

	if err := json.Unmarshal(respuestaHTTP, &flightOffers); err != nil {
		fmt.Println("Error al deserializar el JSON:", err)
		return nil, ""
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"VUELO", "NÚMERO", "HORA DE SALIDA", "HORA DE LLEGADA", "AVIÓN", "PRECIO TOTAL"})

	for _, offer := range flightOffers {
		// Añadir una fila a la tabla
		dateTimeStr := offer.Itineraries[0].Segments[0].Departure.At
		dateTimeStr2 := offer.Itineraries[0].Segments[0].Arrival.At
		layout := "2006-01-02T15:04:05"
		parsedTime, err := time.Parse(layout, dateTimeStr)
		if err != nil {
			fmt.Println("Error al analizar la fecha y hora:", err)
		}
		parsedTime2, err := time.Parse(layout, dateTimeStr2)
		if err != nil {
			fmt.Println("Error al analizar la fecha y hora:", err)
		}

		// Formatear el objeto de tiempo en el formato deseado
		formattedTime := parsedTime.Format("15:04")
		formattedTime2 := parsedTime2.Format("15:04")
		table.Append([]string{
			offer.Id,
			offer.Itineraries[0].Segments[0].CarrierCode + offer.Itineraries[0].Segments[0].Number,
			formattedTime,
			formattedTime2,
			offer.Itineraries[0].Segments[0].CarrierCode + offer.Itineraries[0].Segments[0].Aircraft.Code,
			offer.Price.Total,
		})
	}

	// Renderizar la tabla
	table.Render()

	return respuestaHTTP, adultos

}

func obtenerPrecio(flight string, numero_vuelo int) []FlightOffer {

	// Convertir el JSON original a una lista de mapas
	var flightOffers []FlightOffer
	if err := json.Unmarshal([]byte(flight), &flightOffers); err != nil {
		fmt.Println("Error al deserializar el JSON:", err)
	}

	// Crear una instancia de FlightOffersPricing y asignar los vuelos a la estructura
	flightOffersPricing := FlightOffersPricing{
		Data: struct {
			Type         string        `json:"type"`
			FlightOffers []FlightOffer `json:"flightOffers"`
		}{
			Type:         "flight-offers-pricing",
			FlightOffers: flightOffers,
		},
	}

	// Convertir la estructura FlightOffersPricing a JSON
	resultJSON, err := json.Marshal(flightOffersPricing)
	if err != nil {
		fmt.Println("Error al serializar el JSON:", err)
	}

	// Imprimir el JSON resultante con la clave "data"

	// Crear una solicitud HTTP POST
	req, err := http.NewRequest("POST", "http://localhost:5000/pricing", bytes.NewBuffer(resultJSON))

	req.Header.Set("Content-Type", "application/json")

	// Enviar la solicitud
	client := &http.Client{}
	resp, err := client.Do(req)

	defer resp.Body.Close()

	var precios []FlightOffer
	// Leer la respuesta
	respBody, err := io.ReadAll(resp.Body)

	if err := json.Unmarshal(respBody, &precios); err != nil {
		fmt.Println("Error al deserializar el JSON:", err)
	}
	var vuelo []FlightOffer
	var id_vuelo string = strconv.Itoa(numero_vuelo)
	for _, offer := range flightOffers {
		// Añadir una fila a la tabla
		if id_vuelo == offer.Id {
			fmt.Println("El precio total final es de: ", offer.TravelerPricings[0].Price.Total)
			vuelo = append(vuelo, offer)
		}

	}
	return vuelo
}

func RealizarReserva(flight []FlightOffer, numero_vuelo int, adultos string) string {

	var pasajeros []Travelers

	entero, err := strconv.Atoi(adultos)

	if err != nil {
		fmt.Println("Error al convertir la cadena a entero:", err)
	}

	for i := 0; i < entero; i++ {
		var fecha, nombre, apellido, sexo, correo, telefono string

		fmt.Printf("Datos del pasajero %d:\n", i+1)
		fmt.Print("Fecha Nacimiento (AAAA-MM-DD): ")
		fmt.Scanln(&fecha)

		fmt.Print("Nombre: ")
		fmt.Scanln(&nombre)

		fmt.Print("Apellido: ")
		fmt.Scanln(&apellido)

		fmt.Print("Sexo: ")
		fmt.Scanln(&sexo)

		fmt.Print("Correo: ")
		fmt.Scanln(&correo)

		fmt.Print("Telefono: ")
		fmt.Scanln(&telefono)

		traveler := Travelers{
			Id:          strconv.Itoa(i + 1), // Puedes usar el número de vuelo como ID
			DateOfBirth: fecha,
			Gender:      sexo,
			Name: struct {
				FirstName string `json:"firstName"`
				LastName  string `json:"lastName"`
			}{
				FirstName: nombre,
				LastName:  apellido,
			},
			Contact: struct {
				Phones       []Phones `json:"phones"`
				EmailAddress string   `json:"emailAddress"`
			}{
				Phones: []Phones{
					{
						DeviceType:         "MOBILE",      // Puedes establecer esto según corresponda
						CountryCallingCode: telefono[1:3], // Puedes establecer esto según corresponda
						Number:             telefono[3:],  // Usar el número de teléfono recopilado
					},
				},
				EmailAddress: correo,
			},
		}

		pasajeros = append(pasajeros, traveler)
	}

	// Convertir el JSON original a una lista de mapas

	// Crear una instancia de FlightBooking y asignar pasajero
	flightbooking := FlightBooking{
		Data: struct {
			Type         string        `json:"type"`
			FlightOffers []FlightOffer `json:"flightOffers"`
			Travelers    []Travelers   `json:"travelers"`
		}{
			Type:         "flight-order",
			FlightOffers: flight,
			Travelers:    pasajeros,
		},
	}

	// Convertir la estructura FlightOffersPricing a JSON
	resultJSON, err := json.Marshal(flightbooking)
	if err != nil {
		fmt.Println("Error al serializar el JSON:", err)
		return ""
	}

	req, err := http.NewRequest("POST", "http://localhost:5000/booking", bytes.NewBuffer(resultJSON))
	if err != nil {
		return ""
	}

	req.Header.Set("Content-Type", "application/json")

	// Enviar la solicitud
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	// Leer la respuesta
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	return string(respBody)
}

func ObtenerReserva() error {
	var id string

	fmt.Print("Ingrese el ID de la Reserva: ")
	fmt.Scanln(&id)

	url := fmt.Sprintf("http://localhost:5000/booking?id=%s", id)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error al hacer la solicitud HTTP:", err)
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var reserva Reserva
	if err := json.Unmarshal([]byte(respBody), &reserva); err != nil {
		fmt.Println("Error al analizar JSON:", err)
	}

	table1 := tablewriter.NewWriter(os.Stdout)
	fmt.Println("Resultado:")
	table1.SetHeader([]string{"NÚMERO", "HORA DE SALIDA", "HORA DE LLEGADA", "AVIÓN", "PRECIO TOTAL"})
	// Renderizar la tabla
	for _, offer := range reserva.FlightOffers {
		// Añadir una fila a la tabla
		dateTimeStr := offer.Itineraries[0].Segments[0].Departure.At
		dateTimeStr2 := offer.Itineraries[0].Segments[0].Arrival.At
		layout := "2006-01-02T15:04:05"
		parsedTime, err := time.Parse(layout, dateTimeStr)
		if err != nil {
			fmt.Println("Error al analizar la fecha y hora:", err)
		}
		parsedTime2, err := time.Parse(layout, dateTimeStr2)
		if err != nil {
			fmt.Println("Error al analizar la fecha y hora:", err)
		}

		// Formatear el objeto de tiempo en el formato deseado
		formattedTime := parsedTime.Format("15:04")
		formattedTime2 := parsedTime2.Format("15:04")
		table1.Append([]string{
			offer.Itineraries[0].Segments[0].CarrierCode + offer.Itineraries[0].Segments[0].Number,
			formattedTime,
			formattedTime2,
			offer.Itineraries[0].Segments[0].CarrierCode + offer.Itineraries[0].Segments[0].Aircraft.Code,
			offer.Price.Total,
		})
	}

	// Renderizar la tabla
	table1.Render()

	table := tablewriter.NewWriter(os.Stdout)
	fmt.Println("Pasajeros:")
	table.SetHeader([]string{"Nombre", "Apellido"})
	// Renderizar la tabla
	for _, traveler := range reserva.Travelers {

		table.Append([]string{
			traveler.Name.FirstName,
			traveler.Name.LastName,
		})
	}

	// Renderizar la tabla
	table.Render()

	return nil

}
