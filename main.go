package main

import (
	"fmt"
	"log"
	"github.com/joho/godotenv"
	"graphtesting/client"
)

func main() {
	fmt.Println("Go Mail Experimental Program")
	fmt.Println()
	log.Println("Go Mail")


	godotenv.Load("env")
	err := godotenv.Load()

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
			fmt.Println("Entering Email Sender...")
		case 3:
			// Send an email message
			MailSender(Client)
		default:
			fmt.Println("Invalid choice! Please try again.")
		}
		if choice == 2 {
			MailSender(Client)
		}

		if choice == 0 {
			break
		}
	}

}


func MailSender(Client *client.Client){
	subject := ""
	address := ""
	body := ""

	fmt.Println("Enter Recipient Email Address:")
	fmt.Scanf("%s", address)
	
	fmt.Println("Enter Subject:")
	fmt.Scanf("%s", subject)

	fmt.Println("Enter Recipient Email Address")
	fmt.Scanf("%s", body)
	Client.SendMail(&subject, &body, &address)
}