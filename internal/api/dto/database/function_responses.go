package database

type FunctionResponse struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	DataType   string `json:"dataType"`
	Definition string `json:"definition"`
	Language   string `json:"language"`
}
