package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"project-x/config"
	"project-x/models"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Connect to database
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto migrate to ensure tables exist
	err = db.AutoMigrate(&models.User{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	fmt.Println("=== Admin User Creation Utility ===")
	fmt.Println()

	// Get user input
	username := getInput("Enter username: ")
	if username == "" {
		log.Fatal("Username cannot be empty")
	}

	// Check if user already exists
	var existingUser models.User
	if err := db.Where("username = ?", username).First(&existingUser).Error; err == nil {
		log.Fatal("User with this username already exists")
	}

	// Get password
	password := getInput("Enter password: ")
	if password == "" {
		log.Fatal("Password cannot be empty")
	}

	// Confirm password
	confirmPassword := getInput("Confirm password: ")
	if password != confirmPassword {
		log.Fatal("Passwords do not match")
	}

	// Get department
	department := getInput("Enter department: ")
	if department == "" {
		log.Fatal("Department cannot be empty")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Failed to hash password:", err)
	}

	// Create admin user
	adminUser := models.User{
		Username:   username,
		Password:   string(hashedPassword),
		Role:       models.RoleAdmin,
		Department: department,
	}

	// Save to database
	if err := db.Create(&adminUser).Error; err != nil {
		log.Fatal("Failed to create admin user:", err)
	}

	fmt.Println()
	fmt.Println("✅ Admin user created successfully!")
	fmt.Printf("Username: %s\n", adminUser.Username)
	fmt.Printf("Role: %s\n", adminUser.Role)
	fmt.Printf("Department: %s\n", adminUser.Department)
	fmt.Printf("User ID: %d\n", adminUser.ID)
}

func getInput(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal("Failed to read input:", err)
	}
	return strings.TrimSpace(input)
}
