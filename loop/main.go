package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/joho/godotenv"
)

// "fmt"
// "regexp"
// "strings"

// "github.com/go-rod/rod"

// type Ticket struct {
// 	Price   string `json:"price"`
// 	Row     string `json:"row"`
// 	Section string `json:"section"`
// }

// func main() {
// 	browser := rod.New().MustConnect()

// 	page := browser.MustPage("https://www.viagogo.com/ww/Sports-Tickets/NFL/NFL-Matchups/Bears-vs-Jaguars/E-153572300?quantity=2&listingQty=&sections=&ticketClasses=&rows=&seats=&seatTypes=").MustWaitIdle()
// 	page.MustWaitRequestIdle()

// 	divs := page.MustElements(".sc-57jg3s-0.uNHdb")

// 	tickets := []Ticket{}

// 	for _, div := range divs {
// 		text := div.MustText()
// 		if strings.Contains(text, "Sold") {
// 			continue
// 		}

// 		re := regexp.MustCompile(`£\d+`)
// 		price := re.FindString(text)

// 		reRow := regexp.MustCompile(`Row\s+(\d+)`)
// 		row := reRow.FindStringSubmatch(text)

// 		reSection := regexp.MustCompile(`Section\s+(\d+)`)
// 		section := reSection.FindStringSubmatch(text)

// 		// Check if row was found
// 		if len(row) <= 1 {
// 			break
// 		}

// 		// Check if section was found
// 		if len(section) <= 1 {
// 			break
// 		}

// 		ticket := Ticket{
// 			Price:   price,
// 			Row:     row[1],
// 			Section: section[1],
// 		}

// 		tickets = append(tickets, ticket)
// 	}

// 	cheapestTicket, err := getCheapestTicket(tickets)
// 	if err != nil {
// 		fmt.Println("Error:", err)
// 	} else {
// 		fmt.Printf("Cheapest ticket: %+v", cheapestTicket)
// 	}
// }

// func getCheapestTicket(tickets []Ticket) (Ticket, error) {
// 	cheapest := Ticket{
// 		Price: "£999999",
// 	}

// 	for _, ticket := range tickets {
// 		if ticket.Price < cheapest.Price {
// 			cheapest = ticket
// 		}
// 	}

// 	if cheapest.Price == "£999999" {
// 		return cheapest, fmt.Errorf("No tickets found")
// 	}

// 	return cheapest, nil
// }

type AppData struct {
	AppName string `json:"appName"`
	Grid    Grid   `json:"grid"`
}

type Grid struct {
	Items []Item `json:"items"`
}

type Item struct {
	ID                                  int64               `json:"id"`
	ClientApplicationID                 int                 `json:"clientApplicationId"`
	EventID                             int                 `json:"eventId"`
	Section                             string              `json:"section"`
	SectionID                           int                 `json:"sectionId"`
	SectionMapName                      string              `json:"sectionMapName"`
	SectionType                         int                 `json:"sectionType"`
	Row                                 string              `json:"row"`
	SeatFromInternal                    string              `json:"seatFromInternal"`
	HasSeatDetails                      bool                `json:"hasSeatDetails"`
	HasSeatDetailsUS                    bool                `json:"hasSeatDetailsUS"`
	AvailableTickets                    int                 `json:"availableTickets"`
	ListingPreviewPriceAndFeeDisclosure ValueDisclosure     `json:"listingPreviewPriceAndFeeDisclosure"`
	SoldXTimeAgoSiteMessage             SoldMessage         `json:"soldXTimeAgoSiteMessage"`
	ShowRecentlySold                    bool                `json:"showRecentlySold"`
	AvailableQuantities                 []int               `json:"availableQuantities"`
	TicketClass                         int                 `json:"ticketClass"`
	TicketClassName                     string              `json:"ticketClassName"`
	MaxQuantity                         int                 `json:"maxQuantity"`
	HasListingNotes                     bool                `json:"hasListingNotes"`
	ListingNotes                        []ListingNote       `json:"listingNotes"`
	RowID                               int                 `json:"rowId"`
	IsUsersListing                      bool                `json:"isUsersListing"`
	IsPreUploaded                       bool                `json:"isPreUploaded"`
	RowContent                          string              `json:"rowContent"`
	RawPrice                            float64             `json:"rawPrice"`
	Price                               string              `json:"price"`
	TicketTypeID                        int                 `json:"ticketTypeId"`
	TicketTypeGroupID                   int                 `json:"ticketTypeGroupId"`
	ListingTypeID                       int                 `json:"listingTypeId"`
	ListingCurrencyCode                 string              `json:"listingCurrencyCode"`
	BuyerCurrencyCode                   string              `json:"buyerCurrencyCode"`
	QualityRank                         int                 `json:"qualityRank"`
	FaceValue                           float64             `json:"faceValue"`
	FaceValueCurrencyCode               string              `json:"faceValueCurrencyCode"`
	VfsURL                              string              `json:"vfsUrl"`
	FormattedActiveSince                string              `json:"formattedActiveSince"`
	IsSeatedTogether                    bool                `json:"isSeatedTogether"`
	SellerUserID                        string              `json:"sellerUserId"`
	ShowVfsInListing                    bool                `json:"showVfsInListing"`
	HideSeatAndRowInfo                  bool                `json:"hideSeatAndRowInfo"`
	SellerHideSeatInfo                  bool                `json:"sellerHideSeatInfo"`
	AipHash                             string              `json:"aipHash"`
	IsMLBVerified                       bool                `json:"isMLBVerified"`
	IsStanding                          bool                `json:"isStanding"`
	CreatedDateTime                     string              `json:"createdDateTime"`
	IsHighestListingScore               bool                `json:"isHighestListingScore"`
	IsMostAffordable                    bool                `json:"isMostAffordable"`
	IsSponsored                         bool                `json:"isSponsored"`
	IsCheapestListing                   bool                `json:"isCheapestListing"`
	InventoryListingScore               *InventoryScore     `json:"inventoryListingScore,omitempty"`
	TicketsRemainingMessage             *RemainingMessage   `json:"ticketsRemainingMessage,omitempty"`
	BestSellingInSectionMessage         *BestSellingMessage `json:"bestSellingInSectionMessage,omitempty"`
}

type ValueDisclosure struct {
	HasValue bool `json:"hasValue"`
}

type SoldMessage struct {
	Message            string `json:"message"`
	Qualifier          string `json:"qualifier"`
	HasValue           bool   `json:"hasValue"`
	FeatureTrackingKey string `json:"featureTrackingKey"`
}

type ListingNote struct {
	ListingNoteID                   int    `json:"listingNoteId"`
	ListingNoteContentID            int    `json:"listingNoteContentId"`
	FormattedListingNoteContent     string `json:"formattedListingNoteContent"`
	ListingNoteTypeID               int    `json:"listingNoteTypeId"`
	ShowToBuyer                     bool   `json:"showToBuyer"`
	HideInMock                      bool   `json:"hideInMock"`
	SiteAddedListingNote            bool   `json:"siteAddedListingNote"`
	AisleListingNoteWithSplit       bool   `json:"aisleListingNoteWithSplit"`
	ListingNoteDescriptionContentID int    `json:"listingNoteDescriptionContentId"`
	FormattedListingNoteDescription string `json:"formattedListingNoteDescription"`
}

type InventoryScore struct {
	Discount         float64 `json:"discount"`
	StarRating       float64 `json:"starRating"`
	DealScore        float64 `json:"dealScore"`
	SeatQualityScore float64 `json:"seatQualityScore"`
}

type RemainingMessage struct {
	Message            string `json:"message"`
	Qualifier          string `json:"qualifier"`
	HasValue           bool   `json:"hasValue"`
	FeatureTrackingKey string `json:"featureTrackingKey"`
}

type BestSellingMessage struct {
	Message            string `json:"message"`
	Qualifier          string `json:"qualifier"`
	Disclaimer         string `json:"disclaimer"`
	HasValue           bool   `json:"hasValue"`
	FeatureTrackingKey string `json:"featureTrackingKey"`
}

type Ticket struct {
	ID      int64  `json:"id"`
	Price   int    `json:"price"`
	Row     string `json:"row"`
	Section string `json:"section"`
}

var (
	set = make(map[int64]bool)
)

const (
	maxPrice         = 100
	ethanPhoneNumber = "+447476133726"
	dadPhoneNumber   = "+447725841566"
	url              = "https://www.viagogo.com/ww/Sports-Tickets/NFL/NFL-Matchups/Bears-vs-Jaguars/E-153572300?quantity=2&listingQty=&sections=&ticketClasses=&rows=&seats=&seatTypes="
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalln("Error loading .env file")
	}

	logicLoop()
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

func logic() {
	cheapestTicket, err := getCheapestTicket()
	if err != nil {
		log.Println("Error:", err)
	} else {
		log.Printf("Cheapest ticket: %+v", cheapestTicket)
	}

	if cheapestTicket.Price < maxPrice {
		fmt.Println("Ticket found for $", cheapestTicket.Price)
		sendSMS(*cheapestTicket, ethanPhoneNumber)
		// sendSMS(*cheapestTicket, dadPhoneNumber)
	}
}

func sendSMS(ticket Ticket, phoneNumber string) error {
	link := fmt.Sprintf("%s&listingId=%d", url, ticket.ID)
	str := fmt.Sprintf("Ticket found for $%d in section %s, row %s.   Link:%s", ticket.Price, ticket.Section, ticket.Row, link)

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

func getCheapestTicket() (*Ticket, error) {
	appData, err := getAppData()
	if err != nil {
		return nil, err
	}

	tickets := []Ticket{}
	for _, item := range appData.Grid.Items {
		if item.SoldXTimeAgoSiteMessage.HasValue {
			continue
		}

		re := regexp.MustCompile("[^0-9]+")

		numericString := re.ReplaceAllString(item.Price, "")
		price, err := strconv.Atoi(numericString)

		if err != nil {
			return nil, err
		}

		ticket := Ticket{
			ID:      item.ID,
			Price:   price,
			Row:     item.Row,
			Section: item.Section,
		}

		if set[item.ID] {
			continue
		}

		tickets = append(tickets, ticket)
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

	set[cheapestTicket.ID] = true
	return &cheapestTicket, nil
}

func getAppData() (*AppData, error) {

	// Create a new HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Add headers to prevent caching
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("Pragma", "no-cache")

	// Perform the request
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("Error: Failed to fetch the page. Status code: %d", response.StatusCode)
	}

	// Parse the HTML document
	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return nil, err
	}

	scriptContent := doc.Find("script#index-data").Text()

	var appData AppData
	err = json.Unmarshal([]byte(scriptContent), &appData)

	if err != nil {
		return nil, err
	}

	return &appData, nil
}
