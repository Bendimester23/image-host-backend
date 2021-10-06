package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/Bendimester23/image-host/db"
	"github.com/gin-gonic/gin"
)

var (
	chars          = strings.Split("abcdefghijklmnopqrstuovwxyz1234567890ABCDEFGHIJKLMNOPQRSTUOVWXYZ", "")
	id_lenght      = flag.Int("id-lenght", 10, "Image random id lenght")
	base_url       = flag.String("base-url", "http://localhost:3000/", "Url to copy excluding id")
	img_path       = flag.String("img-path", "/i/:id", "Image get path.")
	port           = flag.Int("port", 3000, "Server port")
	admin_password = flag.String("admin-password", "admin", "Password for user creation")
)

func GetRandomString(lenght int) string {
	res := ""
	for i := 0; i < lenght; i++ {
		res = fmt.Sprintf("%s%s", res, chars[rand.Intn(len(chars)-1)])
	}
	return res
}

func main() {
	log.Println("Starting...")

	flag.Parse()
	gin.SetMode(gin.ReleaseMode)
	rand.Seed(time.Now().Unix())

	client := db.NewClient()

	if client.Connect() != nil {
		log.Fatalln("Error connecting!")
	}

	defer func() {
		if client.Prisma.Disconnect() != nil {
			log.Fatalln("Error disconnecting!")
		}
	}()

	ctx := context.Background()

	r := gin.Default()

	r.POST("/register", func(c *gin.Context) {
		if c.GetHeader("Authorization") != *admin_password {
			c.JSON(401, gin.H{
				"error": "unauthorized",
			})
			return
		}

		d, _ := c.GetRawData()

		if len(d) < 3 {
			c.JSON(400, gin.H{
				"error": "short name",
			})
			return
		}

		uu, _ := client.User.FindFirst(
			db.User.Name.Equals(string(d)),
		).Exec(ctx)

		if uu != nil {
			c.JSON(409, gin.H{
				"error": "already exists",
			})
			return
		}

		u, err := client.User.CreateOne(
			db.User.Token.Set(GetRandomString(30)),
			db.User.Name.Set(string(d)),
		).Exec(ctx)

		if err != nil {
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(201, gin.H{
			"token": u.Token,
		})
	})

	r.POST("/upload", func(c *gin.Context) {
		token := c.Query("token")

		if token == "" {
			c.JSON(401, gin.H{
				"error": "unauthorized",
			})
			return
		}

		u, err := client.User.FindFirst(
			db.User.Token.Equals(token),
		).Exec(ctx)

		if err != nil {
			c.JSON(401, gin.H{
				"error": "unauthorized",
			})
			return
		}

		log.Println(u)

		id := GetRandomString(*id_lenght)
		fo, _ := os.Create(fmt.Sprintf("./images/%s.png", id))

		defer fo.Close()

		io.Copy(fo, c.Request.Body)

		i, err := client.Image.CreateOne(
			db.Image.ID.Set(id),
			db.Image.User.Link(
				db.User.Token.Equals(token),
			),
		).Exec(ctx)

		c.JSON(200, gin.H{
			"url": fmt.Sprintf("%s%s", *base_url, i.ID),
		})
	})

	r.GET(*img_path, func(c *gin.Context) {
		id := c.Param("id")

		f, err := os.Open(fmt.Sprintf("./images/%s.png", id))
		defer func(f *os.File) {
			err := f.Close()
			if err != nil {
				log.Printf("ERROR:\n%s\n", err.Error())
			}
		}(f)
		if err != nil {
			c.JSON(404, gin.H{
				"error": "image not found",
			})
			return
		}
		c.Status(200)
		c.Header("Content-Type", "image/png")

		io.Copy(c.Writer, f)
	})

	log.Printf("Listening on port %d.", *port)

	r.Run(fmt.Sprintf(":%d", *port))
}
