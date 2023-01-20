package backend

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql" // mysql driver
)

func httpJSONPost(endpoint string, data io.Reader) (*http.Response, []byte, error) {
	req, err := http.NewRequest("POST", endpoint, data)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	return resp, body, nil
}

func httpJSONGet(endpoint string) (*http.Response, []byte, error) {
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	return resp, body, nil
}

func getOrders(db *sql.DB) (string, error) {
	rows, err := db.QueryContext(context.Background(), "select * from orders order by id desc limit 5000")
	if err != nil {
		return "", err
	}

	// https://stackoverflow.com/a/60386531 - database/sql rows to json
	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return "", err
	}

	count := len(columnTypes)
	finalRows := []interface{}{}

	for rows.Next() {
		scanArgs := make([]interface{}, count)

		for i, v := range columnTypes {
			switch v.DatabaseTypeName() {
			case "VARCHAR", "TEXT", "UUID", "TIMESTAMP":
				scanArgs[i] = new(sql.NullString)
			case "BOOL":
				scanArgs[i] = new(sql.NullBool)
			case "INT4":
				scanArgs[i] = new(sql.NullInt64)
			default:
				scanArgs[i] = new(sql.NullString)
			}
		}

		err := rows.Scan(scanArgs...)

		if err != nil {
			log.Fatalf("bad, %s", err.Error())
		}

		masterData := map[string]interface{}{}

		for i, v := range columnTypes {

			if z, ok := (scanArgs[i]).(*sql.NullBool); ok {
				masterData[v.Name()] = z.Bool
				continue
			}

			if z, ok := (scanArgs[i]).(*sql.NullString); ok {
				masterData[v.Name()] = z.String
				continue
			}

			if z, ok := (scanArgs[i]).(*sql.NullInt64); ok {
				masterData[v.Name()] = z.Int64
				continue
			}

			if z, ok := (scanArgs[i]).(*sql.NullFloat64); ok {
				masterData[v.Name()] = z.Float64
				continue
			}

			if z, ok := (scanArgs[i]).(*sql.NullInt32); ok {
				masterData[v.Name()] = z.Int32
				continue
			}

			masterData[v.Name()] = scanArgs[i]
		}

		finalRows = append(finalRows, masterData)
	}

	z, err := json.Marshal(finalRows)
	if err != nil {
		return "", err
	}

	return string(z), err
}

func StartFrontend(reporterEndpoint string, dbType string, dbConnString string, buildRoot string) {
	db, err := sql.Open(dbType, dbConnString)
	if err != nil {
		log.Fatal(err.Error())
	}

	r := gin.Default()
	r.Use(cors.Default())
	r.Use(static.Serve("/", static.LocalFile(buildRoot, true)))
	g := r.Group("/api")
	g.GET("/orders", func(c *gin.Context) {
		o, err := getOrders(db)
		if err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}
		c.String(http.StatusOK, o)
	})
	g.GET("/archives", func(c *gin.Context) {
		var ret string
		archiveURL := c.Query("archiveUrl")

		// fetch single report if reportURL param
		if archiveURL != "" {
			var fetchURL string
			switch { // enable lookups from all reporting services from this instance
			case strings.HasPrefix(archiveURL, "http"):
				fetchURL = archiveURL
			case strings.HasPrefix(archiveURL, "/archive/"): // support older url lookups
				fetchURL = fmt.Sprintf("%s/api/archives/%s", reporterEndpoint, strings.Split(archiveURL, "/archive/")[1])
			default: // enable lookups for the well-known configured endpoint
				fetchURL = fmt.Sprintf("%s/api/archives/%s", reporterEndpoint, archiveURL)
			}

			resp, body, err := httpJSONGet(fetchURL)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			if resp.StatusCode != 200 {
				c.JSON(http.StatusInternalServerError, gin.H{"error": string(body)})
				return
			}

			// TODO Convert into report type
			ret = string(body)
		} else {
			// fetch archives if no reportURL param
			resp, bits, err := httpJSONGet(fmt.Sprintf("%s/api/archives", reporterEndpoint))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			if resp.StatusCode != 200 {
				c.JSON(http.StatusInternalServerError, gin.H{"error": string(bits)})
				return
			}

			var reports []ArchiveURL
			if err := json.Unmarshal(bits, &reports); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, reports)
			return
		}

		c.String(http.StatusOK, ret)
	})

	g.POST("/archives", func(c *gin.Context) {
		// Persist the POST body (json) via the reporter
		resp, bits, err := httpJSONPost(fmt.Sprintf("%s/api/archive", reporterEndpoint), c.Request.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if resp.StatusCode != 201 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": string(bits)})
			return
		}

		var retdata map[string]interface{}
		if err := json.Unmarshal(bits, &retdata); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"url": retdata["key"]})
	})

	r.NoRoute(func(ctx *gin.Context) {
		ctx.Redirect(http.StatusPermanentRedirect, "/")
	})

	if err := r.Run(); err != nil {
		log.Fatal(err)
	}
}

type ArchiveURL struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

func StartReporter(objectStorageEndpoint string, bucketName string, accessKey string, secretAccessKey string, staticRegion string) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatal(err.Error())
	}

	if staticRegion != "" {
		cfg.Region = staticRegion
	}

	if accessKey != "" && secretAccessKey != "" {
		cfg.Credentials = credentials.NewStaticCredentialsProvider(accessKey, secretAccessKey, "")
	}

	if objectStorageEndpoint != "" {
		staticResolver := aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
			return aws.Endpoint{
				PartitionID:       "aws",
				URL:               objectStorageEndpoint,
				SigningRegion:     staticRegion,
				HostnameImmutable: true,
			}, nil
		})
		cfg.EndpointResolver = staticResolver
	}

	s3Client := s3.NewFromConfig(cfg)
	r := gin.Default()
	r.Use(cors.Default())
	g := r.Group("/api")
	g.GET("/archives", func(c *gin.Context) {
		keys := []ArchiveURL{}

		out, err := s3Client.ListObjects(context.Background(), &s3.ListObjectsInput{Bucket: aws.String(bucketName)})
		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]string{"err": err.Error()})
			return
		}
		for _, o := range out.Contents {
			keys = append(keys, ArchiveURL{
				Name: *o.Key,
				URL:  *o.Key,
			})
		}

		// Get S3 listing, parse it, and return ids (epoch)
		c.JSON(http.StatusOK, keys)
	})

	g.GET("/archives/:id", func(c *gin.Context) {
		id, ok := c.Params.Get("id")
		if !ok {
			c.JSON(http.StatusInternalServerError, map[string]string{"err": "must supply archive id"})
			return
		}

		obj, err := s3Client.GetObject(context.Background(), &s3.GetObjectInput{Bucket: aws.String(bucketName), Key: aws.String(id)})
		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]string{"err": err.Error()})
			return
		}

		bits, err := ioutil.ReadAll(obj.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]string{"err": err.Error()})
			return
		}

		// Get S3 object for id; where id == epoch
		c.String(http.StatusOK, string(bits))
	})
	g.POST("/archive", func(c *gin.Context) {
		// Persist the POST body (json) to a new s3 object named :id
		objName := fmt.Sprintf("%d", time.Now().Unix())

		uploader := manager.NewUploader(s3Client)
		result, err := uploader.Upload(context.TODO(), &s3.PutObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(objName),
			Body:   c.Request.Body,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]string{"err": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, map[string]string{"key": *result.Key})
	})

	if err := r.Run(":9999"); err != nil {
		log.Fatal(err)
	}
}
