package client

import (
	"fmt"
	"time"
	"html/template"
	"strings"
	"io/ioutil"
	"log"

	"github.com/akhilrex/podgrab/service"
	"github.com/akhilrex/podgrab/db"
	_ "github.com/jessevdk/go-assets"
)

//go:generate go-assets-builder ./ -s / -o content.go -p client -v Templates

var funcMap = template.FuncMap{
	"intRange": func(start, end int) []int {
		n := end - start + 1
		result := make([]int, n)
		for i := 0; i < n; i++ {
			result[i] = start + i
		}
		return result
	},
	"removeStartingSlash": func(raw string) string {
		fmt.Println(raw)
		if string(raw[0]) == "/" {
			return raw
		}
		return "/" + raw
	},
	"isDateNull": func(raw time.Time) bool {
		return raw == (time.Time{})
	},
	"formatDate": func(raw time.Time) string {
		if raw == (time.Time{}) {
			return ""
		}

		return raw.Format("Jan 2 2006")
	},
	"naturalDate": func(raw time.Time) string {
		return service.NatualTime(time.Now(), raw)
		//return raw.Format("Jan 2 2006")
	},
	"latestEpisodeDate": func(podcastItems []db.PodcastItem) string {
		var latest time.Time
		for _, item := range podcastItems {
			if item.PubDate.After(latest) {
				latest = item.PubDate
			}
		}
		return latest.Format("Jan 2 2006")
	},
	"downloadedEpisodes": func(podcastItems []db.PodcastItem) int {
		count := 0
		for _, item := range podcastItems {
			if item.DownloadStatus == db.Downloaded {
				count++
			}
		}
		return count
	},
	"downloadingEpisodes": func(podcastItems []db.PodcastItem) int {
		count := 0
		for _, item := range podcastItems {
			if item.DownloadStatus == db.NotDownloaded {
				count++
			}
		}
		return count
	},
	"formatFileSize": func(inputSize int64) string {
		size := float64(inputSize)
		const divisor float64 = 1024
		if size < divisor {
			return fmt.Sprintf("%.0f bytes", size)
		}
		size = size / divisor
		if size < divisor {
			return fmt.Sprintf("%.2f KB", size)
		}
		size = size / divisor
		if size < divisor {
			return fmt.Sprintf("%.2f MB", size)
		}
		size = size / divisor
		if size < divisor {
			return fmt.Sprintf("%.2f GB", size)
		}
		size = size / divisor
		return fmt.Sprintf("%.2f TB", size)
	},
	"formatDuration": func(total int) string {
		if total <= 0 {
			return ""
		}
		mins := total / 60
		secs := total % 60
		hrs := 0
		if mins >= 60 {
			hrs = mins / 60
			mins = mins % 60
		}
		if hrs > 0 {
			return fmt.Sprintf("%02d:%02d:%02d", hrs, mins, secs)
		}
		return fmt.Sprintf("%02d:%02d", mins, secs)
	},
}

func MustLoadTemplate() *template.Template {
	tmpl := template.New("")
	for name, file := range Assets.Files {
		if file.IsDir() || !strings.HasSuffix(name, ".html") {
			continue
		}
		h, err := ioutil.ReadAll(file)
		if err != nil {
			log.Fatal(err)
		}
		tmpl, err = tmpl.New(name).Funcs(funcMap).Parse(string(h))
		if err != nil {
			log.Fatal(err)
		}
	}
	return tmpl
}
