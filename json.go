package webmod

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// JSONResponse is the type used for sending around JSON
type JSONResponse struct {
	Error   bool        `json:"error"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ReadJson tries to read the body of a request and
// converts from JSON to go data variable
func (t *Tools) ReadJSON(w http.ResponseWriter, r *http.Request, jdata interface{}) error {
	// limit of the size of the data that can be received
	maxBytes := 1024 * 1024
	if t.MaxJSONSize != 0 {
		maxBytes = t.MaxJSONSize
	}

	// read the body, limiting the amount of data that can be read -
	// http.MaxBytesReader
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	// get a JSON decoder
	jdec := json.NewDecoder(r.Body)
	// disallow unknown fields (optional)
	if !t.AllowUnknownFields {
		jdec.DisallowUnknownFields()
	}
	// decode JSON
	err := jdec.Decode(jdata)
	if err != nil {
		// handle various kinds of errors
		// related to decoding json decoding
		// in human redable form.
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		// syntax error
		case errors.As(err, &syntaxError):
			return fmt.Errorf("Syntax error: (at character: %d)", syntaxError.Offset)

		// JSON ended unexpectedly
		case errors.Is(err, io.ErrUnexpectedEOF):
			return fmt.Errorf("JSON ended unexpectedly, badly-formed JSON")

		// unexpected field type
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("Invalid JSON type for the field: %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("Invalid JSON type (at character: %d)", unmarshalTypeError.Offset)

		// empty body received
		case errors.Is(err, io.EOF):
			return errors.New("Empty body")

		// unknown field received
		case strings.HasPrefix(err.Error(), "json: unknown field"):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field")
			return fmt.Errorf("JSON contains unknown key: %s", fieldName)

		// body too large
		case err.Error() == "http: request body too large":
			return fmt.Errorf("JSON too big! must be limited to %d bytes", maxBytes)

		// unable to unmarshal
		case errors.As(err, &invalidUnmarshalError):
			return fmt.Errorf("Error unmarshaling JSON: %s", err.Error())
		default:
			return err
		}
	}

	// check for more than one json value
	// throw error, if found
	err = jdec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("Body must contain only one JSON value")
	}

	return nil
}

// WriteJson takes a ResponseWriter, Status, Data, and Headers and
// writes JSON to client
func (t *Tools) WriteJSON(w http.ResponseWriter, status int, jdata interface{}, headers ...http.Header) error {
	// encode data into json
	out, err := json.Marshal(jdata)
	if err != nil {
		return err
	}

	// add provided headers to writer
	if len(headers) > 0 {
		for k, v := range headers[0] {
			w.Header()[k] = v
		}
	}
	// set 'Content-Type' header and http status to header
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	// write json
	_, err = w.Write(out)
	if err != nil {
		return err
	}
	return nil
}

// ErrorJSON is a utility function to easily write errors to client in JSON format.
// It optionally takes status code as an argument
func (t *Tools) ErrorJSON(w http.ResponseWriter, err error, status ...int) error {
	statusCode := http.StatusBadRequest
	if len(status) > 0 {
		statusCode = status[0]
	}

	jdata := JSONResponse{
		Error:   true,
		Message: err.Error(),
	}

	return t.WriteJSON(w, statusCode, jdata)
}
