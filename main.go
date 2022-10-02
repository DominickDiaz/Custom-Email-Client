package main

import (
	"fmt"
	"graphtutorial/client"
	"log"
	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("Go Mail Experimental Program")
	fmt.Println()

	godotenv.Load(".env.local")
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error, Check .env file.")
	}

	Client := client.NewClient()
	Client.Greeting()

	var choice int64 = -1

	for {
		fmt.Println("Enter one of the following options:")
		fmt.Println("1. Display access token")
		fmt.Println("2. List my inbox")
		fmt.Println("3. Send mail")
		fmt.Println("0. Exit")

		_, err = fmt.Scanf("%d", &choice)
		if err != nil {
			choice = -1
		}

		switch choice {
		case 0:
			// Exit the program
			fmt.Println("Goodbye...")
		case 1:
			// Display access token
			Client.DisplayAccessToken()
		case 2:
			// List emails from user's inbox
			Client.ListInbox()
		case 3:
			// Send an email message
			Client.SendMail2()
		default:
			fmt.Println("Invalid choice! Please try again.")
		}
		if choice == 0 {
			break
		}
	}
}
