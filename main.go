package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	client *mongo.Client
	db     *mongo.Database
)

type Employee struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name      string             `json:"name" bson:"name"`
	Type      string             `json:"type" bson:"type"` // "staff" or "intern"
	DeletedAt *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
}

type WorkStats struct {
	AverageVideoDuration    string `json:"averageVideoDuration"`
	AverageSoftwareDuration string `json:"averageSoftwareDuration"`
	AverageRevisionDuration string `json:"averageRevisionDuration"` // Revizyon süresi eklendi
	TotalWorks              int    `json:"totalWorks"`
}

type TimelineSlot struct {
	Hour  int    `json:"hour"`
	Works []Work `json:"works"`
}

type Work struct {
	ID                  primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	EmployeeID          primitive.ObjectID `json:"employeeId" bson:"employeeId"`
	EmployeeName        string             `json:"employeeName" bson:"employeeName"`
	WorkType            string             `json:"workType" bson:"workType"`                                   // "software", "video", "review", "revize"
	Description         string             `json:"description" bson:"description"`                             // Work description
	VideoLink           string             `json:"videoLink" bson:"videoLink"`                                 // Optional
	IsFirstVideo        bool               `json:"isFirstVideo" bson:"isFirstVideo"`                           // Only for videos
	IsRevision          bool               `json:"isRevision" bson:"isRevision"`                               // Whether this is a revision
	IsReviewed          bool               `json:"isReviewed" bson:"isReviewed"`                               // Whether this video has been reviewed
	IsRevisionCompleted bool               `json:"isRevisionCompleted" bson:"isRevisionCompleted"`             // Whether the revision is completed
	IsBeingReviewed     bool               `json:"isBeingReviewed" bson:"isBeingReviewed"`                     // Whether this video is currently being revised
	RevisedBy           primitive.ObjectID `json:"revisedBy,omitempty" bson:"revisedBy,omitempty"`             // Employee who did the revision
	RevisedByName       string             `json:"revisedByName,omitempty" bson:"revisedByName,omitempty"`     // Name of employee who did the revision
	ReviewedVideoID     primitive.ObjectID `json:"reviewedVideoId,omitempty" bson:"reviewedVideoId,omitempty"` // ID of the video being reviewed
	OriginalVideoID     primitive.ObjectID `json:"originalVideoId,omitempty" bson:"originalVideoId,omitempty"` // Add this field
	Reviews             []Review           `json:"reviews" bson:"reviews"`                                     // Reviews for this video
	StartTime           time.Time          `json:"startTime" bson:"startTime"`
	EndTime             time.Time          `json:"endTime,omitempty" bson:"endTime,omitempty"`
	Duration            string             `json:"duration,omitempty" bson:"duration,omitempty"`
	DurationMinutes     int                `json:"durationMinutes,omitempty" bson:"durationMinutes,omitempty"`
	Status              string             `json:"status" bson:"status"` // "in_progress" or "completed"
	RevisionStatus      string             `json:"revisionStatus,omitempty" bson:"revisionStatus,omitempty"`
	RevisionNote        string             `json:"revisionNote,omitempty" bson:"revisionNote,omitempty"`
	WorkStatus          string             `json:"workStatus" bson:"workStatus"`                   // "pending_review", "in_review", "needs_revision", "approved"
	AdminReviewed       bool               `json:"adminReviewed" bson:"adminReviewed"`             // Admin tarafından incelenip incelenmediği
	LastReviewType      string             `json:"lastReviewType" bson:"lastReviewType"`           // "admin" veya "employee"
	RevisionCount       int                `json:"revisionCount" bson:"revisionCount"`             // Kaç kez revize edildiği
	RevisionHistory     []RevisionRecord   `json:"revisionHistory" bson:"revisionHistory"`         // Revizyon geçmişi
	NeedsEmployeeReview bool               `json:"needsEmployeeReview" bson:"needsEmployeeReview"` // Personel incelemesi gerekiyor mu
	NeedsAdminReview    bool               `json:"needsAdminReview" bson:"needsAdminReview"`       // Admin incelemesi gerekiyor mu
	ReviewCycle         int                `json:"reviewCycle" bson:"reviewCycle"`                 // Kaçıncı inceleme döngüsünde
	LastReviewerType    string             `json:"lastReviewerType" bson:"lastReviewerType"`       // Son incelemeyi yapan (admin/employee)
}

type Review struct {
	ReviewerID   string    `json:"reviewerId" bson:"reviewerId"`
	ReviewerName string    `json:"reviewerName" bson:"reviewerName"`
	Comment      string    `json:"comment" bson:"comment"`
	CreatedAt    time.Time `json:"createdAt" bson:"createdAt"`
}

type RevisionRecord struct {
	ReviewerId   string    `json:"reviewerId" bson:"reviewerId"`     // İnceleme yapan kişi ID'si
	ReviewerName string    `json:"reviewerName" bson:"reviewerName"` // İnceleme yapan kişi adı
	ReviewType   string    `json:"reviewType" bson:"reviewType"`     // "admin" veya "employee"
	Comment      string    `json:"comment" bson:"comment"`           // İnceleme yorumu
	RevisionDate time.Time `json:"revisionDate" bson:"revisionDate"` // İnceleme tarihi
	Status       string    `json:"status" bson:"status"`             // "needs_revision" veya "approved"
}

func initMongoDB() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(os.Getenv("MONGODB_URI"))
	clientOptions.SetServerSelectionTimeout(5 * time.Second)

	var err error
	client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		return err
	}

	// Ping the database
	err = client.Ping(ctx, nil)
	if err != nil {
		return err
	}

	db = client.Database(os.Getenv("DB_NAME"))
	return nil
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Initialize MongoDB connection
	if err := initMongoDB(); err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		}
	}()

	// Initialize template engine
	engine := html.New("./templates", ".html")

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		Views: engine,
	})

	// Serve static files
	app.Static("/static", "./static")

	// Routes
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{})
	})

	app.Get("/employee", func(c *fiber.Ctx) error {
		return c.Render("employee", fiber.Map{})
	})

	app.Get("/admin", func(c *fiber.Ctx) error {
		return c.Render("admin", fiber.Map{})
	})

	// API Routes
	api := app.Group("/api")
	api.Post("/employees", createEmployee)
	api.Get("/employees", getEmployees)
	api.Delete("/employees/:id", deleteEmployee)
	api.Post("/work", createWork)
	api.Put("/work/:id", updateWork)
	api.Get("/works", getAllWorks)
	api.Get("/work/:id", getWork)
	api.Get("/work-stats/:employeeId", getEmployeeStats)
	api.Get("/daily-timeline", getDailyTimeline)
	api.Get("/approved-videos", getApprovedVideos)
	api.Get("/completed-videos", getCompletedVideos)
	api.Get("/reviewed-videos", getReviewedVideos)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(app.Listen(":" + port))
}

func createEmployee(c *fiber.Ctx) error {
	var employee Employee
	if err := c.BodyParser(&employee); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Validate employee type
	if employee.Type != "staff" && employee.Type != "intern" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid employee type. Must be either 'staff' or 'intern'",
			"type":  "warning",
			"title": "Uyarı",
			"text":  "Geçersiz personel tipi. Personel veya Stajyer seçiniz.",
		})
	}

	employee.ID = primitive.NewObjectID()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := db.Collection("employees").InsertOne(ctx, employee)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create employee: " + err.Error(),
			"type":  "error",
			"title": "Hata",
			"text":  "Personel/stajyer eklenirken bir hata oluştu.",
		})
	}

	employee.ID = result.InsertedID.(primitive.ObjectID)
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"type":  "success",
		"title": "Başarılı",
		"text":  "Personel/stajyer başarıyla eklendi.",
		"data":  employee,
	})
}

func getEmployees(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	includeDeleted := c.Query("includeDeleted") == "true"
	filter := bson.M{}
	if !includeDeleted {
		filter["deletedAt"] = bson.M{"$exists": false}
	}

	cursor, err := db.Collection("employees").Find(ctx, filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch employees: " + err.Error(),
			"type":  "warning",
			"title": "Uyarı",
			"text":  "Lütfen sisteme personel tanımlayınız.",
		})
	}
	defer cursor.Close(ctx)

	var employees []Employee
	if err = cursor.All(ctx, &employees); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to decode employees: " + err.Error(),
			"type":  "warning",
			"title": "Uyarı",
			"text":  "Lütfen sisteme personel tanımlayınız.",
		})
	}

	if len(employees) == 0 {
		return c.JSON(fiber.Map{
			"data":  []Employee{},
			"type":  "info",
			"title": "Bilgi",
			"text":  "Henüz personel tanımlı değil.",
		})
	}

	return c.JSON(fiber.Map{
		"data": employees,
		"type": "success",
	})
}

func deleteEmployee(c *fiber.Ctx) error {
	id, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID format"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()
	result, err := db.Collection("employees").UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"deletedAt": now}},
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete employee: " + err.Error()})
	}

	if result.ModifiedCount == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Employee not found"})
	}

	return c.JSON(fiber.Map{"message": "Employee deleted successfully"})
}

func getEmployeeStats(c *fiber.Ctx) error {
	employeeId, err := primitive.ObjectIDFromHex(c.Params("employeeId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid employee ID format"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Tüm tamamlanmış işleri getir
	completedWorks, err := db.Collection("works").Find(ctx, bson.M{
		"employeeId": employeeId,
		"status":     "completed",
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch works"})
	}
	defer completedWorks.Close(ctx)

	var works []Work
	if err = completedWorks.All(ctx, &works); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to decode works"})
	}

	var videoWorks []Work
	var softwareWorks []Work
	var revisionWorks []Work

	for _, work := range works {
		if isWeekend(work.StartTime) || isWeekend(work.EndTime) {
			continue // Haftasonu işleri hesaplamaya dahil etme
		}

		// Mesai saatlerine göre düzeltilmiş süreyi hesapla
		adjustedDuration := calculateWorkDuration(work.StartTime, work.EndTime)
		work.DurationMinutes = adjustedDuration

		switch work.WorkType {
		case "video":
			videoWorks = append(videoWorks, work)
		case "revize":
			revisionWorks = append(revisionWorks, work)
		case "software":
			softwareWorks = append(softwareWorks, work)
		}
	}

	stats := WorkStats{
		TotalWorks: len(videoWorks) + len(softwareWorks) + len(revisionWorks),
	}

	if len(videoWorks) > 0 {
		var totalMinutes int
		for _, work := range videoWorks {
			totalMinutes += work.DurationMinutes
		}
		avgMinutes := totalMinutes / len(videoWorks)
		stats.AverageVideoDuration = formatDuration(avgMinutes)
	}

	if len(softwareWorks) > 0 {
		var totalMinutes int
		for _, work := range softwareWorks {
			totalMinutes += work.DurationMinutes
		}
		avgMinutes := totalMinutes / len(softwareWorks)
		stats.AverageSoftwareDuration = formatDuration(avgMinutes)
	}

	if len(revisionWorks) > 0 {
		var totalMinutes int
		for _, work := range revisionWorks {
			totalMinutes += work.DurationMinutes
		}
		avgMinutes := totalMinutes / len(revisionWorks)
		stats.AverageRevisionDuration = formatDuration(avgMinutes)
	}

	return c.JSON(stats)
}

// Haftasonu kontrolü
func isWeekend(t time.Time) bool {
	weekday := t.Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}

// Mesai saatlerine göre çalışma süresini hesapla
func calculateWorkDuration(start, end time.Time) int {
	// Başlangıç ve bitiş zamanlarının aynı gün olup olmadığını kontrol et
	startDate := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	endDate := time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, end.Location())

	// Mesai saatleri
	workStartHour := 9
	workEndHour := 18
	lunchStartHour := 12
	lunchEndHour := 13

	var totalMinutes int

	if startDate.Equal(endDate) {
		// Aynı gün içindeki iş
		return calculateSingleDayDuration(start, end, workStartHour, workEndHour, lunchStartHour, lunchEndHour)
	} else {
		// İlk gün için süre hesaplama
		firstDayEnd := time.Date(start.Year(), start.Month(), start.Day(), workEndHour, 0, 0, 0, start.Location())
		totalMinutes += calculateSingleDayDuration(start, firstDayEnd, workStartHour, workEndHour, lunchStartHour, lunchEndHour)

		// Son gün için süre hesaplama
		lastDayStart := time.Date(end.Year(), end.Month(), end.Day(), workStartHour, 0, 0, 0, end.Location())
		totalMinutes += calculateSingleDayDuration(lastDayStart, end, workStartHour, workEndHour, lunchStartHour, lunchEndHour)

		// Aradaki tam günler için süre hesaplama (eğer varsa)
		currentDate := startDate.AddDate(0, 0, 1)
		for currentDate.Before(endDate) {
			if !isWeekend(currentDate) {
				// Tam gün mesai süresi: 8 saat (öğle arası hariç)
				totalMinutes += 8 * 60
			}
			currentDate = currentDate.AddDate(0, 0, 1)
		}
	}

	return totalMinutes
}

// Tek gün içindeki çalışma süresini hesapla
func calculateSingleDayDuration(start, end time.Time, workStartHour, workEndHour, lunchStartHour, lunchEndHour int) int {
	// Başlangıç saatini mesai başlangıcından önce ise mesai başlangıcına ayarla
	adjustedStart := start
	if start.Hour() < workStartHour {
		adjustedStart = time.Date(start.Year(), start.Month(), start.Day(), workStartHour, 0, 0, 0, start.Location())
	}

	// Bitiş saatini mesai bitişinden sonra ise mesai bitişine ayarla
	adjustedEnd := end
	if end.Hour() >= workEndHour {
		adjustedEnd = time.Date(end.Year(), end.Month(), end.Day(), workEndHour, 0, 0, 0, end.Location())
	}

	if adjustedEnd.Before(adjustedStart) {
		return 0
	}

	duration := 0
	currentTime := adjustedStart

	for currentTime.Before(adjustedEnd) {
		hour := currentTime.Hour()

		// Öğle arası kontrolü
		if hour >= lunchStartHour && hour < lunchEndHour {
			currentTime = time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(),
				lunchEndHour, 0, 0, 0, currentTime.Location())
			continue
		}

		if hour >= workStartHour && hour < workEndHour {
			// Her saat için dakika ekle
			nextHour := currentTime.Add(time.Hour)
			if nextHour.After(adjustedEnd) {
				// Son kısmi saat için dakika hesapla
				duration += int(adjustedEnd.Sub(currentTime).Minutes())
			} else {
				// Tam saat için 60 dakika ekle
				duration += 60
			}
		}

		currentTime = currentTime.Add(time.Hour)
	}

	return duration
}

func getDailyTimeline(c *fiber.Ctx) error {
	employeeID := c.Query("employeeId")
	if employeeID == "" {
		return c.JSON(fiber.Map{
			"type":  "warning",
			"title": "Uyarı",
			"text":  "Lütfen bir personel seçiniz.",
			"data":  []TimelineSlot{},
		})
	}

	dateStr := c.Query("date")
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"type":  "error",
			"title": "Hata",
			"text":  "Geçersiz tarih formatı",
		})
	}

	employeeObjID, err := primitive.ObjectIDFromHex(employeeID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"type":  "error",
			"title": "Hata",
			"text":  "Geçersiz personel ID formatı",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Önce çalışanı kontrol et
	var employee Employee
	err = db.Collection("employees").FindOne(ctx, bson.M{"_id": employeeObjID}).Decode(&employee)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"type":  "error",
				"title": "Hata",
				"text":  "Personel bulunamadı",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"type":  "error",
			"title": "Hata",
			"text":  "Personel bilgisi alınırken bir hata oluştu",
		})
	}

	// Eğer personel silinmişse ve seçilen tarih silinme tarihinden sonraysa timeline'ı gösterme
	if employee.DeletedAt != nil && date.After(employee.DeletedAt.Truncate(24*time.Hour)) {
		return c.JSON(fiber.Map{
			"type": "success",
			"data": []TimelineSlot{},
		})
	}

	filter := bson.M{
		"startTime": bson.M{
			"$gte": time.Date(date.Year(), date.Month(), date.Day(), 9, 0, 0, 0, time.Local),
			"$lte": time.Date(date.Year(), date.Month(), date.Day(), 18, 0, 0, 0, time.Local),
		},
		"employeeId": employeeObjID,
	}

	works, err := db.Collection("works").Find(ctx, filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"type":  "error",
			"title": "Hata",
			"text":  "İşler yüklenirken bir hata oluştu",
		})
	}
	defer works.Close(ctx)

	var dailyWorks []Work
	if err = works.All(ctx, &dailyWorks); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"type":  "error",
			"title": "Hata",
			"text":  "İşler yüklenirken bir hata oluştu",
		})
	}

	timeline := make([]TimelineSlot, 10)
	for i := 0; i < 10; i++ {
		timeline[i] = TimelineSlot{
			Hour:  i + 9,
			Works: []Work{},
		}
	}

	for _, work := range dailyWorks {
		if work.EmployeeID == employeeObjID {
			hour := work.StartTime.Hour()
			if hour >= 9 && hour < 18 {
				slotIndex := hour - 9
				timeline[slotIndex].Works = append(timeline[slotIndex].Works, work)
			}
		}
	}

	return c.JSON(fiber.Map{
		"type": "success",
		"data": timeline,
	})
}

func createWork(c *fiber.Ctx) error {
	var work Work
	if err := c.BodyParser(&work); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	work.ID = primitive.NewObjectID()
	work.Status = "in_progress"

	// Set OriginalVideoID for revisions
	if work.WorkType == "revize" && !work.ReviewedVideoID.IsZero() {
		// Get original video's information
		var originalVideo Work
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := db.Collection("works").FindOne(ctx, bson.M{"_id": work.ReviewedVideoID}).Decode(&originalVideo)
		if err == nil {
			// Only copy reviews if the original video has a needs_revision status
			if originalVideo.RevisionStatus == "needs_revision" {
				work.Reviews = originalVideo.Reviews
				work.OriginalVideoID = work.ReviewedVideoID
			} else {
				// If the video doesn't need revision, return an error
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Bu video için revizyon gerekmiyor",
					"type":  "warning",
					"title": "Uyarı",
					"text":  "Bu video için henüz revizyon talebi oluşturulmamış.",
				})
			}
		}
	}

	if work.WorkType == "video" {
		work.WorkStatus = "pending_review"
		work.RevisionCount = 0
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := db.Collection("works").InsertOne(ctx, work)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create work: " + err.Error()})
	}

	work.ID = result.InsertedID.(primitive.ObjectID)
	return c.Status(fiber.StatusCreated).JSON(work)
}

func updateWork(c *fiber.Ctx) error {
	id, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID format"})
	}

	var update struct {
		EndTime             time.Time `json:"endTime"`
		VideoLink           string    `json:"videoLink"`
		Description         string    `json:"description"`
		RevisionStatus      string    `json:"revisionStatus"`
		RevisionNote        string    `json:"revisionNote"`
		Status              string    `json:"status"`
		Reviews             []Review  `json:"reviews"`
		IsBeingReviewed     bool      `json:"isBeingReviewed"`
		IsRevisionCompleted bool      `json:"isRevisionCompleted"`
		WorkType            string    `json:"workType"`
		AdminReviewed       bool      `json:"adminReviewed"`
		WorkStatus          string    `json:"workStatus"`
	}
	if err := c.BodyParser(&update); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var work Work
	err = db.Collection("works").FindOne(ctx, bson.M{"_id": id}).Decode(&work)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Work not found"})
	}

	updateFields := bson.M{}
	if !update.EndTime.IsZero() {
		duration := update.EndTime.Sub(work.StartTime)
		durationMinutes := int(duration.Minutes())
		updateFields["endTime"] = update.EndTime
		updateFields["duration"] = duration.String()
		updateFields["durationMinutes"] = durationMinutes
	}
	if update.VideoLink != "" {
		updateFields["videoLink"] = update.VideoLink
	}
	if update.Description != "" {
		updateFields["description"] = update.Description
	}
	if update.RevisionStatus != "" {
		updateFields["revisionStatus"] = update.RevisionStatus
	}
	if update.RevisionNote != "" {
		updateFields["revisionNote"] = update.RevisionNote
		// Revizyon notu eklendiğinde otomatik olarak bir review oluştur
		review := Review{
			ReviewerID:   "admin",
			ReviewerName: "Admin",
			Comment:      update.RevisionNote,
			CreatedAt:    time.Now(),
		}
		if work.Reviews == nil {
			work.Reviews = []Review{}
		}
		work.Reviews = append(work.Reviews, review)
		updateFields["reviews"] = work.Reviews
	}
	if update.Status != "" {
		updateFields["status"] = update.Status
	}
	if len(update.Reviews) > 0 {
		updateFields["reviews"] = update.Reviews
	}
	if update.IsBeingReviewed {
		updateFields["isBeingReviewed"] = true
	}
	if update.IsRevisionCompleted {
		updateFields["isRevisionCompleted"] = true
	}
	if update.WorkType != "" {
		updateFields["workType"] = update.WorkType
	}

	if work.WorkType == "video" || work.WorkType == "revize" {
		// Video tamamlandığında
		if update.Status == "completed" && work.Status != "completed" {
			updateFields["workStatus"] = "pending_review"
		}

		// Admin incelemesi tamamlandığında
		if update.AdminReviewed {
			if update.WorkStatus == "needs_revision" {
				// Revizyon gerekiyorsa
				record := RevisionRecord{
					ReviewerId:   "admin",
					ReviewerName: "Admin",
					ReviewType:   "admin",
					Comment:      update.RevisionNote,
					RevisionDate: time.Now(),
					Status:       "needs_revision",
				}
				updateFields["revisionHistory"] = append(work.RevisionHistory, record)
				updateFields["revisionCount"] = work.RevisionCount + 1
			} else if update.WorkStatus == "approved" {
				// Video onaylandıysa
				record := RevisionRecord{
					ReviewerId:   "admin",
					ReviewerName: "Admin",
					ReviewType:   "admin",
					Comment:      "Video onaylandı",
					RevisionDate: time.Now(),
					Status:       "approved",
				}
				updateFields["revisionHistory"] = append(work.RevisionHistory, record)
				updateFields["workStatus"] = "approved"
			}
		}

		// Personel incelemesi tamamlandığında
		if len(update.Reviews) > len(work.Reviews) {
			lastReview := update.Reviews[len(update.Reviews)-1]
			record := RevisionRecord{
				ReviewerId:   lastReview.ReviewerID,
				ReviewerName: lastReview.ReviewerName,
				ReviewType:   "employee",
				Comment:      lastReview.Comment,
				RevisionDate: lastReview.CreatedAt,
				Status:       "needs_revision",
			}
			updateFields["revisionHistory"] = append(work.RevisionHistory, record)
			updateFields["revisionCount"] = work.RevisionCount + 1
			updateFields["workStatus"] = "needs_revision"
		}
	}

	result, err := db.Collection("works").UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": updateFields},
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update work: " + err.Error()})
	}

	if result.ModifiedCount == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Work not found"})
	}

	return c.JSON(fiber.Map{"message": "Work updated successfully"})
}

func getAllWorks(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := db.Collection("works").Find(ctx, bson.M{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch works: " + err.Error()})
	}
	defer cursor.Close(ctx)

	var works []Work
	if err = cursor.All(ctx, &works); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to decode works: " + err.Error()})
	}

	return c.JSON(works)
}

func getWork(c *fiber.Ctx) error {
	id, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID format"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var work Work
	err = db.Collection("works").FindOne(ctx, bson.M{"_id": id}).Decode(&work)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Work not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch work: " + err.Error()})
	}

	return c.JSON(work)
}

func getApprovedVideos(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := db.Collection("works").Find(ctx, bson.M{
		"workType":       "video",
		"status":         "completed",
		"revisionStatus": "approved",
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"type":  "error",
			"title": "Hata",
			"text":  "Onaylanmış videolar yüklenirken bir hata oluştu",
			"data":  []Work{},
		})
	}
	defer cursor.Close(ctx)

	var videos []Work
	if err = cursor.All(ctx, &videos); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"type":  "error",
			"title": "Hata",
			"text":  "Onaylanmış videolar yüklenirken bir hata oluştu",
			"data":  []Work{},
		})
	}

	if len(videos) == 0 {
		return c.JSON(fiber.Map{
			"type":  "info",
			"title": "Bilgi",
			"text":  "Henüz onaylanmış video bulunmuyor",
			"data":  []Work{},
		})
	}

	return c.JSON(fiber.Map{
		"type": "success",
		"data": videos,
	})
}

func getCompletedVideos(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dateStr := c.Query("date")
	var startTime, endTime time.Time

	if dateStr != "" {
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"type":  "error",
				"title": "Hata",
				"text":  "Geçersiz tarih formatı",
				"data":  []Work{},
			})
		}

		startTime = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)
		endTime = time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 999999999, time.Local)
	}

	// Ana video işleri için filtre
	filter := bson.M{
		"workType": "video",
		"status":   "completed",
		"revisionStatus": bson.M{
			"$nin": []string{"approved"}, // Onaylanmamış olanlar
		},
	}

	if dateStr != "" {
		filter["endTime"] = bson.M{
			"$gte": startTime,
			"$lte": endTime,
		}
	}

	// Ana videoları bul
	cursor, err := db.Collection("works").Find(ctx, filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"type":  "error",
			"title": "Hata",
			"text":  "Videolar yüklenirken bir hata oluştu",
			"data":  []Work{},
		})
	}
	defer cursor.Close(ctx)

	var videos []Work
	if err = cursor.All(ctx, &videos); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"type": "error",
			"text": "Videolar yüklenirken bir hata oluştu",
			"data": []Work{},
		})
	}

	// Her video için revizyon işlerini bul
	for i := range videos {
		revisionFilter := bson.M{
			"workType":        "revize",
			"reviewedVideoId": videos[i].ID,
		}

		revisionCursor, err := db.Collection("works").Find(ctx, revisionFilter)
		if err != nil {
			continue
		}
		defer revisionCursor.Close(ctx)

		var revisions []Work
		if err = revisionCursor.All(ctx, &revisions); err != nil {
			continue
		}

		// Revizyonları ana video ile birleştir
		if len(revisions) > 0 {
			videos[i].RevisionHistory = []RevisionRecord{}
			for _, rev := range revisions {
				record := RevisionRecord{
					ReviewerId:   rev.EmployeeID.Hex(),
					ReviewerName: rev.EmployeeName,
					ReviewType:   "revize",
					Comment:      rev.Description,
					RevisionDate: rev.StartTime,
					Status:       rev.Status,
				}
				videos[i].RevisionHistory = append(videos[i].RevisionHistory, record)
			}
		}
	}

	if len(videos) == 0 {
		return c.JSON(fiber.Map{
			"type":  "info",
			"title": "Bilgi",
			"text":  "İncelenecek video bulunmuyor",
			"data":  []Work{},
		})
	}

	return c.JSON(fiber.Map{
		"type": "success",
		"data": videos,
	})
}

func getReviewedVideos(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := db.Collection("works").Find(ctx, bson.M{
		"workType": "video",
		"status":   "completed",
		"reviews":  bson.M{"$exists": true, "$ne": []interface{}{}},
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"type":  "error",
			"title": "Hata",
			"text":  "İncelenmiş videolar yüklenirken bir hata oluştu",
			"data":  []Work{},
		})
	}
	defer cursor.Close(ctx)

	var videos []Work
	if err = cursor.All(ctx, &videos); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"type":  "error",
			"title": "Hata",
			"text":  "İncelenmiş videolar yüklenirken bir hata oluştu",
			"data":  []Work{},
		})
	}

	if len(videos) == 0 {
		return c.JSON(fiber.Map{
			"type":  "info",
			"title": "Bilgi",
			"text":  "Henüz incelenmiş video bulunmuyor",
			"data":  []Work{},
		})
	}

	return c.JSON(fiber.Map{
		"type": "success",
		"data": videos,
	})
}

func formatDuration(minutes int) string {
	hours := minutes / 60
	mins := minutes % 60
	if hours > 0 {
		return fmt.Sprintf("%ds %ddk", hours, mins)
	}
	return fmt.Sprintf("%ddk", mins)
}
