# opa-client-go

This package provides a very minimal client for the [OPA Data REST API](https://www.openpolicyagent.org/docs/latest/rest-api/#data-api)
using Go.

## Usage

Install the dependency:

```sh
go get -u github.com/alex-kuck/opa-client-go
```

```go
package example

import (
	"context"
	"fmt"
	opa "github.com/alex-kuck/opa-client-go/pkg"
	"net/http"
)

type legalAgeCheck struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type legalAgeResult struct {
	Access bool `json:"access"`
}

func accessOPA(ctx context.Context) {
	client := opa.NewClient("http://localhost:8181", http.DefaultClient)

	ageToCheck := legalAgeCheck{Name: "Chuch Norris", Age: 21}
	result, err := opa.Query[legalAgeCheck, legalAgeResult](ctx, client, "/example/age_check", ageToCheck)
	if err != nil {
		fmt.Printf("Oh no that didn't work 😱: %s", err.Error())
	}

	// Access the response type-safe 😎
	fmt.Printf("%s has access? %b", ageToCheck.Name, result.Access)
}
```
