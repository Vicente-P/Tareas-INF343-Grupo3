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

type Location struct {
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

	queueName := "Queue1"

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

	log.Printf("Servicio de Inventario escuchando mensajes...")

	// Procesar los mensajes entrantes y actualizar el stock en MongoDB
	forever := make(chan bool)

	go func() {
		for d := range msgs {
			// Deserializar el mensaje JSON
			var orderData OrderData
			err := json.Unmarshal(d.Body, &orderData)
			if err != nil {
				log.Printf("Error al deserializar el mensaje JSON: %v", err)
			} else {
				// Procesar los productos y actualizar el stock en MongoDB
				for _, product := range orderData.Products {
					// Acceder a la colección de productos y actualizar el stock
					collection := client.Database("bookstore").Collection("products")
					filter := bson.M{"title": product.Title}
					update := bson.M{"$inc": bson.M{"quantity": -product.Quantity}}
					updateResult, err := collection.UpdateOne(context.Background(), filter, update)
					if err != nil {
						log.Fatalf("Error al actualizar el producto: %v", err)
					}

					fmt.Printf("Producto actualizado: %s, Documentos modificados: %d\n", product.Title, updateResult.ModifiedCount)
				}
			}
		}
	}()

	fmt.Println("Successfuly connected to our RabbitMQ instance")
	fmt.Println(" [*] - waiting for messages")
	<-forever
}
