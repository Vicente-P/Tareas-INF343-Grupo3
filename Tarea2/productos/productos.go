package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Estructura para representar un producto
type Product struct {
	Title       string  `json:"title"`
	Author      string  `json:"author"`
	Genre       string  `json:"genre"`
	Pages       int     `json:"pages"`
	Publication string  `json:"publication"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
}

func main() {
	// Establecer la cadena de conexión a MongoDB
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Establecer un contexto con timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Conectar al servidor de MongoDB
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	defer client.Disconnect(ctx)

	// Seleccionar la base de datos y la colección
	collection := client.Database("bookstore").Collection("products")

	// Crear un conjunto de productos para agregar
	productsToAdd := []Product{
		{
			Title:       "The Lord of the Rings",
			Author:      "J.R.R. Tolkien",
			Genre:       "Fantasy",
			Pages:       1224,
			Publication: "1954",
			Quantity:    98, // Cambiar la cantidad a 98
			Price:       20,
		},
		{
			Title:       "To Kill a Mockingbird",
			Author:      "Harper Lee",
			Genre:       "Fiction",
			Pages:       336,
			Publication: "1960",
			Quantity:    98, // Cambiar la cantidad a 98
			Price:       15,
		},
		{
			Title:       "Harry Potter and the Sorcerer's Stone",
			Author:      "J.K. Rowling",
			Genre:       "Fantasy",
			Pages:       320,
			Publication: "1997",
			Quantity:    98, // Cambiar la cantidad a 98
			Price:       18,
		},
		{
			Title:       "1984",
			Author:      "George Orwell",
			Genre:       "Dystopian",
			Pages:       328,
			Publication: "1949",
			Quantity:    98, // Cambiar la cantidad a 98
			Price:       12.5,
		},
		{
			Title:       "The Great Gatsby",
			Author:      "F. Scott Fitzgerald",
			Genre:       "Classic",
			Pages:       180,
			Publication: "1925",
			Quantity:    98, // Cambiar la cantidad a 98
			Price:       10,
		},
	}

	// Insertar los productos en la colección
	for _, product := range productsToAdd {
		_, err = collection.InsertOne(ctx, product)
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("Productos insertados con éxito.")
}
