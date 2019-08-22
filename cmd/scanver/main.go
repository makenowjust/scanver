package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/MakeNowJust/scanver"
)

var (
	owner   = flag.String("o", "", "owner (a.k.a. user/org) name to scan repositories")
	pkgName string
)

func init() {
	flag.Parse()
	pkgName = flag.Arg(0)
}

func main() {
	if len(*owner) == 0 || len(pkgName) == 0 {
		fmt.Fprintf(os.Stderr, "owner or pkgName is missing!")
		os.Exit(1)
	}

	ctx := context.Background()

	token, err := scanver.LookupAccessToken()
	if err != nil {
		log.Fatal(err)
	}
	client := scanver.NewClient(ctx, token)

	repos, err := client.SearchRepositories(ctx, *owner, pkgName)
	if err != nil {
		log.Fatal(err)
	}

	for _, repo := range repos {
		vs, err := client.LookupPackageVersions(ctx, repo, pkgName)
		if err != nil {
			log.Print(err)
		} else if len(vs) > 0 {
			fmt.Printf("github.com/%s/%s %s\n", repo.Owner, repo.Name, strings.Join(vs, ","))
		}

		time.Sleep(1 * time.Second)
	}
}
