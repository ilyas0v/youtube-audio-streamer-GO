package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
)

func main(){
	listenAddr := os.Getenv("LISTEN_ADDR")
	addr := listenAddr + `:` + os.Getenv("PORT")

	http.HandleFunc("/", index)
	http.HandleFunc("/watch", stream)

	log.Printf("Starting server at %s", addr)

	log.Fatal(http.ListenAndServe(addr, nil))
}

func index(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func stream(w http.ResponseWriter, r *http.Request) {
	v := r.URL.Query().Get("v")

	if v == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "use correct url format")
		return
	}

	err := downloadVideoAndExtractAudio(v, w)

	if err != nil {
		log.Printf("Stream error: %v", err)
		fmt.Fprintf(w, "Stream error: %v", err)
		return
	}
}

func downloadVideoAndExtractAudio(id string, out io.Writer) error {
	url := fmt.Sprintf("https://youtube.com/watch?v=" + id)

	reader, writer := io.Pipe()
	defer reader.Close()

	ytdl := exec.Command("youtube-dl", url, "-o-")

	ytdl.Stdout = writer
	ytdl.Stderr = os.Stderr

	ffmpeg := exec.Command("ffmpeg", "-i", "/dev/stdin", "-f", "mp3", "-ab", "96000", "-vn", "-")
	ffmpeg.Stdin = reader
	ffmpeg.Stdout = out
	ffmpeg.Stderr = os.Stderr

	go func() {
		if err := ytdl.Run(); err != nil {
			log.Printf("WARN: ytdl error: %v", err)
		}
	}()

	err := ffmpeg.Run()

	log.Printf("stream finished")

	return err

}