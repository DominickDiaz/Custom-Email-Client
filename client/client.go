package client

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	auth "github.com/microsoft/kiota-authentication-azure-go"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/me"
	"github.com/microsoftgraph/msgraph-sdk-go/me/mailfolders/item/messages"
	"github.com/microsoftgraph/msgraph-sdk-go/me/sendmail"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

type Client struct {
	deviceCodeCredential   *azidentity.DeviceCodeCredential
	userClient             *msgraphsdk.GraphServiceClient
	graphUserScopes        []string
	clientSecretCredential *azidentity.ClientSecretCredential
}

func NewClient() *Client {
	user := &Client{}
	user.InitializeClient()
	return user
}

func (user *Client) InitializeClient() error {
	clientId := os.Getenv("CLIENT_ID")
	authTenant := os.Getenv("AUTH_TENANT")
	scopes := os.Getenv("GRAPH_USER_SCOPES")
	user.graphUserScopes = strings.Split(scopes, ",")

	// Create the device code credential
	log.Print("Here 1")
	credential, err := azidentity.NewDeviceCodeCredential(&azidentity.DeviceCodeCredentialOptions{
		ClientID: clientId,
		TenantID: authTenant,
		UserPrompt: func(ctx context.Context, message azidentity.DeviceCodeMessage) error {
			fmt.Println(message.Message)
			return nil
		},
	})
	if err != nil {
		return err
	}
	log.Print("Here 2")
	user.deviceCodeCredential = credential

	// Create an auth provider using the credential
	log.Print("Here 3")
	authProvider, err := auth.NewAzureIdentityAuthenticationProviderWithScopes(credential, user.graphUserScopes)
	if err != nil {
		return err
	}

	// Create a request adapter using the auth provider
	adapter, err := msgraphsdk.NewGraphRequestAdapter(authProvider)
	if err != nil {
		return err
	}

	// Create a Graph client using request adapter
	client := msgraphsdk.NewGraphServiceClient(adapter)
	user.userClient = client

	return nil
}

func (users *Client) Greeting() {
	user, err := users.GetUser()
	if err != nil {
		log.Panicf("Error getting user: %v\n", err)
	}

	fmt.Printf("Hello, %s!\n", *user.GetDisplayName())

	email := user.GetMail() // For Work/school accounts, email is in Mail property
	if email == nil {       // Personal accounts, email is in UserPrincipalName
		email = user.GetUserPrincipalName()
	}

	fmt.Printf("Email: %s\n", *email)
	fmt.Println()
}

func (g *Client) GetInbox() (models.MessageCollectionResponseable, error) {
	var topValue int32 = 25
	query := messages.MessagesRequestBuilderGetQueryParameters{
		// Only request specific properties
		Select: []string{"from", "isRead", "receivedDateTime", "subject"},
		// Get at most 25 results
		Top: &topValue,
		// Sort by received time, newest first
		Orderby: []string{"receivedDateTime DESC"},
	}

	return g.userClient.Me().
		MailFoldersById("inbox").
		Messages().
		GetWithRequestConfigurationAndResponseHandler(
			&messages.MessagesRequestBuilderGetRequestConfiguration{
				QueryParameters: &query,
			},
			nil)
}

func (Client *Client) ListInbox() {
	messages, err := Client.GetInbox()
	if err != nil {
		log.Panicf("Error getting user's inbox: %v", err)
	}
	location, err := time.LoadLocation("Local")
	if err != nil {
		log.Panicf("Error getting local timezone: %v", err)
	}

	// Output each message's details
	for _, message := range messages.GetValue() {
		fmt.Printf("Message: %s\n", *message.GetSubject())
		fmt.Printf("  From: %s\n", *message.GetFrom().GetEmailAddress().GetName())

		status := "Unknown"
		if *message.GetIsRead() {
			status = "Read"
		} else {
			status = "Unread"
		}
		fmt.Printf("  Status: %s\n", status)
		fmt.Printf("  Received: %s\n", (*message.GetReceivedDateTime()).In(location))
	}

	// there are more messages available on the server
	nextLink := messages.GetOdatanextLink()

	fmt.Println()
	fmt.Printf("More messages available? %t\n", nextLink != nil)
	fmt.Println()
}

func (client *Client) SendMail(title *string, contents *string, address *string) {
	// Get the user for their email address
	user, err := client.GetUser()
	if err != nil {
		log.Panicf("Error getting user: %v", err)
	}

	email := user.GetMail()
	if email == nil {
		email = user.GetUserPrincipalName()
	}


	
	client.SendMailHelper(title, contents, address)

	fmt.Println("Mail sent.")
	fmt.Println()
}

func (user *Client) SendMailHelper(title *string, contents *string, address *string) error {
	// Create a new message
	message := models.NewMessage()
	message.SetSubject(title)

	messageBody := models.NewItemBody()
	messageBody.SetContent(contents)
	contentType := models.TEXT_BODYTYPE
	messageBody.SetContentType(&contentType)
	message.SetBody(messageBody)

	toRecipient := models.NewRecipient()
	emailAddress := models.NewEmailAddress()
	emailAddress.SetAddress(address)
	toRecipient.SetEmailAddress(emailAddress)
	message.SetToRecipients([]models.Recipientable{
		toRecipient,
	})

	sendMailBody := sendmail.NewSendMailRequestBody()
	sendMailBody.SetMessage(message)

	// Send the message
	return user.userClient.Me().SendMail().Post(sendMailBody)
}


func (g *Client) EnsureGraphForAppOnlyAuth() error {
	if g.clientSecretCredential == nil {
		clientId := os.Getenv("CLIENT_ID")
		tenantId := os.Getenv("TENANT_ID")
		clientSecret := os.Getenv("CLIENT_SECRET")
		credential, err := azidentity.NewClientSecretCredential(tenantId, clientId, clientSecret, nil)
		if err != nil {
			return err
		}

		g.clientSecretCredential = credential
	}

	if g.userClient == nil {
		// Create an auth provider using the credential
		authProvider, err := auth.NewAzureIdentityAuthenticationProviderWithScopes(g.clientSecretCredential, []string{
			"https://graph.microsoft.com/.default",
		})

		// Create a request adapter using the auth provider
		adapter, err := msgraphsdk.NewGraphRequestAdapter(authProvider)
		if err != nil {
			return err
		}

		// Create a Graph client using request adapter
		client := msgraphsdk.NewGraphServiceClient(adapter)
		g.userClient = client
	}

	return nil
}

func (g *Client) GetUser() (models.Userable, error) {
	query := me.MeRequestBuilderGetQueryParameters{
		// Only request specific properties
		Select: []string{"displayName", "mail", "userPrincipalName"},
	}

	return g.userClient.Me().
		GetWithRequestConfigurationAndResponseHandler(
			&me.MeRequestBuilderGetRequestConfiguration{
				QueryParameters: &query,
			},
			nil)
}

func (g *Client) GetUserToken() (*string, error) {
	token, err := g.deviceCodeCredential.GetToken(context.Background(), policy.TokenRequestOptions{
		Scopes: g.graphUserScopes,
	})
	if err != nil {
		return nil, err
	}

	return &token.Token, nil
}

func (Client *Client) DisplayAccessToken() {
	token, err := Client.GetUserToken()
	if err != nil {
		log.Panicf("Error getting user token: %v\n", err)
	}

	fmt.Printf("User token: %s", *token)
	fmt.Println()
}
