package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
)

const (
    telegramAPIURL = "https://api.telegram.org/bot"
    mistralAPIURL  = "https://api.mistral.ai/v1/chat/completions"
)

// Token del bot de Telegram
var telegramToken = "7925086407:AAHihC5znaPNcjM05qw1SqF7KUDRKld4Cwo" // Reemplaza por tu token real de Telegram
// Token de la API de Mistral
var mistralToken = "8iGHLzq3qmzDOOuqexG0e8W4ZI9Qm9Od" // Reemplaza por tu token real de Mistral

// Estructuras para el mensaje de Telegram y la respuesta de Mistral
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
    Message  MessageReceived  `json:"message"`
}

type MistralRequest struct {
    Model   string          `json:"model"`
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
    url := fmt.Sprintf("%s%s/sendMessage", telegramAPIURL, telegramToken)

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
        Model: "mistral-small-latest",
        //Model: "ag:5a8440b4:20241104:untitled-agent:f2a8b854",
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

    req, err := http.NewRequest("POST", mistralAPIURL, bytes.NewBuffer(jsonPayload))
    if err != nil {
        return "", err
    }

    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", mistralToken))
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
    fmt.Println("Servidor escuchando en el puerto 9020...")
    if err := http.ListenAndServe(":9020", nil); err != nil {
        fmt.Println("Error al iniciar el servidor:", err)
        os.Exit(1)
    }
}

func main() {
    setupServer()
}
