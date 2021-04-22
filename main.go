package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/akhilrex/podgrab/controllers"
	"github.com/akhilrex/podgrab/client"
	"github.com/akhilrex/podgrab/db"
	"github.com/akhilrex/podgrab/service"
	"github.com/gin-contrib/location"
	"github.com/gin-gonic/gin"
	"github.com/jasonlvhit/gocron"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	var err error
	db.DB, err = db.Init()
	if err != nil {
		fmt.Println("statuse: ", err)
	} else {
		db.Migrate()
	}
	r := gin.Default()

	r.Use(setupSettings())
	r.Use(gin.Recovery())
	r.Use(location.Default())

	r.SetHTMLTemplate(client.MustLoadTemplate())

	pass := os.Getenv("PASSWORD")
	var router *gin.RouterGroup
	if pass != "" {
		router = r.Group("/", gin.BasicAuth(gin.Accounts{
			"podgrab": pass,
		}))
	} else {
		router = &r.RouterGroup
	}

	dataPath := os.Getenv("DATA")
	router.Static("/webassets", "./webassets")
	router.Static("/assets", dataPath)
	router.POST("/podcasts", controllers.AddPodcast)
	router.GET("/podcasts", controllers.GetAllPodcasts)
	router.GET("/podcasts/:id", controllers.GetPodcastById)
	router.DELETE("/podcasts/:id", controllers.DeletePodcastById)
	router.GET("/podcasts/:id/items", controllers.GetPodcastItemsByPodcastId)
	router.GET("/podcasts/:id/download", controllers.DownloadAllEpisodesByPodcastId)
	router.DELETE("/podcasts/:id/items", controllers.DeletePodcastEpisodesById)
	router.DELETE("/podcasts/:id/podcast", controllers.DeleteOnlyPodcastById)

	router.GET("/podcastitems", controllers.GetAllPodcastItems)
	router.GET("/podcastitems/:id", controllers.GetPodcastItemById)
	router.GET("/podcastitems/:id/image", controllers.GetPodcastItemImageById)
	router.GET("/podcastitems/:id/file", controllers.GetPodcastItemFileById)
	router.GET("/podcastitems/:id/markUnplayed", controllers.MarkPodcastItemAsUnplayed)
	router.GET("/podcastitems/:id/markPlayed", controllers.MarkPodcastItemAsPlayed)
	router.GET("/podcastitems/:id/bookmark", controllers.BookmarkPodcastItem)
	router.GET("/podcastitems/:id/unbookmark", controllers.UnbookmarkPodcastItem)
	router.PATCH("/podcastitems/:id", controllers.PatchPodcastItemById)
	router.GET("/podcastitems/:id/download", controllers.DownloadPodcastItem)
	router.GET("/podcastitems/:id/delete", controllers.DeletePodcastItem)

	router.GET("/tags", controllers.GetAllTags)
	router.GET("/tags/:id", controllers.GetTagById)
	router.DELETE("/tags/:id", controllers.DeleteTagById)
	router.POST("/tags", controllers.AddTag)
	router.POST("/podcasts/:id/tags/:tagId", controllers.AddTagToPodcast)
	router.DELETE("/podcasts/:id/tags/:tagId", controllers.RemoveTagFromPodcast)

	router.GET("/add", controllers.AddPage)
	router.GET("/search", controllers.Search)
	router.GET("/", controllers.HomePage)
	router.GET("/podcasts/:id/view", controllers.PodcastPage)
	router.GET("/episodes", controllers.AllEpisodesPage)
	router.GET("/allTags", controllers.AllTagsPage)
	router.GET("/settings", controllers.SettingsPage)
	router.POST("/settings", controllers.UpdateSetting)
	router.GET("/backups", controllers.BackupsPage)
	router.POST("/opml", controllers.UploadOpml)
	router.GET("/opml", controllers.GetOmpl)
	router.GET("/player", controllers.PlayerPage)

	r.GET("/ws", func(c *gin.Context) {
		controllers.Wshandler(c.Writer, c.Request)
	})
	go controllers.HandleWebsocketMessages()

	go assetEnv()
	go intiCron()

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")

}
func setupSettings() gin.HandlerFunc {
	return func(c *gin.Context) {

		setting := db.GetOrCreateSetting()
		c.Set("setting", setting)
		c.Writer.Header().Set("X-Clacks-Overhead", "GNU Terry Pratchett")

		c.Next()
	}
}

func intiCron() {
	checkFrequency, err := strconv.Atoi(os.Getenv("CHECK_FREQUENCY"))
	if err != nil {
		checkFrequency = 30
		log.Print(err)
	}
	service.UnlockMissedJobs()
	//gocron.Every(uint64(checkFrequency)).Minutes().Do(service.DownloadMissingEpisodes)
	gocron.Every(uint64(checkFrequency)).Minutes().Do(service.RefreshEpisodes)
	gocron.Every(uint64(checkFrequency)).Minutes().Do(service.CheckMissingFiles)
	gocron.Every(uint64(checkFrequency) * 2).Minutes().Do(service.UnlockMissedJobs)
	gocron.Every(uint64(checkFrequency) * 3).Minutes().Do(service.UpdateAllFileSizes)
	gocron.Every(uint64(checkFrequency)).Minutes().Do(service.DownloadMissingImages)
	gocron.Every(2).Days().Do(service.CreateBackup)
	<-gocron.Start()
}

func assetEnv() {
	log.Println("Config Dir: ", os.Getenv("CONFIG"))
	log.Println("Assets Dir: ", os.Getenv("DATA"))
	log.Println("Check Frequency (mins): ", os.Getenv("CHECK_FREQUENCY"))
}
