package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
)

// 메시지를 저장할 배열과 Mutex를 선언하여 동시성 문제를 방지
var messages []string
var mutex sync.RWMutex

// Define schema
func defineSchema() graphql.Schema {
	// Query: hello
	helloField := &graphql.Field{
		Type: graphql.String,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			return "Hello, graphql-go!", nil
		},
	}
	// Query: getMessages
	getMessagesField := &graphql.Field{
		Type: graphql.NewList(graphql.String),
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			mutex.RLock()
			defer mutex.RUnlock()
			return messages, nil
		},
	}

	// Mutation: updateMessage
	updateMessageField := &graphql.Field{
		Type: graphql.String,
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			input := p.Args["input"].(string)
			mutex.Lock()
			messages = append(messages, input)
			mutex.Unlock()
			return "Message added: " + input, nil
		},
	}

	// Mutation: deleteMessage
	deleteMessageField := &graphql.Field{
		Type: graphql.String,
		Args: graphql.FieldConfigArgument{
			"index": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.Int),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			index := p.Args["index"].(int)
			mutex.Lock()
			defer mutex.Unlock()

			// Index 범위 검사
			if index < 0 || index >= len(messages) {
				return nil, fmt.Errorf("invalid index: %d", index)
			}

			deletedMessage := messages[index]
			messages = append(messages[:index], messages[index+1:]...) // 메시지 삭제
			return "Deleted message: " + deletedMessage, nil
		},
	}

	// Define Query and Mutation types
	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"hello":       helloField,
			"getMessages": getMessagesField,
		},
	})

	mutationType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"updateMessage": updateMessageField,
			"deleteMessage": deleteMessageField,
		},
	})

	// Create schema
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query:    queryType,
		Mutation: mutationType,
	})
	if err != nil {
		log.Fatalf("failed to create schema, error: %v", err)
	}

	return schema
}

func main() {
	schema := defineSchema()

	h := handler.New(&handler.Config{
		Schema:     &schema,
		Pretty:     true,
		Playground: true,
	})

	http.Handle("/graphql", h)
	log.Println("Listening on :8085")
	log.Fatal(http.ListenAndServe(":8085", nil))
}
