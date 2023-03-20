package main 

import (
	"context"
	// "encoding/json"
	// "fmt"
	"log"
	"os"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Employee struct {
	ID string `json:"id,omitempty" bson:"_id,omitempty"`
	Name string `json:"name"`
	Salary float64 `json:"salary"`
	Age float64 `json:"age"`
}


func Connect() {
	
	
	
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatal("You must set your 'MONGODB_URI' environmental variable. See\n\t https://www.mongodb.com/docs/drivers/go/current/usage-examples/#environment-variable")
	}
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	} 
	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()
	coll := client.Database("divanshu").Collection("hrms")


	app := fiber.New()

	app.Get("/employee/", func(c *fiber.Ctx)error {
		query := bson.D{{}}

		cursor, err := coll.Find(c.Context(), query)
		if err!=nil {
			return c.Status(500).SendString(err.Error())
		}
		var employees []Employee = make([]Employee, 0)

		if err := cursor.All(c.Context(), &employees); err != nil{
			return c.Status(500).SendString(err.Error())
		}

		return c.JSON(employees)
	})
	
	app.Post("/employee/", func(c *fiber.Ctx)error{
		employee := new(Employee)

		if err := c.BodyParser(employee); err!= nil {
			return c.Status(400).SendString(err.Error())
		}

		employee.ID = ""

		insertionResult, err:= coll.InsertOne(c.Context(), employee)
		if err!= nil {
			return c.Status(500).SendString(err.Error())
		}

		filter:= bson.D{{Key:"_id", Value: insertionResult.InsertedID}}
		createdRecord := coll.FindOne(c.Context(), filter)

		createdEmployee := &Employee{}
		createdRecord.Decode(createdEmployee)

		return c.Status(201).JSON(createdEmployee)
	})

	app.Put("/employee/:id", func(c *fiber.Ctx)error {
		idParam := c.Params("id")

		employeeID, err := primitive.ObjectIDFromHex(idParam)

		if err!=nil{
			return c.SendStatus(400)
		}

		employee := new(Employee)

		if err := c.BodyParser(employee); err != nil{
			return c.Status(400).SendString(err.Error())
		}

		query := bson.D{{Key:"_id", Value: employeeID}}
		update := bson.D{
			{
				Key: "$set",
				Value: bson.D{
					{Key:"name", Value: employee.Name},
					{Key:"age", Value: employee.Age},
					{Key:"salary", Value: employee.Salary},
				},
			},
		}

		err = coll.FindOneAndUpdate(c.Context(), query, update).Err()

		if err != nil{
			if err == mongo.ErrNoDocuments{
				return c.SendStatus(400)
			}
			return c.SendStatus(500)
		}

		employee.ID = idParam

		return c.Status(200).JSON(employee)
	})
	app.Delete("/employee/:id", func(c *fiber.Ctx)error{

		employeeID, err := primitive.ObjectIDFromHex(
			c.Params("id"),
	)

	   if err != nil {
		return c.SendStatus(400)
	   }

	   query := bson.D{{Key: "_id", Value: employeeID}}
	   result, err := coll.DeleteOne(c.Context(), &query)

	   if err!= nil{
		return c.SendStatus(500)
	   }

	   if result.DeletedCount < 1 {
		return c.SendStatus(404)
	   }

	   return c.Status(200).JSON("record deleted")
	})

	log.Fatal(app.Listen(":3000"))
}