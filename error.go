package vuforia

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type APIError struct {
	ResultCode    string `json:"result_code"`
	TransactionId string `json:"transaction_id"`
}

func (e APIError) Error() string {
	return fmt.Sprintf("vuforia request failed (ResultCode = %s, Transaction ID = %s): %s", e.ResultCode, e.TransactionId, vuforiaAPIErrors[e.ResultCode])
}

var vuforiaAPIErrors = map[string]string{
	"AuthenticationFailure":  "Signature authentication failed",
	"RequestTimeTooSkewed":   "Request timestamp outside allowed range",
	"TargetNameExist":        "The corresponding target name already exists",
	"RequestQuotaReached":    "The maximum number of API calls for this database has been reached",
	"TargetStatusProcessing": "The target is in the processing state and cannot be updated",
	"TargetStatusNotSuccess": "The request could not be completed because the target is not in the success state",
	"TargetQuotaReached":     "The maximum number of targets for this database has been reached",
	"ProjectSuspended":       "The request could not be completed because this database has been suspended",
	"ProjectInactive":        "The request could not be completed because this database is inactive",
	"ProjectHasNoApiAccess":  "The request could not be completed because this database is not allowed to make API requests",
	"UnknownTarget":          "The specified target ID does not exist",
	"BadImage":               "Image corrupted or format not supported",
	"ImageTooLarge":          "Target metadata size exceeds maximum limit",
	"MetadataTooLarge":       "Image size exceeds maximum limit",
	"DateRangeError":         "Start date is after the end date",
	"Fail":                   "The request was invalid and could not be processed (Check the request headers and fields)",
}

func checkError(resp *http.Response) error {
	switch {
	case isServerError(resp.StatusCode):
		return fmt.Errorf("the server encountered an internal error (Status = %d); please retry the request", resp.StatusCode)
	case isAPIError(resp.StatusCode):
		var e APIError
		err := json.NewDecoder(resp.Body).Decode(&e)
		if err != nil {
			return err
		}

		return e
	default:
		return nil
	}
}

func isServerError(status int) bool {
	return status >= 500
}

func isAPIError(status int) bool {
	return status >= 400 && status < 500
}
