package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"
)

type Delivery struct {
	ShippingAddress ShippingAddress `json:"shippingAddress"`
	ShippingMethod  string          `json:"shippingMethod"`
	TrackingNumber  string          `json:"trackingNumber"`
}

type OrderWithDeliveries struct {
	OrderID    string     `json:"orderid"`
	Products   []Product  `json:"products"`
	Customer   Customer   `json:"customer"`
	Deliveries []Delivery `json:"deliveries"`
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

type ShippingAddress struct {
	Name       string `json:"name"`
	LastName   string `json:"lastname"`
	Address1   string `json:"address1"`
	Address2   string `json:"address2"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postalcode"`
	Country    string `json:"country"`
	Phone      string `json:"phone"`
}

type Customer struct {
	Name     string   `json:"name"`
	LastName string   `json:"lastname"`
	Email    string   `json:"email"`
	Location Location `json:"location"`
	Phone    string   `json:"phone"`
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
	OrderID  string    `json:"orderid"`
	Products []Product `json:"products"`
	Customer Customer  `json:"customer"`
}

func main() {
	// Conectar a MongoDB
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatalf("Error al conectar a MongoDB: %v", err)
	}
	defer client.Disconnect(context.TODO())

	// Conectar a RabbitMQ
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Error al conectar a RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Error al abrir un canal: %v", err)
	}
	defer ch.Close()

	// Nombre de la cola
	queueName := "Queue2"

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
	}

	log.Printf("Servicio de Despacho escuchando mensajes...")

	// Procesar los mensajes entrantes y actualizar la orden en MongoDB
	forever := make(chan bool)

	go func() {
		for d := range msgs {
			// Deserializar el mensaje JSON
			var orderData OrderWithDeliveries
			err := json.Unmarshal(d.Body, &orderData)
			if err != nil {
				log.Printf("Error al deserializar el mensaje JSON: %v", err)
			} else {
				// Actualizar la orden en MongoDB con las entregas (deliveries)

				shippingAddress := ShippingAddress{
					Name:       orderData.Customer.Name,
					LastName:   orderData.Customer.LastName,
					Address1:   orderData.Customer.Location.Address1,
					Address2:   orderData.Customer.Location.Address2,
					City:       orderData.Customer.Location.City,
					State:      orderData.Customer.Location.State,
					PostalCode: orderData.Customer.Location.PostalCode,
					Country:    orderData.Customer.Location.Country,
					Phone:      orderData.Customer.Location.Phone,
				}

				delivery := Delivery{
					ShippingAddress: shippingAddress,
					ShippingMethod:  "USPS",
					TrackingNumber:  "12345678901234567890",
				}
				orderData.Deliveries = append(orderData.Deliveries, delivery)
				collection := client.Database("bookstore").Collection("orders")
				filter := bson.M{"orderid": orderData.OrderID}
				update := bson.M{
					"$set": bson.M{"deliveries": []Delivery{delivery}},
				}
				_, err := collection.UpdateOne(context.Background(), filter, update)
				if err != nil {
					log.Fatalf("Error al actualizar la orden: %v", err)
				}

				fmt.Printf("Orden actualizada con entregas: %s\n", orderData.OrderID)
			}
		}
	}()

	fmt.Println("Successfuly connected to our RabbitMQ instance")
	fmt.Println(" [*] - waiting for messages")
	<-forever
}
