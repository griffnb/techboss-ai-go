package main

import (
	"flag"
	"fmt"
	"maps"
	"strings"

	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/models/base"

	"github.com/CrowdShield/go-core/lib/log"
	"github.com/CrowdShield/go-core/lib/model"
	"github.com/griffnb/techboss-ai-go/internal/models"
)

func main() {
	var table string
	flag.StringVar(&table, "table", "", "Name of the table")
	flag.Parse()

	environment.CreateEnvironment()

	_ = models.LoadModels()

	log.Debug("Models Loaded")

	defaultPtrs := environment.GetDBClient(environment.CLIENT_DEFAULT).GetTablePtrs()

	allPtrs := make(map[string]any)

	maps.Copy(allPtrs, defaultPtrs)

	if table != "" {
		tables := strings.Split(table, ",")
		for _, t := range tables {
			val, exists := allPtrs[t]
			if !exists || val == nil {
				panic("Table " + t + " does not exist in the models")
			}

			fmt.Printf("---------------------- %s ----------------------\n", t)
			model.GenerateSingleTypescript(val, &base.Structure{})
			model.GeneratePublicTypeScriptModel(val, t)
			fmt.Printf("\n------------------------------------------------\n")

		}
		return
	}
	model.GenerateAllTypescript(allPtrs, &base.Structure{})
}
