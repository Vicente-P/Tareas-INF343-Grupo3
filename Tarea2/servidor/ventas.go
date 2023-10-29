package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	pb "tarea2/proto"

	"github.com/google/uuid"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
)

var port = flag.Int("port", 50051, "The server port")

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

type server struct {
	pb.UnimplementedSalesServiceServer
	mongoClient *mongo.Client
}

func generaIDUnico() string {
	uuid := uuid.New()
	return uuid.String()
}

func (s *server) CreateOrder(ctx context.Context, order *pb.Order) (*pb.OrderResponse, error) {
	collection := s.mongoClient.Database("bookstore").Collection("orders")
	orderID := generaIDUnico()

	// Realiza la conversión de los productos de pb.Product a Product
	var products []Product
	for _, pbProduct := range order.Products {
		product := Product{
			Title:       pbProduct.Title,
			Author:      pbProduct.Author,
			Genre:       pbProduct.Genre,
			Pages:       int(pbProduct.Pages),
			Publication: pbProduct.Publication,
			Quantity:    int(pbProduct.Quantity),
			Price:       float64(pbProduct.Price),
		}
		products = append(products, product)
	}

	orderData := OrderData{
		OrderID:  orderID,
		Products: products,
		Customer: Customer{
			Name:     order.Customer.Name,
			LastName: order.Customer.Lastname,
			Email:    order.Customer.Email,
			Location: Location{
				Address1:   order.Customer.Location.Address1,
				Address2:   order.Customer.Location.Address2,
				City:       order.Customer.Location.City,
				State:      order.Customer.Location.State,
				PostalCode: order.Customer.Location.PostalCode,
				Country:    order.Customer.Location.Country,
				Phone:      order.Customer.Location.Phone,
			},
			Phone: order.Customer.Phone,
		},
	}

	// Convierte la orden a JSON
	orderDataBSON, err := bson.Marshal(orderData)
	if err != nil {
		log.Fatalf("Error al convertir la orden a BSON: %v", err)
		return nil, err
	}

	// Almacena la orden en la base de datos
	_, err = collection.InsertOne(ctx, orderDataBSON)
	if err != nil {
		log.Fatalf("Error al insertar la orden en MongoDB: %v", err)
		return nil, err
	}

	sendToRabbitMQ(orderData)

	// Devuelve una respuesta con el ID de la orden generada
	return &pb.OrderResponse{OrderID: orderID}, nil
}

func sendToRabbitMQ(orderData OrderData) {
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

	// Nombres de las colas de destino
	queues := []string{"Queue1", "Queue2", "Queue3"}

	// Declarar la cola
	// Declarar las colas de destino
	for _, queueName := range queues {
		_, err = ch.QueueDeclare(
			queueName, // Nombre de la cola
			false,     // Durabilidad
			false,     // Eliminar cuando no esté en uso
			false,     // Exclusiva
			false,     // No esperar confirmación
			nil,       // Argumentos adicionales
		)
		if err != nil {
			log.Fatalf("Error al declarar la cola %s: %v", queueName, err)
			return
		}
	}

	// Convierte el objeto OrderData a JSON
	orderDataJSON, err := json.Marshal(orderData)
	if err != nil {
		log.Fatalf("Error al convertir OrderData a JSON: %v", err)
		return
	}

	// Publicar el mismo mensaje en las tres colas
	for _, queueName := range queues {
		err = ch.Publish(
			"",        // Intercambio (en blanco para una cola directa)
			queueName, // Nombre de la cola
			false,     // Mandatorio
			false,     // Inmediato
			amqp.Publishing{
				ContentType: "application/json",
				Body:        orderDataJSON,
			},
		)
		if err != nil {
			log.Fatalf("Error al publicar mensaje en la cola %s: %v", queueName, err)
			return
		}

		log.Printf("Mensaje publicado en la cola '%s': %s", queueName, orderDataJSON)
	}

}

// message := fmt.Sprintf("Nueva orden creada: %s", orderID)
func main() {
	flag.Parse()

	// Conecta a MongoDB
	mongoClient, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatalf("Error al conectar a MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(context.TODO())

	// Crea un índice único en la colección de órdenes para garantizar IDs únicos.
	orderCollection := mongoClient.Database("bookstore").Collection("orders")
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "_id", Value: 1}},
		Options: options.Index(),
	}

	_, err = orderCollection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		log.Fatalf("Error al crear el índice en MongoDB: %v", err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Error al escuchar: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterSalesServiceServer(s, &server{mongoClient: mongoClient})

	log.Printf("Servidor de ventas escuchando en el puerto %d", *port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Error al servir: %v", err)
	}
}
