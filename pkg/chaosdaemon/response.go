package chaosdaemon

const (
	// StatusOK represents OK status code
	StatusOK = 200
	// StatusOtherError represents Error status code
	StatusOtherError = 1
)

// Response is the body part of HTTP Response
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}
