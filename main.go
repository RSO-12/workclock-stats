package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/graphql-go/graphql"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type EventEntity struct {
	ID              int       `json:"id"`
	Name            string    `json:"name"`
	StartDate       time.Time `json:"start_date"`
	EndDate         time.Time `json:"end_date"`
	Notes           string    `json:"notes"`
	PreviousEventID int       `json:"previous_event_id"`
	UserID          int       `json:"user_id"`
	EventTypeID     int       `json:"event_type_id"`
}

var eventEntityType = graphql.NewObject(graphql.ObjectConfig{
	Name: "EventEntity",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.Int,
		},
		"name": &graphql.Field{
			Type: graphql.String,
		},
		"start_date": &graphql.Field{
			Type: graphql.DateTime,
		},
		"end_date": &graphql.Field{
			Type: graphql.DateTime,
		},
		"notes": &graphql.Field{
			Type: graphql.String,
		},
		"previous_event_id": &graphql.Field{
			Type: graphql.Int,
		},
		"user_id": &graphql.Field{
			Type: graphql.Int,
		},
		"event_type_id": &graphql.Field{
			Type: graphql.Int,
		},
	},
})

func main() {
	pool, err := pgxpool.Connect(context.Background(), "postgresql://username:password@localhost:5432/your_database")
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()

	rootQuery := graphql.ObjectConfig{Name: "Query", Fields: graphql.Fields{
		"events": &graphql.Field{
			Type:        graphql.NewList(eventEntityType),
			Description: "Get all events",
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				events := []EventEntity{}

				rows, err := pool.Query(context.Background(), "SELECT * FROM events")
				if err != nil {
					return nil, err
				}
				defer rows.Close()

				for rows.Next() {
					var event EventEntity
					err := rows.Scan(&event.ID, &event.Name, &event.StartDate, &event.EndDate, &event.Notes, &event.PreviousEventID, &event.UserID, &event.EventTypeID)
					if err != nil {
						return nil, err
					}
					events = append(events, event)
				}

				return events, nil
			},
		},
	}}

	schemaConfig := graphql.SchemaConfig{Query: graphql.NewObject(graphql.ObjectConfig{Name: "RootQuery", Fields: rootQuery})}
	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		log.Fatalf("Error creating schema: %v", err)
	}

	// Execute GraphQL query
	result := graphql.Do(graphql.Params{
		Schema:        schema,
		RequestString: "{ events { id name start_date end_date notes previous_event_id user_id event_type_id } }",
	})
	if len(result.Errors) > 0 {
		log.Fatalf("Error executing query: %v", result.Errors)
	}
	fmt.Printf("%v\n", result)
}
