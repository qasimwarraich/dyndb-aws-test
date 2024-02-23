package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gocolly/colly"
)

var (
	TargetDomain = "www.scrapethissite.com"
	TargetURL    = "https://www.scrapethissite.com/pages/simple/"
	logger       = slog.New(slog.NewJSONHandler(os.Stdout, nil))
)

type Country struct {
	Name       string
	Population string
}

type DB struct {
	client *dynamodb.Client
	table  string
}

func NewDB() (*DB, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}
	dbClient := dynamodb.NewFromConfig(cfg)
	return &DB{
		client: dbClient,
		table:  "Countries",
	}, nil
}

func app(logger *slog.Logger) error {
	logger.Info("App Start")

	countryList := []Country{}

	logger.Info("Scraping Start")
	c := colly.NewCollector(
		colly.AllowedDomains(
			TargetDomain,
		),
	)

	c.OnHTML(".country", func(e *colly.HTMLElement) {
		countryName := e.ChildText(".country-name")
		countryPopulation := e.ChildText(".country-population")

		country := Country{
			Name:       countryName,
			Population: countryPopulation,
		}
		countryList = append(countryList, country)
	})

	err := c.Visit(TargetURL)
	if err != nil {
		logger.Error(err.Error())
		return err
	}
	logger.Info("Scraping Complete")

	logger.Info("Setting Up DB Client")
	db, err := NewDB()
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	itemList := []map[string]types.AttributeValue{}

	for _, country := range countryList {
		item, errM := attributevalue.MarshalMap(country)
		if errM != nil {
			logger.Error(errM.Error())
			return err
		}
		itemList = append(itemList, item)
	}

	logger.Info("Database Writes Start")
	dbWriteStart := time.Now()
	for i, item := range itemList {
		logger.Info(fmt.Sprintf("Writing Item %v", i))
		_, err = db.client.PutItem(context.TODO(), &dynamodb.PutItemInput{TableName: &db.table, Item: item})
		if err != nil {
			logger.Error(err.Error())
			return err
		}
	}
	logger.Info("Database Writes Complete")
	logger.Info(fmt.Sprintf("Database writing took: %f seconds", time.Since(dbWriteStart).Seconds()))
	return nil
}

func main() {
	if os.Getenv("AWS_EXECUTION_ENV") != "" {
		lambda.Start(app(logger))
	} else {
		err := app(logger)
		if err != nil {
			logger.Error(err.Error())
		}
	}

	logger.Info("Ciao")
}