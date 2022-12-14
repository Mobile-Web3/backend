package http

type apiResponse struct {
	IsSuccess bool        `json:"isSuccess"`
	Error     string      `json:"error"`
	Result    interface{} `json:"result"`
}

func newSuccessResponse(result interface{}) apiResponse {
	return apiResponse{
		IsSuccess: true,
		Result:    result,
	}
}

func newErrorResponse(error string) apiResponse {
	return apiResponse{
		IsSuccess: false,
		Error:     error,
	}
}
