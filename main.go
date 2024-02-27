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
	db           = &DB{}
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
		return nil, fmt.Errorf("configuring aws db client: %w", err)
	}
	dbClient := dynamodb.NewFromConfig(cfg)
	return &DB{
		client: dbClient,
		table:  "Countries",
	}, nil
}

func handler() error {
	logger.Info("Handler Start")

	countryList := []Country{}

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

	logger.Info("Scraping Start")
	err := c.Visit(TargetURL)
	if err != nil {
		return fmt.Errorf("visiting target url: %w", err)
	}
	logger.Info("Scraping Complete")

	itemList := []map[string]types.AttributeValue{}
	for _, country := range countryList {
		item, errM := attributevalue.MarshalMap(country)
		if errM != nil {
			return fmt.Errorf("marshalling country into item: %w", err)
		}
		itemList = append(itemList, item)
	}

	logger.Info("Database Writes Start")
	dbWriteStart := time.Now()
	for i, item := range itemList {
		logger.Info(fmt.Sprintf("Writing Item %v", i))
		_, err = db.client.PutItem(context.TODO(), &dynamodb.PutItemInput{TableName: &db.table, Item: item})
		if err != nil {
			return fmt.Errorf("writing item %d: %w", i, err)
		}
	}
	logger.Info("Database Writes Complete")
	logger.Info(fmt.Sprintf("Database writing took: %f seconds", time.Since(dbWriteStart).Seconds()))

	itemList = []map[string]types.AttributeValue{}
	for _, country := range countryList {
		item, errM := attributevalue.Marshal(country.Name)
		if errM != nil {
			return fmt.Errorf("marshalling country into item: %w", err)
		}
		itemList = append(itemList, map[string]types.AttributeValue{"Name": item})
	}

	logger.Info("Database Reads Start")
	dbReadsStart := time.Now()
	for i, item := range itemList {
		logger.Info(fmt.Sprintf("Reading Item %v", i))
		_, err := db.client.GetItem(context.TODO(), &dynamodb.GetItemInput{TableName: &db.table, Key: item})
		if err != nil {
			return fmt.Errorf("reading item %d: %w", i, err)
		}
	}
	logger.Info("Database Reads Complete")
	logger.Info(fmt.Sprintf("Database reading took: %f seconds", time.Since(dbReadsStart).Seconds()))

	return nil
}

func main() {
	logger.Info("App Start")
	logger.Info("Setting Up DB Client")
	var err error
	db, err = NewDB()
	if err != nil {
		logger.Error(fmt.Errorf("creating db client: %w", err).Error())
		os.Exit(1)
	}

	if os.Getenv("AWS_LAMBDA_RUNTIME_API") != "" {
		lambda.Start(handler)
	} else {
		err := handler()
		if err != nil {
			logger.Error(err.Error())
			os.Exit(1)
		}
	}
	logger.Info("Ciao")
	os.Exit(0)
}
