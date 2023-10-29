package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"

	pb "tarea2/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	addr = flag.String("addr", "localhost:50051", "the address to connect to")
)

func main() {

	flag.Parse()
	// Set up a connection to the server.
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewSalesServiceClient(conn)

	data, err := os.ReadFile("data.json")
	if err != nil {
		log.Fatalf("Error al leer el archivo data.json: %v", err)
	}

	// Parsea el contenido del archivo JSON en una estructura de orden
	var order pb.Order
	if err := json.Unmarshal(data, &order); err != nil {
		log.Fatalf("Error al parsear el archivo JSON: %v", err)
	}

	// Envía la orden al servicio de ventas
	response, err := c.CreateOrder(context.Background(), &order)
	if err != nil {
		log.Fatalf("Error al enviar la orden: %v", err)
	}

	// Imprime el ID de la orden generada por el servidor de ventas
	log.Printf("ID de la orden generada: %s", response.OrderID)
}
