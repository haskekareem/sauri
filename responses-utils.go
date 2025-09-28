package sauri

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
)

const contentType = "Content-Type"

// Response struct holds the http.ResponseWriter and a map of headers
type Response struct {
	Rw http.ResponseWriter
	Hd http.Header
}

// NewResponse Initializes a new Response object.
func (s *Sauri) NewResponse() *Response {
	return &Response{
		Hd: make(http.Header),
	}
}

// WriteJSON sets the content type to JSON, marshals the data,
// and sends the response
func (s *Sauri) WriteJSON(w http.ResponseWriter, statusCode int, data interface{}, headers ...http.Header) error {

	// Marshal the data into JSON format
	content, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}

	// calling the header method here to add a single header
	w.Header().Set(contentType, "application/json")

	// Write the HTTP status code to the response
	w.WriteHeader(statusCode)

	// Write the response content
	_, err = w.Write(content)
	if err != nil {
		return err
	}

	return nil
}

func (s *Sauri) ReadJSON(w http.ResponseWriter, r *http.Request, data interface{}) error {
	maxByte := 1048576 // one megabyte
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxByte))

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(data); err != nil {
		return err
	}

	err := dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("invalid JSON: body must have a single json value")
	}

	return nil
}

// SetResponseWriter sets the http.ResponseWriter for the Response object
func (r *Response) SetResponseWriter(w http.ResponseWriter) *Response {
	r.Rw = w
	return r
}

// Header Sets a single header.
func (r *Response) Header(key, value string) *Response {
	r.Hd.Set(key, value)
	return r
}

// WithHeaders sets multiple headers at once by iterating over the
// provided map and adding each header to the Hd map.
func (r *Response) WithHeaders(headers http.Header) *Response {
	for key, values := range headers {
		for _, value := range values {
			r.Hd.Add(key, value)
		}
	}
	return r
}

// Send writes all headers and the content to the response.
// It sets the status code and then writes the content.
func (r *Response) Send(content []byte, statusCode int) error {

	for key, values := range r.Hd {
		for _, value := range values {
			r.Rw.Header().Add(key, value)
		}
	}

	// Write the HTTP status code to the response
	r.Rw.WriteHeader(statusCode)
	// Write the response content
	_, err := r.Rw.Write(content)
	if err != nil {
		return err
	}

	return nil
}

// JSON method sets the content type to JSON, marshals the data,
// and sends the response
func (r *Response) JSON(data interface{}, statusCode int) error {
	// Marshal the data into JSON format
	content, err := json.Marshal(data)
	if err != nil {
		http.Error(r.Rw, err.Error(), http.StatusInternalServerError)
		return err
	}

	// calling the header method here to add a single header
	r.Header(contentType, "application/json")

	// Send the JSON content with the given status code
	if err := r.Send(content, statusCode); err != nil {
		http.Error(r.Rw, err.Error(), http.StatusInternalServerError)
		return err
	}
	// otherwise return if everything works very well
	return nil
}

// XML method sets the content type to XML, marshals the data,
// and sends the response
func (r *Response) XML(data interface{}, statusCode int) error {
	// Marshal the data into XML format
	content, err := xml.Marshal(data)
	if err != nil {
		http.Error(r.Rw, err.Error(), http.StatusInternalServerError)
		return err
	}

	// calling the header method here to add a single header
	r.Header(contentType, "application/xml")

	// Send the XML content with the given status code
	if err := r.Send(content, statusCode); err != nil {
		http.Error(r.Rw, err.Error(), http.StatusInternalServerError)
		return err
	}
	// otherwise return if everything works very well
	return nil
}

// HTML method sets the content type to HTML and sends the response
func (r *Response) HTML(content string, status int) error {
	r.Header(contentType, "text/html")
	if err := r.Send([]byte(content), status); err != nil {
		http.Error(r.Rw, err.Error(), http.StatusInternalServerError)
		return err
	}
	return nil
}

// Redirect method sends an HTTP redirect to the client
func (r *Response) Redirect(url string, status int) error {
	r.Header("Location", url)
	if err := r.Send(nil, status); err != nil {
		http.Error(r.Rw, err.Error(), http.StatusInternalServerError)
		return err
	}

	return nil
}

// RedirectPermanent method sends a 301 Moved Permanently redirect
func (r *Response) RedirectPermanent(url string) error {
	err := r.Redirect(url, http.StatusMovedPermanently)
	if err != nil {
		http.Error(r.Rw, err.Error(), http.StatusInternalServerError)
		return err
	}

	return nil
}

// RedirectTemporary method sends a 302 Found redirect
func (r *Response) RedirectTemporary(url string) error {
	err := r.Redirect(url, http.StatusFound)
	if err != nil {
		http.Error(r.Rw, err.Error(), http.StatusInternalServerError)
		return err
	}

	return nil
}

// SetCORS sets CORS(Cross-Origin Resource Sharing)headers to allow all origins
func (r *Response) SetCORS() *Response {
	r.Header("Access-Control-Allow-Origin", "*")
	r.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	r.Header("Access-Control-Allow-Hd", "Content-Type, Authorization")
	return r
}

// SetCORSWithOrigin sets CORS(Cross-Origin Resource Sharing)headers to allow a specific origin
func (r *Response) SetCORSWithOrigin(origin string) *Response {
	r.Header("Access-Control-Allow-Origin", origin)
	r.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	r.Header("Access-Control-Allow-Hd", "Content-Type, Authorization")
	return r
}

// JSONP Sets the content type to JavaScript, wraps the JSON data in a callback
// function, and sends the response.
func (r *Response) JSONP(data interface{}, callback string, statusCode int) error {
	r.Header(contentType, "application/javascript")
	content, err := json.Marshal(data)
	if err != nil {
		http.Error(r.Rw, err.Error(), http.StatusInternalServerError)
		return err
	}

	// Wrap the JSON content in the callback function
	callBackContent := []byte(callback + "(" + string(content) + ");")

	if err := r.Send(callBackContent, statusCode); err != nil {
		http.Error(r.Rw, err.Error(), http.StatusInternalServerError)
		return err
	}

	return nil
}

// DownloadFile method sets headers for downloading a file and
// streams it to the client
func (r *Response) DownloadFile(pathToFile, fileName string, rr *http.Request) error {
	// Open the file specified by filePath
	filePath := path.Join(pathToFile, fileName)
	fileToServe := filepath.Clean(filePath)

	r.Rw.Header().Set("Content-Disposition", "attachment; filename=\""+fileName+"\"")

	http.ServeFile(r.Rw, rr, fileToServe)

	return nil
}

// StreamDownload method uses a callback function to stream data to the client
// as a download
func (r *Response) StreamDownload(callBack func(writer io.Writer), fileName string, headers map[string]string) error {
	r.Rw.Header().Set("Content-Disposition", "attachment; filename=\""+fileName+"\"")

	for key, value := range headers {
		r.Rw.Header().Set(key, value)
	}

	r.Rw.WriteHeader(http.StatusOK)
	// Execute the callback function, passing the ResponseWriter to stream the data
	callBack(r.Rw)

	return nil
}

// File method sets headers for displaying a file in the browser
// and streams it to the client
func (r *Response) File(fileRoad, fileName string, headers map[string]string) error {
	filePath := path.Join(fileRoad, fileName)
	fileToShow := filepath.Clean(filePath)

	file, err := os.Open(fileToShow)
	if err != nil {
		http.Error(r.Rw, "file not found", http.StatusInternalServerError)
		return err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	for key, value := range headers {
		r.Rw.Header().Set(key, value)
	}

	r.Rw.WriteHeader(http.StatusOK)

	if _, err := io.Copy(r.Rw, file); err != nil {
		http.Error(r.Rw, err.Error(), http.StatusInternalServerError)
		return err
	}
	return nil
}

// HandleFileUpload handles file uploads and saves them to the specified directory
func (r *Response) HandleFileUpload(fieldName, uploadDir string, req *http.Request) (string, error) {
	file, fileHeader, err := req.FormFile(fieldName)
	if err != nil {
		return "", err
	}
	defer func(file multipart.File) {
		_ = file.Close()
	}(file)

	filePath := filepath.Join(uploadDir, fileHeader.Filename)
	outFile, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer func(outFile *os.File) {
		_ = outFile.Close()
	}(outFile)

	if _, err := io.Copy(outFile, file); err != nil {
		http.Error(r.Rw, err.Error(), http.StatusInternalServerError)
		return "", err
	}
	return filePath, nil
}

// =================== errors for the response =================

func (r *Response) Error404() {
	r.errorStatus(http.StatusNotFound)
}

func (r *Response) Error500() {
	r.errorStatus(http.StatusInternalServerError)
}

func (r *Response) ErrorUnauthorized() {
	r.errorStatus(http.StatusUnauthorized)
}

func (r *Response) ErrorForbidden() {
	r.errorStatus(http.StatusForbidden)
}

func (r *Response) errorStatus(status int) {
	http.Error(r.Rw, http.StatusText(status), status)
}

// general errors for the gudu package

// Error404 returns page not found response
func (s *Sauri) Error404(w http.ResponseWriter, r *http.Request) {
	s.ErrorStatus(w, http.StatusNotFound)
}

// Error500 returns internal server error response
func (s *Sauri) Error500(w http.ResponseWriter, r *http.Request) {
	s.ErrorStatus(w, http.StatusInternalServerError)
}

// ErrorUnauthorized sends an unauthorized status (client is not known)
func (s *Sauri) ErrorUnauthorized(w http.ResponseWriter, r *http.Request) {
	s.ErrorStatus(w, http.StatusUnauthorized)
}

// ErrorForbidden returns a forbidden status message (client is known)
func (s *Sauri) ErrorForbidden(w http.ResponseWriter, r *http.Request) {
	s.ErrorStatus(w, http.StatusForbidden)
}

// ErrorStatus returns a response with the supplied http status
func (s *Sauri) ErrorStatus(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}
