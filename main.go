package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const dataFile = "data.json"

type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Task struct {
	ID         int        `json:"id"`
	Name       string     `json:"name"`
	Date       time.Time  `json:"date"`
	Status     TaskStatus `json:"status"`
	UserID     int        `json:"user_id"`
	CategoryID int        `json:"category_id"`
}

type Category struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	UserID int    `json:"user_id"`
}

type Storage struct {
	Users      []User     `json:"users"`
	Tasks      []Task     `json:"tasks"`
	Categories []Category `json:"categories"`
}

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "inProgress"
	TaskStatusCompleted  TaskStatus = "completed"
)

var storage Storage

var loggedInUser User // session

func main() {
	fmt.Println("Hello todo app")

	loadData()

	command := flag.String("c", "no-command", "Command")
	flag.Parse()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		if *command != "" {
			runCommand(*command)
		} else {
			fmt.Println("Empty command. Please enter a valid command.")
		}

		fmt.Println("Please enter another command:")

		if !scanner.Scan() {
			fmt.Println("Failed to read input")
			continue
		}

		input := scanner.Text()

		if input == "exit" {
			fmt.Println("Goodbye!")
			break
		}

		*command = input
	}
}

type Commands map[string]func()

func runCommand(command string) {

	if command != "register" && command != "exit" && loggedInUser == (User{}) {
		login()
		return
	}

	commands := Commands{
		"create-task":     createTask,
		"create-category": createCategory,
		"register":        register,
		"login":           login,
		"list-tasks":      listTasks,
	}

	call, ok := commands[command]
	if !ok {
		fmt.Errorf("unknown command: %s", command)
		return
	}

	call()
}

func createTask() {
	name := scanInput("Enter your task name: ")
	date := scanInput("Enter your task date: ")

	userCategories := getUserCategories()

	for _, category := range userCategories {
		fmt.Printf("%d. %s\n", category.ID, category.Name)
	}

	categoryIDStr := scanInput("Enter your task category ID: ")

	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		parsedDate, err = time.Parse(time.RFC3339, date)
		if err != nil {
			fmt.Println("Invalid date format. Use YYYY-MM-DD (e.g., 2025-12-02) or RFC3339 (e.g., 2025-12-02T10:00:00Z)")
			return
		}
	}

	categoryID, err := strconv.Atoi(categoryIDStr)
	if err != nil {
		fmt.Println("Invalid category ID. Please enter a number.")
		return
	}

	task := Task{
		ID:         len(storage.Tasks) + 1,
		Name:       name,
		Date:       parsedDate,
		CategoryID: categoryID,
		Status:     TaskStatusPending,
		UserID:     loggedInUser.ID,
	}

	storage.Tasks = append(storage.Tasks, task)
	saveData()
}

func createCategory() {
	name := scanInput("Enter your category name: ")

	category := Category{
		ID:     len(storage.Categories) + 1,
		Name:   name,
		UserID: loggedInUser.ID,
	}

	storage.Categories = append(storage.Categories, category)
	saveData()

	fmt.Printf("Category created: %s\n", name)
}

func login() {
	email := scanInput("Enter your email: ")
	password := scanInput("Enter your password: ")

	loggedIn := false
	for _, user := range storage.Users {
		if user.Email == email && bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) == nil {
			println("You have been logged in!")
			loggedIn = true
			loggedInUser = user
		}
	}

	if !loggedIn {
		fmt.Println("Wrong credentials!")
		os.Exit(1)
	}
	fmt.Printf("Logged in as: %s\n", email)
}

func register() {
	name := scanInput("Enter your name: ")
	email := scanInput("Enter your email: ")
	password := scanInput("Enter your password: ")

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("Error hashing password:", err)
		return
	}

	user := User{
		ID:       len(storage.Users) + 1,
		Name:     name,
		Email:    email,
		Password: string(hashedPassword),
	}

	storage.Users = append(storage.Users, user)
	loggedInUser = user
	saveData()

	fmt.Println("User registered successfully!")
}

func listTasks() {
	for _, task := range storage.Tasks {
		if task.UserID == loggedInUser.ID {
			fmt.Printf("ID: %d, Name: %s, Date: %s, Status: %s\n",
				task.ID, task.Name, task.Date.Format("2006-01-02"), task.Status)
		}
	}
}

func scanInput(prompt string) string {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print(prompt)
	scanner.Scan()
	return scanner.Text()
}

func getUserCategories() []Category {
	var categories []Category
	for _, category := range storage.Categories {
		if category.UserID == loggedInUser.ID {
			categories = append(categories, category)
		}
	}
	return categories
}

func loadData() {
	file, err := os.ReadFile(dataFile)
	if err != nil {
		// File doesn't exist, start with empty storage
		storage = Storage{
			Users:      []User{},
			Tasks:      []Task{},
			Categories: []Category{},
		}
		return
	}

	err = json.Unmarshal(file, &storage)
	if err != nil {
		fmt.Println("Error loading data:", err)
		storage = Storage{
			Users:      []User{},
			Tasks:      []Task{},
			Categories: []Category{},
		}
	}
}

func saveData() {
	data, err := json.MarshalIndent(storage, "", "  ")
	if err != nil {
		fmt.Println("Error saving data:", err)
		return
	}

	err = os.WriteFile(dataFile, data, 0644)
	if err != nil {
		fmt.Println("Error writing data file:", err)
	}
}
