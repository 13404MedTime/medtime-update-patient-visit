package function

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Datas This is response struct from create
type Datas struct {
	Data struct {
		Data struct {
			Data map[string]interface{} `json:"data"`
		} `json:"data"`
	} `json:"data"`
}

// ClientApiResponse This is get single api response
type ClientApiResponse struct {
	Data ClientApiData `json:"data"`
}

type ClientApiData struct {
	Data ClientApiResp `json:"data"`
}

type ClientApiResp struct {
	Response map[string]interface{} `json:"response"`
}

type Response struct {
	Status string                 `json:"status"`
	Data   map[string]interface{} `json:"data"`
}

// NewRequestBody's Data (map) field will be in this structure
//.   fields
// objects_ids []string
// table_slug string
// object_data map[string]interface
// method string
// app_id string

// but all field will be an interface, you must do type assertion

type HttpRequest struct {
	Method  string      `json:"method"`
	Path    string      `json:"path"`
	Headers http.Header `json:"headers"`
	Params  url.Values  `json:"params"`
	Body    []byte      `json:"body"`
}

type AuthData struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

type NewRequestBody struct {
	RequestData HttpRequest            `json:"request_data"`
	Auth        AuthData               `json:"auth"`
	Data        map[string]interface{} `json:"data"`
}
type Request struct {
	Data map[string]interface{} `json:"data"`
}

// GetListClientApiResponse This is get list api response
type GetListClientApiResponse struct {
	Data GetListClientApiData `json:"data"`
}

type GetListClientApiData struct {
	Data GetListClientApiResp `json:"data"`
}

type GetListClientApiResp struct {
	Response []map[string]interface{} `json:"response"`
}

func DoRequest(url string, method string, body interface{}, appId string) ([]byte, error) {
	data, err := json.Marshal(&body)
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Timeout: time.Duration(5 * time.Second),
	}
	request, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	request.Header.Add("authorization", "API-KEY")
	request.Header.Add("X-API-KEY", appId)

	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respByte, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return respByte, nil
}

func Send(text string) {
	bot, _ := tgbotapi.NewBotAPI("6241555505:AAHPpkXj-oHBGblWd_7O9kxc9a05tJUIFRw")

	msg := tgbotapi.NewMessage(1194897882, text)

	bot.Send(msg)
}

// Handle a serverless request
func Handle(req []byte) string {
	var response Response
	var request NewRequestBody
	const urlConst = "https://api.admin.u-code.io"

	err := json.Unmarshal(req, &request)
	if err != nil {
		response.Data = map[string]interface{}{"message": "Error while unmarshalling request"}
		response.Status = "error"
		responseByte, _ := json.Marshal(response)
		return string(responseByte)
	}
	if request.Data["app_id"] == nil {
		response.Data = map[string]interface{}{"message": "App id required"}
		response.Status = "error"
		responseByte, _ := json.Marshal(response)
		return string(responseByte)
	}
	appId := request.Data["app_id"].(string)

	updateReq := request.Data["object_data"].(map[string]interface{})

	var tableSlug = "naznachenie"
	// create patient visits
	if request.Data["method"].(string) == "UPDATE" {
		oldNaznacheniyaData, err, response := GetSingleObject(urlConst, tableSlug, appId, updateReq["guid"].(string))
		if err != nil {
			responseByte, _ := json.Marshal(response)
			return string(responseByte)
		}

		if updateReq["cleints_id"].(string) == oldNaznacheniyaData.Data.Data.Response["cleints_id"].(string) && updateReq["doctor_id"].(string) == oldNaznacheniyaData.Data.Data.Response["doctor_id"].(string) {
		} else {
			tableSlug = "patient_visits"
			//get list objects response example
			getListObjectRequest := Request{
				// some filters
				Data: map[string]interface{}{
					"doctor_id":  oldNaznacheniyaData.Data.Data.Response["doctor_id"].(string),
					"cleints_id": oldNaznacheniyaData.Data.Data.Response["cleints_id"].(string),
				},
			}
			patientVisits, err, response := GetListObject(urlConst, tableSlug, appId, getListObjectRequest)
			if err != nil {
				responseByte, _ := json.Marshal(response)
				return string(responseByte)
			}

			reqBody := map[string]interface{}{
				"id_from":    patientVisits.Data.Data.Response[0]["guid"].(string),
				"id_to":      []string{updateReq["guid"].(string)},
				"table_from": "patient_visits",
				"table_to":   "naznachenie",
			}
			// fmt.Println("reqbody", reqBody)
			// testdat, _ := json.Marshal(reqBody)
			// Send(string(testdat))
			_, err = DoRequest("https://api.admin.u-code.io/v1/many-to-many?from-ofs=true?project-id=a4dc1f1c-d20f-4c1a-abf5-b819076604bc", "DELETE", reqBody, appId)
			if err != nil {
				response.Data = map[string]interface{}{"message": err.Error()}
				response.Status = "error"
				responseByte, _ := json.Marshal(response)
				return string(responseByte)
			}

			tableSlug = "patient_visits"
			//get list objects response example
			getListObjectRequest = Request{
				// some filters
				Data: map[string]interface{}{
					"doctor_id":  updateReq["doctor_id"].(string),
					"cleints_id": updateReq["cleints_id"].(string),
				},
			}
			patientVisits, err, response = GetListObject(urlConst, tableSlug, appId, getListObjectRequest)
			if err != nil {
				responseByte, _ := json.Marshal(response)
				return string(responseByte)
			}

			if len(patientVisits.Data.Data.Response) < 1 {

				//create objects response example
				createtObjectRequest := Request{
					// some filters
					Data: map[string]interface{}{
						"doctor_id":  updateReq["doctor_id"].(string),
						"cleints_id": updateReq["cleints_id"].(string),
						"date":       updateReq["created_time"].(string),
					},
				}

				createResp, err, response := CreateObject(urlConst, tableSlug, appId, createtObjectRequest)
				if err != nil {
					responseByte, _ := json.Marshal(response)
					return string(responseByte)
				}

				body := map[string]interface{}{
					"id_from":    createResp.Data.Data.Data["guid"].(string),
					"id_to":      []string{updateReq["guid"].(string)},
					"table_from": "patient_visits",
					"table_to":   "naznachenie",
				}
				// testdat, _ := json.Marshal(body)
				// Send(string(testdat))
				_, err = DoRequest("https://api.admin.u-code.io/v1/many-to-many?from-ofs=true?project-id=a4dc1f1c-d20f-4c1a-abf5-b819076604bc", "PUT", body, appId)
				if err != nil {
					response.Data = map[string]interface{}{"message": err.Error()}
					response.Status = "error"
					responseByte, _ := json.Marshal(response)
					return string(responseByte)
				}
			} else if len(patientVisits.Data.Data.Response) == 1 {

				tableSlug = "patient_visits"
				patientVisitsId := patientVisits.Data.Data.Response[0]["guid"].(string)

				updateRequest := Request{
					Data: map[string]interface{}{
						"guid": patientVisitsId,
						"date": updateReq["created_time"].(string),
					},
				}
				err, response = UpdateObject(urlConst, tableSlug, appId, updateRequest)
				if err != nil {
					responseByte, _ := json.Marshal(response)
					return string(responseByte)
				}
				body := map[string]interface{}{
					"id_from":    patientVisits.Data.Data.Response[0]["guid"].(string),
					"id_to":      []string{updateReq["guid"].(string)},
					"table_from": "patient_visits",
					"table_to":   "naznachenie",
				}
				// testdat, _ := json.Marshal(body)
				// Send(string(testdat))
				_, err := DoRequest("https://api.admin.u-code.io/v1/many-to-many?from-ofs=true?project-id=a4dc1f1c-d20f-4c1a-abf5-b819076604bc", "PUT", body, appId)
				if err != nil {
					response.Data = map[string]interface{}{"message": err.Error()}
					response.Status = "error"
					responseByte, _ := json.Marshal(response)
					return string(responseByte)
				}
				// Send("create " + string(resp))

			}

		}

	}

	if request.Data["method"].(string) == "DELETE" {

		var tableSlug = "naznachenie"
		naznacheniyaData, err, response := GetSingleObject(urlConst, tableSlug, appId, updateReq["id"].(string))
		if err != nil {
			responseByte, _ := json.Marshal(response)
			return string(responseByte)
		}
		tableSlug = "patient_visits"
		//get list objects response example
		getListObjectRequest := Request{
			// some filters
			Data: map[string]interface{}{
				"doctor_id":  naznacheniyaData.Data.Data.Response["doctor_id"].(string),
				"cleints_id": naznacheniyaData.Data.Data.Response["cleints_id"].(string),
			},
		}
		patientVisits, err, response := GetListObject(urlConst, tableSlug, appId, getListObjectRequest)
		if err != nil {
			responseByte, _ := json.Marshal(response)
			return string(responseByte)
		}

		if len(patientVisits.Data.Data.Response) > 0 {
			reqBody := map[string]interface{}{
				"id_from":    patientVisits.Data.Data.Response[0]["guid"].(string),
				"id_to":      []string{updateReq["id"].(string)},
				"table_from": "patient_visits",
				"table_to":   "naznachenie",
			}

			_, err = DoRequest("https://api.admin.u-code.io/v1/many-to-many?from-ofs=true?project-id=a4dc1f1c-d20f-4c1a-abf5-b819076604bc", "DELETE", reqBody, appId)
			if err != nil {
				response.Data = map[string]interface{}{"message": err.Error()}
				response.Status = "error"
				responseByte, _ := json.Marshal(response)
				return string(responseByte)
			}

		}

		// --------------------------------------------------------------------------------------------------------------------------------------------------
		// delete admin reports
		tableSlug = "report_for_admin"

		newReq := Request{
			Data: map[string]interface{}{
				"naznachenie_id": updateReq["id"].(string),
				"limit":          1,
			},
		}

		reportData, err, response := GetListObject(urlConst, tableSlug, appId, newReq)
		if err != nil {
			response.Data = map[string]interface{}{"message": err.Error()}
			response.Status = "error"
			responseByte, _ := json.Marshal(response)
			return string(responseByte)
		}
		if len(reportData.Data.Data.Response) > 0 {
			err, response = DeleteObject(urlConst, tableSlug, appId, reportData.Data.Data.Response[0]["guid"].(string))
			if err != nil {
				response.Data = map[string]interface{}{"message": err.Error()}
				response.Status = "error"
				responseByte, _ := json.Marshal(response)
				return string(responseByte)
			}
		}
		// --------------------------------------------------------------------------------------------------------------------------------------------------
		// delete doctor reports
		tableSlug = "report_for_doctor"

		newReq = Request{
			Data: map[string]interface{}{
				"naznachenie_id": updateReq["id"].(string),
				"limit":          1,
			},
		}

		doctorReportData, err, response := GetListObject(urlConst, tableSlug, appId, newReq)
		if err != nil {
			response.Data = map[string]interface{}{"message": err.Error()}
			response.Status = "error"
			responseByte, _ := json.Marshal(response)
			return string(responseByte)
		}
		if len(reportData.Data.Data.Response) > 0 {
			err, response = DeleteObject(urlConst, tableSlug, appId, doctorReportData.Data.Data.Response[0]["guid"].(string))
			if err != nil {
				response.Data = map[string]interface{}{"message": err.Error()}
				response.Status = "error"
				responseByte, _ := json.Marshal(response)
				return string(responseByte)
			}
		}

	}

	response.Data = map[string]interface{}{}
	response.Status = "done" //if all will be ok else "error"
	responseByte, _ := json.Marshal(response)

	return string(responseByte)
}

func GetListObject(url, tableSlug, appId string, request Request) (GetListClientApiResponse, error, Response) {
	response := Response{}

	getListResponseInByte, err := DoRequest(url+"/v1/object/get-list/"+tableSlug+"?from-ofs=true&project-id=a4dc1f1c-d20f-4c1a-abf5-b819076604bc", "POST", request, appId)
	if err != nil {
		response.Data = map[string]interface{}{"message": "Error while getting single object"}
		response.Status = "error"
		return GetListClientApiResponse{}, errors.New("error"), response
	}
	var getListObject GetListClientApiResponse
	err = json.Unmarshal(getListResponseInByte, &getListObject)
	if err != nil {
		response.Data = map[string]interface{}{"message": "Error while unmarshalling get list object"}
		response.Status = "error"
		return GetListClientApiResponse{}, errors.New("error"), response
	}
	return getListObject, nil, response
}
