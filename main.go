package main

import (
	"database/sql"
	"fmt"
	"os"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	_ "github.com/lib/pq"
	"net/http"
)

type Event struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	dbHost := os.Getenv("POSTGRES_HOST")
	dbPort := os.Getenv("POSTGRES_PORT")
	dbUser := os.Getenv("POSTGRES_USER")
	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")

	dbinfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)
	db, err := sql.Open("postgres", dbinfo)
	checkErr(err)

	defer db.Close()

	eventType := graphql.NewObject(graphql.ObjectConfig{
		Name:        "Event",
		Description: "An event",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.Int),
				Description: "The identifier of the event.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if event, ok := p.Source.(*Event); ok {
						return event.ID, nil
					}

					return nil, nil
				},
			},
			"name": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "The name of the event.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if event, ok := p.Source.(*Event); ok {
						return event.Name, nil
					}

					return nil, nil
				},
			},
		},
	})

	
	rootQuery := graphql.NewObject(graphql.ObjectConfig{
		Name: "RootQuery",
		Fields: graphql.Fields{
			"events": &graphql.Field{
				Type:        graphql.NewList(eventType),
				Description: "List of events.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					rows, err := db.Query("SELECT id, name FROM events")
					checkErr(err)
					var events []*Event

					for rows.Next() {
						event := &Event{}

						err = rows.Scan(&event.ID, &event.Name)
						checkErr(err)
						events = append(events, event)
					}

					return events, nil
				},
			},
		},
	})

	schema, _ := graphql.NewSchema(graphql.SchemaConfig{
		Query:    rootQuery,
	})

	h := handler.New(&handler.Config{
		Schema:   &schema,
		Pretty:   true,
		GraphiQL: true,
	})

	// serve HTTP
	http.Handle("/graphql", h)
	http.ListenAndServe(":8080", nil)
}