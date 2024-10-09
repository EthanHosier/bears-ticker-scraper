package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"slices"
	"time"

	"github.com/joho/godotenv"
)

type Pricing struct {
	Options string  `json:"options"`
	Prices  []Price `json:"prices"`
}

type Price struct {
	ID              *string `json:"id"`
	CurrencyCode    string  `json:"currencyCode"`
	Label           string  `json:"label"`
	FaceValue       float64 `json:"faceValue"`
	OriginalFee     float64 `json:"originalFee"`
	NetFee          float64 `json:"netFee"`
	NetSellingPrice float64 `json:"netSellingPrice"`
}

type ResponseData struct {
	Type                     string          `json:"type"`
	Area                     string          `json:"area"`
	Section                  string          `json:"section"`
	Row                      string          `json:"row"`
	ID                       string          `json:"id"`
	Pricing                  Pricing         `json:"pricing"`
	CommonAttributes         []interface{}   `json:"commonAttributes"`
	IndividualAttributes     [][]interface{} `json:"individualAttributes"`
	Splits                   []int           `json:"splits"`
	DeliveryMethodTypes      []string        `json:"deliveryMethodTypes"`
	SellerWillConsiderOffers bool            `json:"sellerWillConsiderOffers"`
	SegmentID                string          `json:"segmentId"`
}

type Response struct {
	ResponseData []ResponseData `json:"responseData"`
	ResponseCode int            `json:"responseCode"`
	Description  string         `json:"description"`
	Clock        string         `json:"clock"`
}

type Ticket struct {
	Section string
	Row     string
	Price   float64
	ID      string // from section, row and price
}

const (
	tickets_url      = "https://www.twickets.live/app/block/"
	url              = "https://www.twickets.live/services/g2/inventory/listings/1836398181106065408?api_key=83d6ec0c-54bb-4da3-b2a1-f3cb47b984f1"
	split            = 2
	maxPrice         = 115.0
	loopTime         = 10 * time.Second
	ethanPhoneNumber = "+447476133726"
	dadPhoneNumber   = "+447725841566"
)

var (
	set = map[string]bool{}
)

func getResponseData() ([]ResponseData, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("Error creating request: %v", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("DNT", "1") // Do Not Track request header

	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error making HTTP request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response body: %v", err)
	}

	var data Response
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, fmt.Errorf("Error parsing JSON: %v", err)
	}

	return data.ResponseData, nil
}

func generateTicketId(section string, row string, price float64) string {
	return fmt.Sprintf("%s-%s-%f", section, row, price)
}

func extractId(str string) (string, error) {
	re := regexp.MustCompile(`@(.+)`)

	match := re.FindStringSubmatch(str)

	if len(match) > 1 {
		return match[1], nil
	}

	return "", fmt.Errorf("No match found")
}

func getRelevantTickets(responseDatas []ResponseData) []Ticket {
	tickets := make([]Ticket, 0)

	for _, responseData := range responseDatas {

		if !slices.Contains(responseData.Splits, split) {
			continue
		}

		ticketPrice := responseData.Pricing.Prices[0]

		id, err := extractId(responseData.ID)
		if err != nil {
			log.Println("Error extracting ID:", err)
			continue
		}

		p := (ticketPrice.NetSellingPrice + ticketPrice.NetFee) / 100
		ticket := Ticket{
			Section: responseData.Section,
			Row:     responseData.Row,
			Price:   p,
			ID:      id,
		}

		tickets = append(tickets, ticket)
	}

	return tickets
}

func getCheapestTicket(tickets []Ticket) *Ticket {
	var cheapestTicket *Ticket

	for _, ticket := range tickets {
		if set[ticket.ID] {
			continue
		}

		if cheapestTicket == nil || ticket.Price < cheapestTicket.Price {
			cheapestTicket = &ticket
		}
	}

	return cheapestTicket
}

func sendSMS(ticket Ticket, phoneNumber string) error {
	link := fmt.Sprintf("%s%s,%d", tickets_url, ticket.ID, split)
	str := fmt.Sprintf("Ticket found for £%v in section %s, row %s.   Link:%s", ticket.Price, ticket.Section, ticket.Row, link)

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
	responseData, err := getResponseData()
	if err != nil {
		log.Println(err)
		return
	}

	tickets := getRelevantTickets(responseData)
	if len(tickets) == 0 {
		log.Println("No tickets found")
		return
	}

	cheapestTicket := getCheapestTicket(tickets)
	if cheapestTicket == nil {
		log.Println("No new tickets available")
		return
	}

	set[cheapestTicket.ID] = true
	if cheapestTicket.Price > maxPrice {
		log.Println(cheapestTicket.Price)
		log.Println("No tickets available within the price range")
		return
	}

	log.Printf("Ticket found for £%v in section %v, row %v\n", cheapestTicket.Price, cheapestTicket.Section, cheapestTicket.Row)

	if err := sendSMS(*cheapestTicket, ethanPhoneNumber); err != nil {
		log.Println("Error sending SMS:", err)
	}

	if err := sendSMS(*cheapestTicket, dadPhoneNumber); err != nil {
		log.Println("Error sending SMS:", err)
	}
}

func logicLoop() {
	ticker := time.NewTicker(loopTime)

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
