package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gopkg.in/resty.v1"
)

type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
}

// ESTRUCTURA PARA EL BOOKING? XD
type AssociatedRecords struct {
	Reference        string `json:"reference"`
	CreationDate     string `json:"creationDate"`
	OriginSystemCode string `json:"originSystemCode"`
	FlightOfferId    string `json:"flightOfferId"`
}

type Documents struct {
	Number           string `json:"number"`
	IssuanceDate     string `json:"issuanceDate"`
	ExpiryDate       string `json:"expiryDate"`
	IssuanceCountry  string `json:"issuanceCountry"`
	IssuanceLocation string `json:"issuanceLocation"`
	Nationality      string `json:"nationality"`
	BirthPlace       string `json:"birthPlace"`
	DocumentType     string `json:"documentType"`
	Holder           bool   `json:"holder"`
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
	Documents []Documents `json:"documents"`
	Contact   struct {
		Purpose      string   `json:"purpose"`
		Phones       []Phones `json:"phones"`
		EmailAddress string   `json:"emailAddress"`
	} `json:"contact"`
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
type Links struct {
	Self string `json:"self"`
}

type Meta struct {
	Count int   `json:"count"`
	Links Links `json:"links"`
}

type FlightOffersResponse struct {
	Meta Meta          `json:"meta"`
	Data []FlightOffer `json:"data"`
}

type FlightOffersPricing struct {
	Data struct {
		Type         string        `json:"type"`
		FlightOffers []FlightOffer `json:"flightOffers"`
	} `json:"data"`
}

func obtenerToken() (string, error) {
	// Obtén el Client ID y la Client Secret desde las variables de entorno
	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("SECRECT_ID")

	if clientID == "" || clientSecret == "" {
		return "", fmt.Errorf("Las variables de entorno CLIENT_ID y CLIENT_SECRET no están configuradas")
	}

	// Configura la URL de la API de Amadeus para obtener el token
	apiUrl := "https://test.api.amadeus.com/v1/security/oauth2/token"

	// Configura los datos del formulario para la solicitud POST
	data := map[string]string{
		"grant_type":    "client_credentials",
		"client_id":     clientID,
		"client_secret": clientSecret,
	}

	// Crea una instancia de Resty
	client := resty.New()

	// Realiza la solicitud POST para obtener el token
	resp, err := client.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetFormData(data).
		Post(apiUrl)

	if err != nil {
		return "", err
	}

	// Deserializa la respuesta JSON y obtén el token de acceso
	var tokenResponse AccessTokenResponse
	if err := json.Unmarshal(resp.Body(), &tokenResponse); err != nil {
		return "", err
	}

	return tokenResponse.AccessToken, nil
}

func buscarVuelos(c *gin.Context) {

	token, err := obtenerToken()
	if err != nil {
		fmt.Println("Error al obtener el token:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener el token"})
		return
	}

	origen := c.Query("originLocationCode")
	destino := c.Query("destinationLocationCode")
	fecha := c.Query("departureDate")
	adultos := c.Query("adults")

	params := map[string]string{
		"originLocationCode":      origen,
		"destinationLocationCode": destino,
		"departureDate":           fecha,
		"adults":                  adultos,
		"includedAirlineCodes":    "H2,LA,JA",
		"nonStop":                 "true",
		"currencyCode":            "CLP",
		"travelClass":             "ECONOMY",
	}

	apiUrl := "https://test.api.amadeus.com/v2/shopping/flight-offers"

	// Crea una instancia de Resty

	query := url.Values{}

	// Agregar los parámetros al objeto url.Values
	for key, value := range params {
		query.Add(key, value)
	}
	fullUrl := apiUrl + "?" + query.Encode()
	client := &http.Client{}

	// Crear una solicitud GET
	req, err := http.NewRequest("GET", fullUrl, nil)
	if err != nil {
		fmt.Println("Error al crear la solicitud:", err)
		return
	}

	req.Header.Add("Authorization", "Bearer "+token)

	// Realizar la solicitud con la instancia de http.Client personalizada
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error al realizar la solicitud:", err)
		return
	}
	defer resp.Body.Close()

	// Leer la respuesta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error al leer la respuesta:", err)
		return
	}
	var response FlightOffersResponse
	//Deserializar respuesta
	if err := json.Unmarshal(body, &response); err != nil {
		fmt.Println("Error al deserializar el JSON:", err)
		return
	}

	c.JSON(http.StatusOK, response.Data)
}

func obtenerPreciosAmadeus(c *gin.Context) {

	token, err := obtenerToken()
	if err != nil {
		fmt.Println("Error al obtener el token:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener el token"})
		return
	}

	apiUrl := "https://test.api.amadeus.com/v1/shopping/flight-offers/pricing"

	// Leer los datos JSON del cuerpo de la solicitud
	var datosJSON map[string]interface{}
	if err := c.ShouldBindJSON(&datosJSON); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convertir los datos JSON en bytes
	datosBytes, err := json.Marshal(datosJSON)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Crear una solicitud HTTP POST
	req, err := http.NewRequest("POST", apiUrl, bytes.NewBuffer(datosBytes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Enviar la solicitud
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error al leer la respuesta:", err)
		return
	}

	var response FlightOffersPricing
	//Deserializar respuesta
	if err := json.Unmarshal(body, &response); err != nil {
		fmt.Println("Error al deserializar el JSON:", err)
		return
	}

	// Devolver la respuesta al cliente
	c.JSON(http.StatusOK, response.Data.FlightOffers)
}

func hacerreserva(c *gin.Context) {
	token, err := obtenerToken()
	if err != nil {
		fmt.Println("Error al obtener el token:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener el token"})
		return
	}

	apiUrl := "https://test.api.amadeus.com/v1/booking/flight-orders"

	// Leer los datos JSON del cuerpo de la solicitud
	var datosJSON map[string]interface{}
	if err := c.ShouldBindJSON(&datosJSON); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convertir los datos JSON en bytes
	datosBytes, err := json.Marshal(datosJSON)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Crear una solicitud HTTP POST
	req, err := http.NewRequest("POST", apiUrl, bytes.NewBuffer(datosBytes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Enviar la solicitud
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error al leer la respuesta:", err)
		return
	}

	//DB
	// Establece la cadena de conexión de MongoDB
	connectionString := os.Getenv("CONNECTION_STRING")

	// Configura las opciones de conexión
	clientOptions := options.Client().ApplyURI(connectionString)

	// Crea un nuevo cliente de MongoDB
	clientDB, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Comprueba si la conexión es exitosa
	err = clientDB.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Conexión a MongoDB exitosa!")

	// Cierre la conexión cuando haya terminado
	defer clientDB.Disconnect(context.TODO())

	var response Booking
	//Deserializar respuesta
	if err := json.Unmarshal(body, &response); err != nil {
		fmt.Println("Error al deserializar el JSON:", err)
		return
	}

	//guardar en la coleccion de mongodb
	database := clientDB.Database("testgo")
	collection := database.Collection("flightofferts")
	collection.InsertOne(context.TODO(), response)

	// Devolver la respuesta al cliente
	c.JSON(http.StatusOK, response.Data.Id)

}

func buscarId(c *gin.Context) {
	token, err := obtenerToken()
	if err != nil {
		fmt.Println("Error al obtener el token:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener el token"})
		return
	}

	//Leer parametros
	id_aux := c.Query("id")
	apiUrl := "https://test.api.amadeus.com/v1/booking/flight-orders/" + id_aux

	// Crear una solicitud HTTP GET
	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		fmt.Println("Error al crear la solicitud:", err)
		return
	}

	req.Header.Set("Authorization", "Bearer "+token)

	// Enviar la solicitud
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error al leer la respuesta:", err)
		return
	}

	var response Booking
	//Deserializar respuesta
	if err := json.Unmarshal(body, &response); err != nil {
		fmt.Println("Error al deserializar el JSON:", err)
		return
	}

	// Devolver la respuesta al cliente
	c.JSON(http.StatusOK, response.Data)

}

func main() {

	if err := godotenv.Load(); err != nil {
		fmt.Println("Error cargando el archivo .env:", err)
		return
	}

	r := gin.Default()

	r.GET("/search", buscarVuelos)
	r.POST("/pricing", obtenerPreciosAmadeus)
	r.POST("/booking", hacerreserva)
	r.GET("/booking", buscarId)

	r.Run("localhost:5000")
}
