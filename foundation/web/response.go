package web

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"net/http"
)

func Respond(ctx context.Context, w http.ResponseWriter, data interface{}, statusCode int) error{
	if v, ok := ctx.Value(KeyValues).(*Values); ok {
		v.StatusCode = statusCode
	}

	// If there is nothing to marshal then set status code and return.
	if statusCode == http.StatusNoContent {
		w.WriteHeader(statusCode)
		return nil
	}

	// Convert the response value to JSON.
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Set the content type and headers once we know marshaling has succeeded.
	w.Header().Set("Content-Type", "application/json")

	// Write the status code to the response.
	w.WriteHeader(statusCode)

	// Send the result back to the client.
	if _, err := w.Write(jsonData); err != nil {
		return err
	}

	return nil
}



func RespondError(ctx context.Context, w http.ResponseWriter, err error) error{
	if webErr, ok := errors.Cause(err).(*Error); ok{
		er := ErrorResponse{
			Error: 	webErr.Err.Error(),
			Fields: webErr.Fields,
		}
		if err := Respond(ctx, w, er, webErr.Status); err!=nil{
			return err
		}
	}

	er := ErrorResponse{
		Error: http.StatusText(http.StatusInternalServerError),
	}

	if err := Respond(ctx, w, er, http.StatusInternalServerError); err!= nil{
		return err
	}



	return nil
}