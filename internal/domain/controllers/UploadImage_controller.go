package controllers

import (
    "bytes"
    "fmt"
    "io"
    "mime/multipart"
    "net/http"
    "os"
)

func UploadImageToSupabase(fileHeader *multipart.FileHeader) (string, error) {
    // Buka file yang diunggah
    file, err := fileHeader.Open()
    if err != nil {
        return "", fmt.Errorf("failed to open file: %w", err)
    }
    defer file.Close()

    // Baca konten file ke buffer
    buffer := &bytes.Buffer{}
    writer := multipart.NewWriter(buffer)

    part, err := writer.CreateFormFile("file", fileHeader.Filename)
    if err != nil {
        return "", fmt.Errorf("failed to create form file: %w", err)
    }

    // Salin file ke form
    _, err = io.Copy(part, file)
    if err != nil {
        return "", fmt.Errorf("failed to copy file: %w", err)
    }

    // Selesaikan pembuatan multipart form
    writer.Close()

    // Endpoint Supabase Storage
    supabaseURL := os.Getenv("SUPABASE_URL")           // Contoh: "https://<your-project>.supabase.co"
    supabaseBucket := os.Getenv("SUPABASE_BUCKET")     // Contoh: "trashure-images"
    supabaseAPIKey := os.Getenv("SUPABASE_API_KEY")    // API Key untuk Supabase

    uploadURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", supabaseURL, "upload", supabaseBucket)

    // Membuat permintaan HTTP POST
    req, err := http.NewRequest("POST", uploadURL, buffer)
    if err != nil {
        return "", fmt.Errorf("failed to create request: %w", err)
    }

    // Menambahkan header
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", supabaseAPIKey))
    req.Header.Set("Content-Type", writer.FormDataContentType())

    // Kirim permintaan
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return "", fmt.Errorf("failed to execute request: %w", err)
    }
    defer resp.Body.Close()

    // Periksa apakah unggahan berhasil
    if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
        body, _ := io.ReadAll(resp.Body)
        return "", fmt.Errorf("failed to upload image: %s", string(body))
    }

    // Membuat URL gambar publik (opsional jika bucket bersifat public)
    publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", supabaseURL, supabaseBucket, fileHeader.Filename)
    return publicURL, nil
}
