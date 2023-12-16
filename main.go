package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"

	"golang.org/x/net/html"
)

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

// Need to use the golang.org/x/net/html package to parse the HTML and find the list of cards
func extractItemNamesAndQuantities(node *html.Node) {
	var crawler func(*html.Node)
	crawler = func(n *html.Node) {
		// Check if the node is an <img> tag
		if n.Type == html.ElementNode && n.Data == "img" {
			// Find the next sibling and check if it's a <b> tag
			for sibling := n.NextSibling; sibling != nil; sibling = sibling.NextSibling {
				if sibling.Type == html.ElementNode && sibling.Data == "b" {
					// Extract and print the text inside the <b> tag
					if sibling.FirstChild != nil {
						fmt.Println("Card Name:", sibling.FirstChild.Data)
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

func main() {
	// jar, _ := cookiejar.New(nil)
	// client := &http.Client{Jar: jar}

	// username := os.Args[1]
	// password := os.Args[2]

	// err := login(client, username, password)
	// if err != nil {
	// 	fmt.Println("Login failed:", err)
	// 	return
	// }

	//neodeckPage, err := getNeodeck(client, username)
	// if err != nil {
	// 	fmt.Println("Failed to get Neodeck:", err)
	// 	return
	// }

	// Reading in a file for testing so we aren't logging in each time we run it
	neodeckPage, err := os.Open("output.html")
	if err != nil {
		fmt.Errorf("Failed to read file:", err)
		return
	}
	defer neodeckPage.Close() // closes the file after everything is done

	doc, err := html.Parse(neodeckPage)
	if err != nil {
		fmt.Errorf("Failed to parse file:", err)
		return
	}

	extractItemNamesAndQuantities(doc)
	//fmt.Println("Cards in Neodeck:", doc)
}
