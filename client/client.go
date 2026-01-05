package client

import (
	"bataille-navale/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	MaxRetries int
	RetryDelay time.Duration
}

func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		MaxRetries: 3,
		RetryDelay: 500 * time.Millisecond,
	}
}

func (c *Client) retryRequest(fn func() error) error {
	var lastErr error
	for i := 0; i < c.MaxRetries; i++ {
		err := fn()
		if err == nil {
			return nil
		}
		lastErr = err
		
		if httpErr, ok := err.(*httpError); ok && httpErr.StatusCode >= 400 && httpErr.StatusCode < 500 {
			return err
		}
		
		if i < c.MaxRetries-1 {
			time.Sleep(c.RetryDelay)
		}
	}
	return fmt.Errorf("échec après %d tentatives: %w", c.MaxRetries, lastErr)
}

type httpError struct {
	StatusCode int
	Message    string
}

func (e *httpError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}

func (c *Client) GetBoard() (*models.BoardResponse, error) {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/board")
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération du plateau: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erreur serveur: %d", resp.StatusCode)
	}
	
	var board models.BoardResponse
	if err := json.NewDecoder(resp.Body).Decode(&board); err != nil {
		return nil, fmt.Errorf("erreur lors du décodage de la réponse: %w", err)
	}
	
	return &board, nil
}

func (c *Client) GetBoatsCount() (int, error) {
	var boats models.BoatsResponse
	
	err := c.retryRequest(func() error {
		resp, err := c.HTTPClient.Get(c.BaseURL + "/boats")
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			return &httpError{StatusCode: resp.StatusCode, Message: "erreur serveur"}
		}
		
		if err := json.NewDecoder(resp.Body).Decode(&boats); err != nil {
			return fmt.Errorf("erreur décodage: %w", err)
		}
		
		return nil
	})
	
	if err != nil {
		return 0, fmt.Errorf("erreur lors de la récupération des bateaux: %w", err)
	}
	
	return boats.RemainingBoats, nil
}

func (c *Client) Hit(x, y int) (*models.HitResponse, error) {
	var hitResp models.HitResponse
	
	hitReq := models.HitRequest{X: x, Y: y}
	body, err := json.Marshal(hitReq)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la création de la requête: %w", err)
	}
	
	err = c.retryRequest(func() error {
		resp, err := c.HTTPClient.Post(c.BaseURL+"/hit", "application/json", bytes.NewBuffer(body))
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			return &httpError{StatusCode: resp.StatusCode, Message: string(bodyBytes)}
		}
		
		if err := json.NewDecoder(resp.Body).Decode(&hitResp); err != nil {
			return fmt.Errorf("erreur décodage: %w", err)
		}
		
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("erreur lors de l'envoi du tir: %w", err)
	}
	
	return &hitResp, nil
}

func (c *Client) GetHits() (*models.HitsResponse, error) {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/hits")
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des tirs: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erreur serveur: %d", resp.StatusCode)
	}
	
	var hits models.HitsResponse
	if err := json.NewDecoder(resp.Body).Decode(&hits); err != nil {
		return nil, fmt.Errorf("erreur lors du décodage de la réponse: %w", err)
	}
	
	return &hits, nil
}

func (c *Client) IsAlive() bool {
	boats, err := c.GetBoatsCount()
	if err != nil {
		return false
	}
	return boats > 0
}
