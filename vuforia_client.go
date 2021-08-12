package vuforia

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

// vuforiaUrl is the endpoint for the Vuforia Web Services API
const vuforiaUrl = "vws.vuforia.com"

type Client interface {
	// PostTarget adds a new target
	PostTarget(*PostTargetRequest) (*PostTargetResponse, error)
	// GetTarget retrieves the target
	GetTarget(*GetTargetRequest) (*GetTargetResponse, error)
	// UpdateTarget updates the target
	UpdateTarget(*UpdateTargetRequest) (*UpdateTargetResponse, error)
	// DeleteTarget deletes the target
	DeleteTarget(*DeleteTargetRequest) (*DeleteTargetResponse, error)
	// TargetSummary retrieves summary of the target
	TargetSummary(*TargetSummaryRequest) (*TargetSummaryResponse, error)
	// DatabaseSummary retrieves the summary of the database
	DatabaseSummary() (*DatabaseSummaryResponse, error)
}

type ClientConfig struct {
	SecretKey, AccessKey string
	Client               *http.Client
}

type client struct {
	cfg ClientConfig
}

func NewClient(cfg ClientConfig) (Client, error) {
	if cfg.SecretKey == "" {
		return nil, fmt.Errorf("vuforia SecretKey must be set")
	}

	if cfg.AccessKey == "" {
		return nil, fmt.Errorf("vuforia AccessKey must be set")
	}

	if cfg.Client == nil {
		cfg.Client = http.DefaultClient
	}

	return &client{cfg: cfg}, nil
}

type PostTargetRequest struct {
	// Name of the target, unique within a database
	Name string `json:"name"`
	// Width of the target in scene unit
	Width float64 `json:"width"`
	// Image is the base64 encoded binary recognition image data
	Image string `json:"image"`
	// Active indicates whether or not the target is active for query (Optional)
	Active *bool `json:"active_flag,omitempty"`
	// Metadata is the base64 encoded application metadata associated with the target (Optional)
	Metadata *string `json:"application_metadata,omitempty"`
}

type PostTargetResponse struct {
	// TargetId is the ID of the target
	TargetId string `json:"target_id"`
	// TransactionId is the ID of the transaction
	TransactionId string `json:"transaction_id"`
	// ResultCode is one of the VWS API Result Code.
	// https://library.vuforia.com/articles/Solution/How-To-Use-the-Vuforia-Web-Services-API.html#How-To-Interperete-VWS-API-Result-Codes
	ResultCode string `json:"result_code"`
}

func (c *client) PostTarget(input *PostTargetRequest) (*PostTargetResponse, error) {
	if input == nil {
		panic("input is <nil>")
	}

	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("https://%s/targets", vuforiaUrl), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	if err = prepare(c.cfg.SecretKey, c.cfg.AccessKey, req, body); err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer safeClose(resp)

	var v PostTargetResponse
	err = json.NewDecoder(resp.Body).Decode(&v)
	if err != nil {
		return nil, err
	}

	return &v, nil
}

type GetTargetRequest struct {
	// TargetId is the ID of the target to retrieve
	TargetId string
}

type GetTargetResponse struct {
	// TransactionId is the ID of the transaction
	TransactionId string `json:"transaction_id"`
	// ResultCode is one of the VWS API Result Code.
	// https://library.vuforia.com/articles/Solution/How-To-Use-the-Vuforia-Web-Services-API.html#How-To-Interperete-VWS-API-Result-Codes
	ResultCode string `json:"result_code"`
	// Status of the target (Processing, Success, Failure)
	Status string `json:"status"`
	// TargetRecord is the target information
	TargetRecord struct {
		// TargetId is the ID of the target
		TargetId string `json:"target_id"`
		// Active indicates whether or not the target is active for query; default is true
		Active bool `json:"active_flag"`
		// Name of the target, unique within a database
		Name string `json:"name"`
		// Width of the target in scene unit
		Width float64 `json:"width"`
		// TrackingRating is the rating of the target recognition image for tracking purposes
		TrackingRating int `json:"tracking_rating"`
	} `json:"target_record"`
}

// https://library.vuforia.com/articles/Solution/How-To-Use-the-Vuforia-Web-Services-API.html#How-To-Retrieve-a-Target-Record
func (c *client) GetTarget(input *GetTargetRequest) (*GetTargetResponse, error) {
	if input == nil {
		panic("input is <nil>")
	}
	if input.TargetId == "" {
		return nil, errors.New("TargetId must be provided")
	}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://%s/targets/%s", vuforiaUrl, input.TargetId), nil)
	if err != nil {
		return nil, err
	}

	if err = prepare(c.cfg.SecretKey, c.cfg.AccessKey, req, nil); err != nil {
		return nil, err
	}

	resp, err := c.cfg.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer safeClose(resp)

	var v GetTargetResponse
	err = json.NewDecoder(resp.Body).Decode(&v)
	if err != nil {
		return nil, err
	}

	return &v, nil
}

type UpdateTargetRequest struct {
	// TargetId is the ID of the target to update
	TargetId string `json:"-"`
	// Name of the target, unique within a database (Optional)
	Name *string `json:"name,omitempty"`
	// Width of the target in scene unit (Optional)
	Width *float64 `json:"width,omitempty"`
	// Image is the base64 encoded binary recognition image data (Optional)
	// https://library.vuforia.com/features/images/image-targets.html
	Image *string `json:"image,omitempty"`
	// Active Iidicates whether or not the target is active for query (Optional)
	Active *bool `json:"active_flag,omitempty"`
	// Metadata is the base64 encoded application metadata associated with the target (Optional)
	Metadata *string `json:"application_metadata,omitempty"`
}

type UpdateTargetResponse struct {
	// TransactionId is the ID of the transaction
	TransactionId string `json:"transaction_id"`
	// ResultCode is one of the VWS API Result Code.
	// https://library.vuforia.com/articles/Solution/How-To-Use-the-Vuforia-Web-Services-API.html#How-To-Interperete-VWS-API-Result-Codes
	ResultCode string `json:"result_code"`
}

func (c *client) UpdateTarget(input *UpdateTargetRequest) (*UpdateTargetResponse, error) {
	if input == nil {
		panic("input is <nil>")
	}
	if input.TargetId == "" {
		return nil, errors.New("TargetId must be provided")
	}

	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("https://%s/targets/%s", vuforiaUrl, input.TargetId), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	if err = prepare(c.cfg.SecretKey, c.cfg.AccessKey, req, body); err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer safeClose(resp)

	var v UpdateTargetResponse
	err = json.NewDecoder(resp.Body).Decode(&v)
	if err != nil {
		return nil, err
	}

	return &v, nil
}

type DeleteTargetRequest struct {
	// TargetId is the ID of target to delete
	TargetId string
}

type DeleteTargetResponse struct {
	// TransactionId is the ID of the transaction
	TransactionId string `json:"transaction_id"`
	// ResultCode is one of the VWS API Result Code.
	// https://library.vuforia.com/articles/Solution/How-To-Use-the-Vuforia-Web-Services-API.html#How-To-Interperete-VWS-API-Result-Codes
	ResultCode string `json:"result_code"`
}

// https://library.vuforia.com/articles/Solution/How-To-Use-the-Vuforia-Web-Services-API.html#How-To-Delete-a-Target
func (c *client) DeleteTarget(input *DeleteTargetRequest) (*DeleteTargetResponse, error) {
	if input == nil {
		panic("input is <nil>")
	}
	if input.TargetId == "" {
		return nil, errors.New("TargetId must be provided")
	}

	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("https://%s/targets/%s", vuforiaUrl, input.TargetId), nil)
	if err != nil {
		return nil, err
	}

	if err = prepare(c.cfg.SecretKey, c.cfg.AccessKey, req, nil); err != nil {
		return nil, err
	}

	resp, err := c.cfg.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer safeClose(resp)

	var v DeleteTargetResponse
	err = json.NewDecoder(resp.Body).Decode(&v)
	if err != nil {
		return nil, err
	}

	return &v, nil
}

type TargetSummaryRequest struct {
	// TargetId is the ID of the target to retrieve summary of
	TargetId string
}

type TargetSummaryResponse struct {
	// TransactionId is the ID of the transaction
	TransactionId string `json:"transaction_id"`
	// ResultCode is one of the VWS API Result Code.
	// https://library.vuforia.com/articles/Solution/How-To-Use-the-Vuforia-Web-Services-API.html#How-To-Interperete-VWS-API-Result-Codes
	ResultCode string `json:"result_code"`
	// Status of the target (Processing, Success, Failure)
	Status string `json:"status"`
	// DatabaseName of is the database name the current target resides in
	DatabaseName string `json:"database_name"`
	// TargteName is the target name
	TargetName string `json:"target_name"`
	// UploadDate is the date of target upload (Specified as YYYY-MM-DD)
	UploadDate string `json:"upload_date"`
	// Active indicates whether or not the target is active for query; default is true
	Active bool `json:"active_flag"`
	// TrackingRating is the rating of the target recognition image for tracking purposes
	TrackingRating int `json:"tracking_rating"`
	// TotalRecos is the total count of the recognitions for this target
	TotalRecos int `json:"total_recos"`
	// CurrentMonthRecos is the total count of recognitions in the current month (Set to 0 if the Status is not "Success")
	CurrentMonthRecos int `json:"current_month_recos"`
	// PreviousMonthRecos is the total count of the recognitions in the previous month (Set to 0 if the Status is not "Success")
	PreviousMonthRecos int `json:"previous_month_recos"`
}

// https://library.vuforia.com/articles/Solution/How-To-Use-the-Vuforia-Web-Services-API.html#How-To-Retrieve-a-Target-Summary-Report
func (c *client) TargetSummary(input *TargetSummaryRequest) (*TargetSummaryResponse, error) {
	if input == nil {
		panic("input is <nil>")
	}
	if input.TargetId == "" {
		return nil, errors.New("TargetId must be provided")
	}
	
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://%s/summary/%s", vuforiaUrl, input.TargetId), nil)
	if err != nil {
		return nil, err
	}

	if err = prepare(c.cfg.SecretKey, c.cfg.AccessKey, req, nil); err != nil {
		return nil, err
	}

	resp, err := c.cfg.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer safeClose(resp)

	var v TargetSummaryResponse
	err = json.NewDecoder(resp.Body).Decode(&v)
	if err != nil {
		return nil, err
	}

	return &v, nil
}

type DatabaseSummaryResponse struct {
	// TransactionId is the ID of the transaction
	TransactionId string `json:"transaction_id"`
	// ResultCode is one of the VWS API Result Code.
	// https://library.vuforia.com/articles/Solution/How-To-Use-the-Vuforia-Web-Services-API.html#How-To-Interperete-VWS-API-Result-Codes
	ResultCode string `json:"result_code"`
	// Name is the database name
	Name string `json:"name"`
	// ActiveImages is the total number of images with active_flag = true, and status = success for the database
	ActiveImages int `json:"active_images"`
	// InactiveImages is the total number of images with active_flag = false, and status = success for the database
	InactiveImages int `json:"inactive_images"`
	// FailedImages is the total number of images with status = fail for the data
	FailedImages int `json:"failed_images"`
}

// https://library.vuforia.com/articles/Solution/How-To-Use-the-Vuforia-Web-Services-API.html#How-To-Get-a-Database-Summary-Report
func (c *client) DatabaseSummary() (*DatabaseSummaryResponse, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://%s/summary", vuforiaUrl), nil)
	if err != nil {
		return nil, err
	}

	if err = prepare(c.cfg.SecretKey, c.cfg.AccessKey, req, nil); err != nil {
		return nil, err
	}

	resp, err := c.cfg.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer safeClose(resp)

	var v DatabaseSummaryResponse
	err = json.NewDecoder(resp.Body).Decode(&v)
	if err != nil {
		return nil, err
	}

	return &v, nil
}

func safeClose(resp *http.Response) {
	if resp.Body != nil {
		_, _ = ioutil.ReadAll(resp.Body)
		_ = resp.Body.Close()
	}
}
