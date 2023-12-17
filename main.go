package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// var (
// 	ownedCards   []string
// 	cardsInStock []string
// )

func login(client *http.Client, username, password string) error {
	loginURL := "https://www.neopets.com/login.phtml"
	data := url.Values{
		"username": {username},
		"password": {password},
	}

	resp, err := client.PostForm(loginURL, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func getNeodeck(client *http.Client, username string) (*html.Node, error) {
	neodeckURL := fmt.Sprintf("https://www.neopets.com/games/neodeck/index.phtml?owner=%s&show=cards", username)
	resp, err := client.Get(neodeckURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func getCollectableCardShopStock(client *http.Client) (*html.Node, error) {
	shopURL := fmt.Sprintf("https://www.neopets.com/objects.phtml?type=shop&obj_type=8")

	resp, err := client.Get(shopURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// Need to use the golang.org/x/net/html package to parse the HTML and find the list of cards
func extractOwnedCards(node *html.Node) []string {
	var ownedCards []string
	var crawler func(*html.Node)
	crawler = func(n *html.Node) {
		// Check if the node is an <img> tag
		if n.Type == html.ElementNode && n.Data == "img" {
			// Find the next sibling and check if it's a <b> tag
			for sibling := n.NextSibling; sibling != nil; sibling = sibling.NextSibling {
				if sibling.Type == html.ElementNode && sibling.Data == "b" {
					// Extract and print the text inside the <b> tag
					if sibling.FirstChild != nil {
						ownedCards = append(ownedCards, sibling.FirstChild.Data)
						// fmt.Println("Card Name:", sibling.FirstChild.Data)
					}
					break
				}
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			crawler(child)
		}
	}
	crawler(node)

	return ownedCards
}

func extractShopStock(node *html.Node) []string {
	var cardsInStock []string
	var crawler func(*html.Node)
	crawler = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "div" {
			// Check if this is a 'shop-item' div
			isShopItem := false
			for _, a := range n.Attr {
				if a.Key == "class" && strings.Contains(a.Val, "shop-item") {
					isShopItem = true
					break
				}
			}

			if isShopItem {
				// Find the 'item-img' div inside this 'shop-item' div
				for child := n.FirstChild; child != nil; child = child.NextSibling {
					if child.Type == html.ElementNode && child.Data == "div" {
						for _, a := range child.Attr {
							if a.Key == "class" && strings.Contains(a.Val, "item-img") {
								// Extract and print the 'data-name' attribute
								for _, a := range child.Attr {
									if a.Key == "data-name" {
										cardsInStock = append(cardsInStock, a.Val)
										// fmt.Println("Card Name in stock:", a.Val)
										break
									}
								}
							}
						}
					}
				}
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			crawler(child)
		}
	}
	crawler(node)

	return cardsInStock
}

// Helper function to get the value of a specific attribute from a node
func getAttributeValue(node *html.Node, attrName string) string {
	for _, a := range node.Attr {
		if a.Key == attrName {
			return a.Val
		}
	}
	return ""
}

// contains checks if slice1 contains all elements of slice2.
func contains(slice1, slice2 []string) []string {
	var missingCards []string

	for _, val2 := range slice2 {
		found := false
		for _, val1 := range slice1 {
			if val1 == val2 {
				found = true
				break
			}
		}
		if !found {
			missingCards = append(missingCards, val2)
		}
	}
	return missingCards
}

func sleepRandom() {
	rand.Seed(time.Now().UnixNano())

	randomSeconds := rand.Intn(81) + 5

	fmt.Println("Sleeping for", randomSeconds)

	time.Sleep(time.Duration(randomSeconds) * time.Second)
}

func refreshPage() {
	rand.Seed(time.Now().UnixNano())

	randomSeconds := rand.Intn(180) + 400

	fmt.Printf("Sleeping for %d seconds\n", randomSeconds)

	time.Sleep(time.Duration(randomSeconds) * time.Second)
}

func main() {
	var shopStock *html.Node
	var neodeckCards *html.Node

	username := os.Args[1]
	password := os.Args[2]

	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}

	if len(os.Args) == 0 {
		fmt.Println("No username or password passed in. Aborting")
		return
	}

	err := login(client, username, password)
	if err != nil {
		fmt.Println("Login failed:", err)
		return
	}
	fmt.Println("Login Successful")

	sleepRandom()

	// Gets the current cards in your NeoDeck
	neodeckCards, err = getNeodeck(client, username)
	if err != nil {
		fmt.Errorf("Failed to read file:", err)
		return
	}

	sleepRandom()

	for {
		fmt.Println("Retrieving card shop stock")
		shopStock, err = getCollectableCardShopStock(client)
		if err != nil {
			fmt.Println("Failed to get Card shop stock:", err)
			return
		}
		fmt.Println("Card shop stock received")

		ownedCards := extractOwnedCards(neodeckCards)
		cardsInStock := extractShopStock(shopStock)

		missingCards := contains(ownedCards, cardsInStock)

		fmt.Println(time.Now())
		if len(missingCards) == 0 {
			fmt.Println("We own all cards in stock")
		} else {
			fmt.Println("Missing cards currently in stock:")
			for _, card := range missingCards {
				fmt.Println(card)
			}
		}

		refreshPage()
		fmt.Println("=============")
	}
}
