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
	fmt.Fprint(w, `
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta http-equiv="X-UA-Compatible" content="IE=edge">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>Yotube audio streaming</title>
	</head>
	<body>
		<h1>Youtube audio</h1>
		<form action="" onsubmit="extractVideoId(event)">
			<label for="">Video url:</label>
			<input type="text" name="v" id="v">
			<input type="submit" value="Convert" id="btn">
		</form>

		<script>

			function youtubeParser(url){
				var regExp = /^https?\:\/\/(?:www\.youtube(?:\-nocookie)?\.com\/|m\.youtube\.com\/|youtube\.com\/)?(?:ytscreeningroom\?vi?=|youtu\.be\/|vi?\/|user\/.+\/u\/\w{1,2}\/|embed\/|watch\?(?:.*\&)?vi?=|\&vi?=|\?(?:.*\&)?vi?=)([^#\&\?\n\/<>"']*)/i;
				var match = url.match(regExp);
				return (match && match[1].length==11)? match[1] : false;
			}

			function extractVideoId(event)
			{
				event.preventDefault();
				var input = document.getElementById("v");
				var url = input.value;
				var btn = document.getElementById("btn");
				
				var id = youtubeParser(url);

				if(id)
				{
					window.location.href = "/watch?v=" + id;
					input.disabled = true;
					btn.disabled = true;
				}
			}
		</script>
	</body>
	</html>`)
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