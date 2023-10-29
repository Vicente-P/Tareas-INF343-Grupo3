package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/streadway/amqp"
)

type Customer struct {
	Name     string   `json:"name"`
	LastName string   `json:"lastname"`
	Email    string   `json:"email"`
	Location Location `json:"location"`
	Phone    string   `json:"phone"`
}

type Location struct {
	Address1   string `json:"address1"`
	Address2   string `json:"address2"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postalcode"`
	Country    string `json:"country"`
	Phone      string `json:"phone"`
}

type Product struct {
	Title       string  `json:"title"`
	Author      string  `json:"author"`
	Genre       string  `json:"genre"`
	Pages       int     `json:"pages"`
	Publication string  `json:"publication"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
}

type OrderData struct {
	OrderID  string    `json:"orderID"`
	GroupID  string    `json:"groupID"`
	Products []Product `json:"products"`
	Customer Customer  `json:"customer"`
}

func main() {
	// Conectar a RabbitMQ
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672")
	if err != nil {
		log.Fatalf("Error al conectar a RabbitMQ: %v", err)
		return
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Error al abrir un canal: %v", err)
		return
	}
	defer ch.Close()

	// Nombre de la cola
	queueName := "Queue3"

	// Declarar la cola
	_, err = ch.QueueDeclare(
		queueName, // Nombre de la cola
		false,     // Durabilidad
		false,     // Eliminar cuando no esté en uso
		false,     // Exclusiva
		false,     // No esperar confirmación
		nil,       // Argumentos adicionales
	)
	if err != nil {
		log.Fatalf("Error al declarar la cola: %v", err)
		return
	}

	// Configura la función que manejará los mensajes entrantes
	msgs, err := ch.Consume(
		queueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Error al consumir mensajes: %v", err)
		return
	}

	log.Printf("Servicio de Notificaciones escuchando mensajes...")

	// Procesar los mensajes entrantes y enviar notificaciones
	for d := range msgs {
		// Deserializar el mensaje JSON
		var orderData OrderData
		err := json.Unmarshal(d.Body, &orderData)
		if err != nil {
			log.Printf("Error al deserializar el mensaje JSON: %v", err)
		} else {
			orderData.GroupID = "K4q!6D2f#8"

			jsonData, err := json.Marshal(orderData)
			if err != nil {
				fmt.Println("Error al convertir la estructura a JSON:", err)
				return
			}

			apiURL := "https://sjwc0tz9e4.execute-api.us-east-2.amazonaws.com/Prod"

			// Realizar la solicitud POST a la API
			resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				fmt.Println("Error al enviar la solicitud POST:", err)
				return
			}
			defer resp.Body.Close()

			// Verificar la respuesta de la API
			if resp.StatusCode == http.StatusOK {
				fmt.Println("Solicitud exitosa. Respuesta de la API:", resp.Status)
			} else {
				fmt.Println("La solicitud no fue exitosa. Respuesta de la API:", resp.Status)
			}
		}
	}

}
