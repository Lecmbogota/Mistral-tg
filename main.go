package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "gopkg.in/yaml.v2"
    "io/ioutil"
    "net/http"
    "os"
)

type Config struct {
    Telegram struct {
        Token  string `yaml:"token"`
        APIURL string `yaml:"api_url"`
    } `yaml:"telegram"`
    Mistral struct {
        Token  string `yaml:"token"`
        Model  string `yaml:"model"`
        APIURL string `yaml:"api_url"`
    } `yaml:"mistral"`
    Server struct {
        Port string `yaml:"port"`
    } `yaml:"server"`
}

var config Config

// Cargar la configuración desde el archivo YAML
func loadConfig() error {
    data, err := ioutil.ReadFile("config.yml")
    if err != nil {
        return err
    }
    err = yaml.Unmarshal(data, &config)
    if err != nil {
        return err
    }
    return nil
}

// Estructuras para los mensajes de Telegram y la respuesta de Mistral
type Message struct {
    ChatID int64  `json:"chat_id"`
    Text   string `json:"text"`
}

type Chat struct {
    ID int64 `json:"id"`
}

type MessageReceived struct {
    Chat      Chat   `json:"chat"`
    Text      string `json:"text"`
    MessageID int64  `json:"message_id"`
}

type Update struct {
    UpdateID int             `json:"update_id"`
    Message  MessageReceived `json:"message"`
}

type MistralRequest struct {
    Model    string          `json:"model"`
    Messages []MessageContent `json:"messages"`
}

type MessageContent struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type MistralResponse struct {
    Choices []struct {
        Message MessageContent `json:"message"`
    } `json:"choices"`
}

// Función para enviar un mensaje a un chat de Telegram
func sendMessage(chatID int64, text string) error {
    url := fmt.Sprintf("%s%s/sendMessage", config.Telegram.APIURL, config.Telegram.Token)
    message := Message{
        ChatID: chatID,
        Text:   text,
    }
    jsonPayload, err := json.Marshal(message)
    if err != nil {
        return err
    }
    resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("Error en la respuesta de Telegram: %s", resp.Status)
    }
    return nil
}

// Función para obtener respuesta de Mistral
func getMistralResponse(messageText string) (string, error) {
    mistralRequest := MistralRequest{
        Model: config.Mistral.Model,
        Messages: []MessageContent{
            {
                Role:    "user",
                Content: messageText,
            },
        },
    }
    jsonPayload, err := json.Marshal(mistralRequest)
    if err != nil {
        return "", err
    }
    req, err := http.NewRequest("POST", config.Mistral.APIURL, bytes.NewBuffer(jsonPayload))
    if err != nil {
        return "", err
    }
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.Mistral.Token))
    req.Header.Set("Content-Type", "application/json")
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("Error en la respuesta de Mistral: %s", resp.Status)
    }
    var mistralResponse MistralResponse
    if err := json.NewDecoder(resp.Body).Decode(&mistralResponse); err != nil {
        return "", err
    }
    if len(mistralResponse.Choices) > 0 {
        return mistralResponse.Choices[0].Message.Content, nil
    }
    return "No hay respuesta", nil
}

// Handler para recibir mensajes de Telegram
func receiveMessageHandler(w http.ResponseWriter, r *http.Request) {
    var update Update
    if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Procesar el mensaje recibido
    chatID := update.Message.Chat.ID
    messageText := update.Message.Text
    fmt.Printf("Mensaje recibido de chatID %d: %s\n", chatID, messageText)

    // Obtener respuesta de Mistral
    replyText, err := getMistralResponse(messageText)
    if err != nil {
        fmt.Println("Error obteniendo respuesta de Mistral:", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Enviar la respuesta al usuario
    err = sendMessage(chatID, replyText)
    if err != nil {
        fmt.Println("Error enviando mensaje:", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Responder con éxito
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Mensaje recibido y respondido."))
}

// Configura el servidor HTTP
func setupServer() {
    http.HandleFunc("/webhook", receiveMessageHandler)
    fmt.Printf("Servidor escuchando en el puerto %s...\n", config.Server.Port)
    if err := http.ListenAndServe(":" + config.Server.Port, nil); err != nil {
        fmt.Println("Error al iniciar el servidor:", err)
        os.Exit(1)
    }
}

func main() {
    // Cargar la configuración desde el archivo YAML
    if err := loadConfig(); err != nil {
        fmt.Println("Error cargando configuración:", err)
        os.Exit(1)
    }

    setupServer()
}
