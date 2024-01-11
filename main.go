package main

import (
	"database/sql"
	"fmt"
	"os"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	_ "github.com/lib/pq"
	"net/http"
	"time"
)

type Event struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Notes 	  string    `json:"notes"`
	UserID    int   	`json:"user_id"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
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
			"notes": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "The notes of the event.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if event, ok := p.Source.(*Event); ok {
						return event.Notes, nil
					}

					return nil, nil
				},
			},
			"user_id": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.Int),
				Description: "The user on the event.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if event, ok := p.Source.(*Event); ok {
						return event.UserID, nil
					}

					return nil, nil
				},
			},
			"start_date": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "The start date of the event.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if event, ok := p.Source.(*Event); ok {
						return event.StartDate, nil
					}

					return nil, nil
				},
			},
			"end_date": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "The end date of the event.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if event, ok := p.Source.(*Event); ok {
						return event.EndDate, nil
					}

					return nil, nil
				},
			},
		},
	})

	
	rootQuery := graphql.NewObject(graphql.ObjectConfig{
		Name: "RootQuery",
		Fields: graphql.Fields{
			"userevents": &graphql.Field{
				Type:        graphql.NewList(eventType),
				Description: "List of events for user.",
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
					"get_type": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					id, _ := params.Args["id"].(int)

					rows, err := db.Query("SELECT id, name, notes, user_id, start_date, end_date FROM events WHERE user_id = $1", id)
					checkErr(err)
					var events []*Event

					for rows.Next() {
						event := &Event{}

						err = rows.Scan(&event.ID, &event.Name, &event.Notes, &event.UserID, &event.StartDate, &event.EndDate)
						checkErr(err)
						events = append(events, event)
					}

					return events, nil
				},
			},
			"events": &graphql.Field{
				Type:        graphql.NewList(eventType),
				Description: "List of events.",
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {

					rows, err := db.Query("SELECT id, name, notes, user_id, start_date, end_date FROM events")
					checkErr(err)
					var events []*Event

					for rows.Next() {
						event := &Event{}

						err = rows.Scan(&event.ID, &event.Name, &event.Notes, &event.UserID, &event.StartDate, &event.EndDate)
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

	mux := http.NewServeMux()
    mux.Handle("/graphql", h)

    mux.HandleFunc("/heartbeat", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        fmt.Fprintf(w, "OK")
    })

	mux.HandleFunc("/heartbeat-db", func(w http.ResponseWriter, r *http.Request) {
        err := db.Ping()
        if err != nil {
            w.WriteHeader(http.StatusServiceUnavailable)
            fmt.Fprintf(w, "Database not accessible")
            return
        }
        w.WriteHeader(http.StatusOK)
        fmt.Fprintf(w, "Database is accessible")
    })

	http.ListenAndServe(":8080", mux)
}