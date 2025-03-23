package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

// Scenario defines a load test scenario
type Scenario interface {
	// Setup initializes the scenario for a virtual user
	Setup(ctx context.Context, baseURL string, userID int) error

	// Execute executes a single iteration of the scenario
	Execute(ctx context.Context) (*ScenarioResult, error)

	// Teardown cleans up resources after the test
	Teardown(ctx context.Context) error

	// Name returns the scenario name
	Name() string
}

// ScenarioResult represents the result of a scenario execution
type ScenarioResult struct {
	Success      bool
	ResponseTime time.Duration
	StatusCode   int
	Error        error
}

// GetScenario returns a scenario by name
func GetScenario(name string) (Scenario, error) {
	switch name {
	case "messages":
		return NewMessagesScenario(), nil
	case "contacts":
		return NewContactsScenario(), nil
	case "mixed":
		return NewMixedScenario(), nil
	default:
		return nil, fmt.Errorf("unknown scenario: %s", name)
	}
}

// MessagesScenario simulates users sending and receiving messages
type MessagesScenario struct {
	baseURL     string
	userID      int
	accessToken string
	contacts    []string
	client      *http.Client
}

// NewMessagesScenario creates a new messages scenario
func NewMessagesScenario() *MessagesScenario {
	return &MessagesScenario{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Setup initializes the messages scenario
func (s *MessagesScenario) Setup(ctx context.Context, baseURL string, userID int) error {
	s.baseURL = baseURL
	s.userID = userID

	// In a real implementation, this would:
	// 1. Register or log in a test user
	// 2. Get an access token
	// 3. Create or retrieve contacts to message

	s.accessToken = "dummy-token"
	s.contacts = []string{"contact1", "contact2", "contact3"}

	return nil
}

// Execute runs one iteration of the messages scenario
func (s *MessagesScenario) Execute(ctx context.Context) (*ScenarioResult, error) {
	start := time.Now()

	// Randomly select an action: send message, get messages, or get conversation
	action := rand.Intn(3)

	var statusCode int
	var err error

	switch action {
	case 0:
		// Simulate sending a message
		statusCode, err = s.sendMessage()
	case 1:
		// Simulate getting messages
		statusCode, err = s.getMessages()
	case 2:
		// Simulate getting a conversation
		statusCode, err = s.getConversation()
	}

	duration := time.Since(start)

	return &ScenarioResult{
		Success:      err == nil && statusCode >= 200 && statusCode < 300,
		ResponseTime: duration,
		StatusCode:   statusCode,
		Error:        err,
	}, nil
}

// sendMessage simulates sending a message
func (s *MessagesScenario) sendMessage() (int, error) {
	// In a real implementation, this would make an API request to send a message
	time.Sleep(time.Duration(50+rand.Intn(100)) * time.Millisecond)
	return 201, nil
}

// getMessages simulates getting messages
func (s *MessagesScenario) getMessages() (int, error) {
	// In a real implementation, this would make an API request to get messages
	time.Sleep(time.Duration(30+rand.Intn(70)) * time.Millisecond)
	return 200, nil
}

// getConversation simulates getting a conversation
func (s *MessagesScenario) getConversation() (int, error) {
	// In a real implementation, this would make an API request to get a conversation
	time.Sleep(time.Duration(40+rand.Intn(80)) * time.Millisecond)
	return 200, nil
}

// Teardown cleans up resources
func (s *MessagesScenario) Teardown(ctx context.Context) error {
	// In a real implementation, this might log out or clean up test data
	return nil
}

// Name returns the scenario name
func (s *MessagesScenario) Name() string {
	return "messages"
}

// ContactsScenario simulates users managing contacts
type ContactsScenario struct {
	baseURL     string
	userID      int
	accessToken string
	client      *http.Client
}

// NewContactsScenario creates a new contacts scenario
func NewContactsScenario() *ContactsScenario {
	return &ContactsScenario{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Setup initializes the contacts scenario
func (s *ContactsScenario) Setup(ctx context.Context, baseURL string, userID int) error {
	s.baseURL = baseURL
	s.userID = userID

	// In a real implementation, this would register/login and get an access token
	s.accessToken = "dummy-token"

	return nil
}

// Execute runs one iteration of the contacts scenario
func (s *ContactsScenario) Execute(ctx context.Context) (*ScenarioResult, error) {
	start := time.Now()

	// Randomly select an action: add contact, get contacts, update contact, or delete contact
	action := rand.Intn(4)

	var statusCode int
	var err error

	switch action {
	case 0:
		// Simulate adding a contact
		statusCode, err = s.addContact()
	case 1:
		// Simulate getting contacts
		statusCode, err = s.getContacts()
	case 2:
		// Simulate updating a contact
		statusCode, err = s.updateContact()
	case 3:
		// Simulate deleting a contact
		statusCode, err = s.deleteContact()
	}

	duration := time.Since(start)

	return &ScenarioResult{
		Success:      err == nil && statusCode >= 200 && statusCode < 300,
		ResponseTime: duration,
		StatusCode:   statusCode,
		Error:        err,
	}, nil
}

// addContact simulates adding a contact
func (s *ContactsScenario) addContact() (int, error) {
	// In a real implementation, this would make an API request to add a contact
	time.Sleep(time.Duration(40+rand.Intn(60)) * time.Millisecond)
	return 201, nil
}

// getContacts simulates getting contacts
func (s *ContactsScenario) getContacts() (int, error) {
	// In a real implementation, this would make an API request to get contacts
	time.Sleep(time.Duration(20+rand.Intn(40)) * time.Millisecond)
	return 200, nil
}

// updateContact simulates updating a contact
func (s *ContactsScenario) updateContact() (int, error) {
	// In a real implementation, this would make an API request to update a contact
	time.Sleep(time.Duration(30+rand.Intn(50)) * time.Millisecond)
	return 200, nil
}

// deleteContact simulates deleting a contact
func (s *ContactsScenario) deleteContact() (int, error) {
	// In a real implementation, this would make an API request to delete a contact
	time.Sleep(time.Duration(30+rand.Intn(50)) * time.Millisecond)
	return 200, nil
}

// Teardown cleans up resources
func (s *ContactsScenario) Teardown(ctx context.Context) error {
	// In a real implementation, this might log out or clean up test data
	return nil
}

// Name returns the scenario name
func (s *ContactsScenario) Name() string {
	return "contacts"
}

// MixedScenario combines messages and contacts scenarios
type MixedScenario struct {
	messagesScenario *MessagesScenario
	contactsScenario *ContactsScenario
	baseURL          string
	userID           int
}

// NewMixedScenario creates a new mixed scenario
func NewMixedScenario() *MixedScenario {
	return &MixedScenario{
		messagesScenario: NewMessagesScenario(),
		contactsScenario: NewContactsScenario(),
	}
}

// Setup initializes the mixed scenario
func (s *MixedScenario) Setup(ctx context.Context, baseURL string, userID int) error {
	s.baseURL = baseURL
	s.userID = userID

	// Set up both scenarios
	if err := s.messagesScenario.Setup(ctx, baseURL, userID); err != nil {
		return err
	}

	if err := s.contactsScenario.Setup(ctx, baseURL, userID); err != nil {
		return err
	}

	return nil
}

// Execute runs one iteration of the mixed scenario
func (s *MixedScenario) Execute(ctx context.Context) (*ScenarioResult, error) {
	// Randomly choose between messages and contacts scenarios
	if rand.Intn(2) == 0 {
		return s.messagesScenario.Execute(ctx)
	}
	return s.contactsScenario.Execute(ctx)
}

// Teardown cleans up resources
func (s *MixedScenario) Teardown(ctx context.Context) error {
	// Tear down both scenarios
	if err := s.messagesScenario.Teardown(ctx); err != nil {
		return err
	}

	if err := s.contactsScenario.Teardown(ctx); err != nil {
		return err
	}

	return nil
}

// Name returns the scenario name
func (s *MixedScenario) Name() string {
	return "mixed"
}
