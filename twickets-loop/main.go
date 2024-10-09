package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/stealth"
	"github.com/joho/godotenv"
)

const (
	maxPrice         = 150
	ethanPhoneNumber = "+447476133726"
	dadPhoneNumber   = "+447725841566"
	url              = "https://www.twickets.live/en/event/1836398181106065408#sort=FirstListed&typeFilter=Any&qFilter=1"
)

type Ticket struct {
	Price   float64
	Row     int
	Section int
	ID      string
}

var (
	set = make(map[string]bool)
)

func generateID(price float64, row, section int) string {
	data := fmt.Sprintf("%.2f:%d:%d", price, section, row)

	hash := sha256.New()
	hash.Write([]byte(data))

	return hex.EncodeToString(hash.Sum(nil))
}

func extractPrice(input string) (float64, error) {
	pricePattern := `£(\d+\.\d{2})`
	re := regexp.MustCompile(pricePattern)

	priceMatch := re.FindStringSubmatch(input)

	if len(priceMatch) < 2 {
		return 0, fmt.Errorf("could not extract price")
	}

	price, err := strconv.ParseFloat(priceMatch[1], 64)
	if err != nil {
		return 0, fmt.Errorf("could not parse price: %v", err)
	}

	return price, nil
}

func extractSection(input string) (int, error) {
	sectionPattern := `Section (\d+)`
	re := regexp.MustCompile(sectionPattern)

	sectionMatch := re.FindStringSubmatch(input)

	if len(sectionMatch) < 2 {
		return 0, fmt.Errorf("could not extract section")
	}

	section, err := strconv.Atoi(sectionMatch[1])
	if err != nil {
		return 0, fmt.Errorf("could not parse section: %v", err)
	}

	return section, nil
}

func extractRow(input string) (int, error) {
	rowPattern := `Row (\d+)`
	re := regexp.MustCompile(rowPattern)

	rowMatch := re.FindStringSubmatch(input)

	if len(rowMatch) < 2 {
		return 0, fmt.Errorf("could not extract row")
	}

	row, err := strconv.Atoi(rowMatch[1])
	if err != nil {
		return 0, fmt.Errorf("could not parse row: %v", err)
	}

	return row, nil
}

func extractTicketInfo(input string) (*Ticket, error) {
	price, err := extractPrice(input)
	if err != nil {
		return nil, err
	}

	section, err := extractSection(input)
	if err != nil {
		return nil, err
	}

	row, err := extractRow(input)
	if err != nil {
		return nil, err
	}

	id := generateID(price, row, section)

	return &Ticket{
		Price:   price,
		Row:     row,
		Section: section,
		ID:      id,
	}, nil
}

func getTickets() ([]Ticket, error) {
	controlUrl := launcher.New().
		Headless(true). // Make this false in the future
		Devtools(false).
		MustLaunch()

	browser := rod.New().ControlURL(controlUrl).MustConnect().MustIgnoreCertErrors(true)

	page := stealth.MustPage(browser)
	page.MustSetExtraHeaders("Cache-Control", "no-store")

	page.MustNavigate(url).MustWaitNavigation()

	page.MustWaitIdle()
	page.MustWaitRequestIdle()
	page.MustWaitDOMStable()

	page.MustElement(".container.sort-filter-row.list-group-item.not-football").MustWaitVisible()

	html := page.MustElement("html").MustHTML()
	// write to file
	fmt.Println(html)

	details := page.MustElements(".details-container")

	tickets := make([]Ticket, 0)
	for _, detail := range details {
		ticket, err := extractTicketInfo(detail.MustText())
		if err != nil {
			continue
		}

		if set[ticket.ID] {
			continue
		}

		set[ticket.ID] = true
		tickets = append(tickets, *ticket)
	}

	return tickets, nil

}

func getCheapestTicket() (*Ticket, error) {
	tickets, err := getTickets()
	if err != nil {
		return nil, fmt.Errorf("could not get tickets: %v", err)
	}

	if len(tickets) == 0 {
		return nil, fmt.Errorf("No tickets found")
	}

	cheapestTicket := tickets[0]
	for _, ticket := range tickets[1:] {
		if ticket.Price < cheapestTicket.Price {
			cheapestTicket = ticket
		}
	}

	return &cheapestTicket, nil
}

func sendSMS(ticket Ticket, phoneNumber string) error {
	str := fmt.Sprintf("Ticket found for £%v in section %d, row %d.   Link:%s", ticket.Price, ticket.Section, ticket.Row, url)

	messagePayload := map[string]interface{}{
		"messages": []map[string]string{
			{
				"body": str,
				"to":   phoneNumber,
			},
		},
	}

	payloadBytes, err := json.Marshal(messagePayload)
	if err != nil {
		log.Println("Error marshalling JSON:", err)
		return err
	}
	payload := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("POST", "https://rest.clicksend.com/v3/sms/send", payload)
	if err != nil {
		log.Println("Error creating request:", err)
		return err
	}

	apiUsername := os.Getenv("CLICKSEND_USERNAME")
	apiKey := os.Getenv("CLICKSEND_KEY")

	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(apiUsername, apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error making request:", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		log.Println("SMS sent successfully!")
	} else {
		log.Println("Failed to send SMS, status code:", resp.StatusCode)
	}

	return err
}

func logic() {
	cheapestTicket, err := getCheapestTicket()
	if err != nil {
		log.Println("Error:", err)
		return
	} else {
		log.Printf("Cheapest ticket: %+v", cheapestTicket)
	}

	if cheapestTicket.Price < maxPrice {
		fmt.Println("Ticket found for $", cheapestTicket.Price)
		sendSMS(*cheapestTicket, ethanPhoneNumber)
		// sendSMS(*cheapestTicket, dadPhoneNumber)
	}
}

func logicLoop() {
	ticker := time.NewTicker(10 * time.Second)

	for {
		select {
		case <-ticker.C:
			go logic()
		}
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalln("Error loading .env file")
	}

	logicLoop()
}
