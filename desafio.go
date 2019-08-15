package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
    "net/http"
    "os"
	"path/filepath"
    "strings"
    "time"
)

type JsonContent struct {
	NumbersCiphered int `json:"numero_casas"`
	Token string `json:"token"`
	Ciphered string `json:"cifrado"`
	Deciphered string `json:"decifrado"`
	CryptographicText string `json:"resumo_criptografico"`
}

func main() {
	var jsonContent JsonContent
	token := "fc96690653c6c61ff1f16e1c0464fa7033082552"

    getJson(token, &jsonContent)
	
	decryptsJulioCesar(&jsonContent)

    encryptSha1(&jsonContent)

    createOrEditJsonFile(jsonContent)

    postJsonFile(token)
}

func getJson(token string, jsonContent *JsonContent) {
    url := "https://api.codenation.dev/v1/challenge/dev-ps/generate-data?token="+token
    response, errGet := http.Get(url)
    if errGet != nil {
    	fmt.Printf("Error: %s", errGet)
        return
    }
    
    data, errReadAll := ioutil.ReadAll(response.Body)
    if errReadAll != nil {
    	fmt.Printf("Error: %s", errGet)
        return
    }

	errUnmarshal := json.Unmarshal(data, &jsonContent)
    if errUnmarshal != nil {
        fmt.Printf("Error: %s", errUnmarshal)
        return
    }
}

func createOrEditJsonFile(jsonContent JsonContent) {
	result, errMarshal := json.Marshal(jsonContent)
    if errMarshal != nil {
        fmt.Printf("Error: %s", errMarshal)
        return
    }

    errWriteFile := ioutil.WriteFile("answer.json", result, 0644)
    if errWriteFile != nil {
        fmt.Printf("Error: %s", errWriteFile)
        return
    }
}

func decryptsJulioCesar(jsonContent *JsonContent) {
    alphabet := "abcdefghijklmnopqrstuvwxyz"
    characters := strings.Split(jsonContent.Ciphered, "")
    deciphered_text := ""

    for _, character := range characters {
        if strings.Contains(alphabet, character) {
            character_position := strings.Index(alphabet, character)
           	deciphered_character_position := character_position - jsonContent.NumbersCiphered
            if(deciphered_character_position < 0) {
                deciphered_character_position = 26 + deciphered_character_position
            }
            character = string(alphabet[deciphered_character_position])
        }
        deciphered_text += character
    }

    jsonContent.Deciphered = deciphered_text
}

func encryptSha1 (jsonContent *JsonContent) {
    content := sha1.New()
    io.WriteString(content, jsonContent.Deciphered)

    cryptographic_text := fmt.Sprintf("% x", content.Sum(nil))
    cryptographic_text = strings.ReplaceAll(cryptographic_text, " ", "")
    jsonContent.CryptographicText = cryptographic_text
}

func postJsonFile(token string) {
	url := "https://api.codenation.dev/v1/challenge/dev-ps/submit-solution?token="+token
	paramName := "answer"
	file := "answer.json"

	request, errFileUpload := newfileUploadRequest(url, paramName, file)
	if errFileUpload != nil {
		fmt.Printf("Error: %s", errFileUpload)
        return
	}

	time.Sleep(65 * time.Second)

	client := &http.Client{}
	resp, errDo := client.Do(request)
	if errDo != nil {
		fmt.Printf("Error: %s", errDo)
        return
	}

	body := &bytes.Buffer{}
	_, errReadForm := body.ReadFrom(resp.Body)
    if errReadForm != nil {
		fmt.Printf("Error: %s", errReadForm)
        return
	}
	
	resp.Body.Close()
	fmt.Println(body)
}

func newfileUploadRequest(url, paramName, path string) (*http.Request, error) {
	
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, filepath.Base(path))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, err
}