package captcha

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// HCaptchaValidation struct
type HCaptchaValidation struct {
	Success bool `json:"success"`
}

func ValidateHCaptcha(hCaptchaResponse string, hCaptchaSecret string) error {
	endpoint := "https://hcaptcha.com/siteverify"
	data := url.Values{}
	data.Set("secret", hCaptchaSecret)
	data.Set("response", hCaptchaResponse)

	client := &http.Client{}
	r, err := http.NewRequest("POST", endpoint, strings.NewReader(data.Encode())) // URL-encoded payload
	if err != nil {
		return err
	}

	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	res, err := client.Do(r)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	validation := &HCaptchaValidation{}
	err = json.Unmarshal(body, validation)
	if err != nil {
		return err
	}

	if !validation.Success {
		return errors.New("hcaptcha validation unsuccessful")
	}

	return nil
}
