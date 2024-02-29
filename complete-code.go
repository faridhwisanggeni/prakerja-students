package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/fiber/v2/middleware/timeout"
	"github.com/gofiber/template/html"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Student model
type Student struct {
	ID          uint      `gorm:"primaryKey"`
	Nama        string    `gorm:"size:50"`
	Alamat      string    `gorm:"size:200"`
	IsActive    bool      `gorm:"default:true"`
	CreatedBy   string    `gorm:"size:10"`
	CreatedDate time.Time `gorm:"default:now()"`
	UpdatedBy   string    `gorm:"size:10"`
	UpdatedDate time.Time
}

var (
	db *gorm.DB
)

func init() {
	var err error
	dsn := "user=your_db_user password=your_db_password dbname=your_db_name host=your_db_host port=your_db_port sslmode=disable TimeZone=Asia/Jakarta"
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal(err)
	}

	// Auto Migrate
	db.AutoMigrate(&Student{})
}

func main() {
	engine := fiber.New(fiber.Config{
		Views: html.New("./views", ".html"),
	})

	engine.Use(requestid.New())
	engine.Use(logger.New())
	engine.Use(recover.New())
	engine.Use(timeout.New())

	// Example of rate limiting
	engine.Use(limiter.New(limiter.Config{
		Max: 5,
	}))

	engine.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, Fiber!")
	})

	// Routes for Student CRUD
	engine.Get("/students", getStudents)
	engine.Get("/students/:id", getStudent)
	engine.Post("/students", createStudent)
	engine.Put("/students/:id", updateStudent)
	engine.Delete("/students/:id", deleteStudent)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Fatal(engine.Listen(":" + port))
}

// Handlers
func getStudents(c *fiber.Ctx) error {
	var students []Student
	if err := db.Find(&students).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	return c.JSON(students)
}

func getStudent(c *fiber.Ctx) error {
	id := c.Params("id")
	var student Student
	if err := db.First(&student, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}
	return c.JSON(student)
}

func createStudent(c *fiber.Ctx) error {
	student := new(Student)
	if err := c.BodyParser(student); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}
	if err := db.Create(student).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	return c.JSON(student)
}

func updateStudent(c *fiber.Ctx) error {
	id := c.Params("id")
	student := new(Student)
	if err := c.BodyParser(student); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}
	if err := db.First(&Student{}, id).Updates(student).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	return c.JSON(student)
}

func deleteStudent(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := db.Delete(&Student{}, id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	return c.SendString(fmt.Sprintf("Student with ID %s is deleted", id))
}
