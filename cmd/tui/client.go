package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mangahub/internal/models"
	"net/http"
	"time"
)

type APIClient struct {
	BaseURL string
	Token   string
	Http    *http.Client
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		BaseURL: baseURL,
		Http: &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *APIClient) Register(username, password string) error {
	body, _ := json.Marshal(models.AuthRequest{
		Username: username,
		Password: password,
	})

	resp, err := c.Http.Post(c.BaseURL + "/api/auth/register", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("register error with status code: %v", resp.StatusCode)
	}

	//automatically log the user in after the register
	if err := c.Login(username, password); err != nil {
		return fmt.Errorf("error login after register: %v", err)
	}

	return nil
}

func (c *APIClient) Login(username, password string) error {
	body, _ := json.Marshal(models.AuthRequest{
		Username: username,
		Password: password,
	})

	//public route (doesnt need authentication) => doesnt need header
	resp, err := c.Http.Post(c.BaseURL + "/api/auth/login", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error login with status code: %v", resp.StatusCode)
	}

	var res struct {
		Token string `json:"token"`
	}
	//read the raw json byte stream from the server's respond
	//parse the json to populate newly created res struct 
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return err
	}

	c.Token = res.Token
	return nil
}

func (c *APIClient) GetLibrary() ([]models.LibraryItem, error) {
	//NewRequest allow the attachment of headers 
	req, _ := http.NewRequest("GET", c.BaseURL + "/api/library", nil)
	req.Header.Set("Authorization", "Bearer " + c.Token)

	resp, err := c.Http.Do(req)
	if err != nil {
		return  nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error fetching library with status code %v", resp.StatusCode)
	}

	var items []models.LibraryItem
	err = json.NewDecoder(resp.Body).Decode(&items)

	return items, err
}

func (c *APIClient) AddToLibrary(mangaID int64, status string) (models.LibraryItem, error) {
	body, err := json.Marshal(models.AddLibraryRequest{
		MangaID: mangaID,
		Status: status,
	})
	if err != nil {
		return models.LibraryItem{}, err
	}

	req, err := http.NewRequest("POST", c.BaseURL + "/api/library", bytes.NewBuffer(body))
	if err != nil {
		return models.LibraryItem{}, err
	}
	
	req.Header.Set("Authorization", "Bearer " + c.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Http.Do(req)
	if err != nil {
		return models.LibraryItem{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return models.LibraryItem{}, fmt.Errorf("error add manga with status code %v", resp.StatusCode)
	}

	var res models.LibraryItem 
	err = json.NewDecoder(resp.Body).Decode(&res)

	return res, err
}

func (c *APIClient) UpdateProgress(mangaID int64, currChap int) (models.LibraryItem, error) {
	body, err := json.Marshal(models.UpdateProgressRequest{
		CurrentChapter: &currChap,
	})
	if err != nil {
		return models.LibraryItem{}, err
	}

	url := fmt.Sprintf("%v/api/library/%v/progress", c.BaseURL, mangaID)

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	if err != nil {
		return models.LibraryItem{}, err
	}
	
	req.Header.Set("Authorization", "Bearer " + c.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Http.Do(req) 
	if err != nil {
		return models.LibraryItem{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return models.LibraryItem{}, fmt.Errorf("update chapter error with status code %v", resp.StatusCode)
	}

	var res models.LibraryItem
	err = json.NewDecoder(resp.Body).Decode(&res)
	
	return res, err
} 

func (c *APIClient) DeleteManga(mangaID int64) error {
	url := fmt.Sprintf("%v/api/library/%v", c.BaseURL, mangaID)

	req, err := http.NewRequest("DELETE", url, nil) 
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer " + c.Token)
	

	resp, err := c.Http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("delete error with status code %v", resp.StatusCode)
	}

	//delete sucessful
	return nil
}