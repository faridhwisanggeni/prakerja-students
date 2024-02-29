package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/template/html/v2"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Pegawai model
type Pegawai struct {
	ID          uint   `gorm:"primaryKey"`
	Nama        string `gorm:"size:50"`
	Alamat      string `gorm:"size:200"`
	Jabatan     string `gorm:"size:50"`
	GajiPokok   int
	Tunjangan   int
	IsActive    bool      `gorm:"default:true"`
	CreatedBy   string    `gorm:"size:10"`
	CreatedDate time.Time `gorm:"default:now()"`
	UpdatedBy   string    `gorm:"size:10"`
	UpdatedDate time.Time
}

func (Pegawai) TableName() string {
	return "data_pegawai"
}

// Absensi model
type Absensi struct {
	IDAbsensi   uint `gorm:"primaryKey"`
	IDPegawai   uint
	IsActive    bool      `gorm:"default:true"`
	CreatedBy   string    `gorm:"size:10"`
	CreatedDate time.Time `gorm:"default:now()"`
	UpdatedBy   string    `gorm:"size:10"`
	UpdatedDate time.Time
}

func (Absensi) TableName() string {
	return "data_absensi_pegawai"
}

// Absensi model
type AbsensiPegawai struct {
	ID               uint   `json:"id_pegawai"`
	Nama             string `json:"nama_pegawai"`
	JumlahHadir      int    `json:"jumlah_hadir" gorm:"column:jumlah_hadir"`
	Jabatan          string `json:"jabatan_pegawai"`
	GajiPokok        int    `json:"gaji_pokok"`
	Tunjangan        int    `json:"tunjangan"`
	TotalGajiPegawai int    `json:"total_gaji_pegawai"`
}

var (
	db *gorm.DB
)

func init() {
	var err error
	dsn := `user=yrtrtredgdgdf 
	password=rararara 
	dbname=hahahahaha 
	host=123.123.123.123 port=32632121213222 
	sslmode=disable TimeZone=Asia/Jakarta`

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal(err)
	}

	// Auto Migrate
	db.AutoMigrate(&Pegawai{}, &Absensi{})
}

func main() {
	engine := fiber.New(fiber.Config{
		Views: html.New("./views", ".html"),
	})

	engine.Use(requestid.New())
	engine.Use(recover.New())

	// Example of rate limiting
	engine.Use(limiter.New(limiter.Config{
		Max: 5,
	}))

	engine.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, Fiber!")
	})

	// Routes for Pegawai CRUD
	engine.Get("/pegawais", getPegawais)
	engine.Get("/pegawais/:id", getPegawai)
	engine.Post("/pegawais", createPegawai)
	engine.Put("/pegawais/:id", updatePegawai)
	engine.Delete("/pegawais/:id", deletePegawai)

	// Routes for Absensi
	engine.Post("/absensi", createAbsensi)
	engine.Get("/absensi", getAllAbsensi)
	engine.Get("/absensi-logic", getAllAbsensiWithLogic)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Fatal(engine.Listen(":" + port))
}

// Handlers
func getPegawais(c *fiber.Ctx) error {
	var pegawais []Pegawai
	if err := db.Find(&pegawais).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	return c.JSON(pegawais)
}

func getPegawai(c *fiber.Ctx) error {
	id := c.Params("id")
	var pegawai Pegawai
	if err := db.Where("is_active = ?", true).First(&pegawai, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}
	return c.JSON(pegawai)
}

func createPegawai(c *fiber.Ctx) error {
	pegawai := new(Pegawai)
	if err := c.BodyParser(pegawai); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}
	if err := db.Create(pegawai).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	return c.JSON(pegawai)
}

func updatePegawai(c *fiber.Ctx) error {
	id := c.Params("id")
	pegawai := new(Pegawai)
	if err := c.BodyParser(pegawai); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	query := fmt.Sprintf(`UPDATE "data_pegawai" 
	SET "is_active"=%t, "nama"='%s', "alamat"='%s', "jabatan"='%s', 
		"gaji_pokok"=%d, "tunjangan"=%d WHERE "data_pegawai"."id" = '%s'`,
		pegawai.IsActive, pegawai.Nama, pegawai.Alamat, pegawai.Jabatan,
		pegawai.GajiPokok, pegawai.Tunjangan, id)

	if err := db.Exec(query).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.JSON(pegawai)
}

// Create Absensi
func createAbsensi(c *fiber.Ctx) error {
	absensi := new(Absensi)
	if err := c.BodyParser(absensi); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}
	if err := db.Create(absensi).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	return c.JSON(absensi)
}

func deletePegawai(c *fiber.Ctx) error {
	id := c.Params("id")

	query := fmt.Sprintf(`UPDATE data_pegawai SET is_active=false WHERE data_pegawai.id = '%s'`, id)

	if err := db.Exec(query).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	return c.SendString(fmt.Sprintf("Pegawai with ID %s is deactivated", id))
}

func deletePegawaiPermanen(c *fiber.Ctx) error {
	id := c.Params("id")
	var pegawai Pegawai
	db.Delete(&pegawai, id)
	return c.SendString(fmt.Sprintf("Pegawai with ID %s deleted", id))
}

// Get all Absensi
func getAllAbsensi(c *fiber.Ctx) error {
	var absensi []AbsensiPegawai

	// Define your raw SQL query
	query := `
		SELECT dp.id AS ID, dp.nama AS Nama, COUNT(1) AS jumlah_hadir,
			   dp.jabatan AS Jabatan, dp.gaji_pokok AS gaji_pokok, dp.tunjangan AS Tunjangan,
			   dp.gaji_pokok + dp.tunjangan AS total_gaji_pegawai
		FROM data_pegawai dp 
		LEFT JOIN data_absensi_pegawai dap ON dp.id = dap.id_pegawai 
		WHERE dp.is_active != false 
		GROUP BY dp.id, dp.nama, dp.jabatan, dp.gaji_pokok, dp.tunjangan
	`

	// Execute the raw SQL query
	if err := db.Raw(query).Scan(&absensi).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.JSON(absensi)
}

func getAllAbsensiWithLogic(c *fiber.Ctx) error {
	var absensi []AbsensiPegawai

	// Define your raw SQL query
	query := `
		SELECT dp.id AS ID, dp.nama AS Nama, COUNT(1) AS jumlah_hadir, dp.jabatan AS Jabatan, dp.gaji_pokok AS gaji_pokok, 
		       dp.tunjangan AS Tunjangan, dp.gaji_pokok + dp.tunjangan AS total_gaji_pegawai
		FROM data_pegawai dp 
		LEFT JOIN data_absensi_pegawai dap ON dp.id = dap.id_pegawai WHERE dp.is_active != false 
		GROUP BY dp.id, dp.nama, dp.jabatan, dp.gaji_pokok, dp.tunjangan`

	// Execute the raw SQL query
	rows, err := db.Raw(query).Rows()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var absensiPegawai AbsensiPegawai
		if err := db.ScanRows(rows, &absensiPegawai); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}
		//simple logic jika karyawan hadir dibawah 5 hari kerja akan di prorata
		if absensiPegawai.JumlahHadir < 5 {
			absensiPegawai.TotalGajiPegawai = (absensiPegawai.GajiPokok / 25) * absensiPegawai.JumlahHadir
		}
		absensi = append(absensi, absensiPegawai)
	}

	return c.JSON(absensi)
}
