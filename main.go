package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"path/filepath"
	"time"

	"github.com/alexflint/go-arg"
	"github.com/pieterclaerhout/export-komoot/komoot"
	"github.com/pieterclaerhout/go-log"
	"github.com/pieterclaerhout/go-waitgroup"
)

type args struct {
	Email        string `arg:"env:KOMOOT_EMAIL,required" help:"Your Komoot email address"`
	Password     string `arg:"env:KOMOOT_PASSWD,required" help:"Your Komoot password"`
	UserID       int64  `arg:"env:KOMOOT_USER_ID,required" help:"Your Komoot user ID"`
	Filter       string `help:"Filter tours with name matching this pattern"`
	To           string `arg:"required" help:"The path to export to"`
	FullDownload bool   `help:"If specified, all data is redownloaded" default:"false"`
	Concurrency  int    `help:"The number of simultaneous downloads" default:"16"`
	TourType     string `help:"The type of tours to download" default:""`
}

func main() {
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			log.Fatal("Error loading .env file:", err.Error())
		}
	}

	var args args
	p := arg.MustParse(&args)

	if args.To == "" {
		p.Fail("you must provide a value for --to")
	}

	log.PrintTimestamp = true
	log.PrintColors = true

	start := time.Now()
	defer func() { log.Info("Elapsed:", time.Since(start)) }()

	client := komoot.NewClient(args.Email, args.Password, args.UserID)

	fullDstPath, _ := filepath.Abs(args.To)
	log.Info("Exporting:", args.Email)
	log.Info("       to:", fullDstPath)

	err := os.MkdirAll(args.To, 0777)
	log.CheckError(err)

	log.Info("Komoot User ID:", args.UserID)

	tours, resp, err := client.Tours(args.Filter, args.TourType)
	log.CheckError(err)

	if len(tours) == 0 {
		log.Info("No tours need to be downloaded")
		return
	}

	log.Info("Found", len(tours), "planned tours")

	var allTours []komoot.Tour

	if !args.FullDownload {

		log.Info("Incremental download, checking what has changed")

		var changedTours []komoot.Tour

		for _, tour := range tours {

			allTours = append(allTours, tour)

			dstPath := filepath.Join(args.To, tour.Filename())
			if !fileExists(dstPath) {
				changedTours = append(changedTours, tour)
			}

		}

		tours = changedTours

		if len(tours) == 0 {
			log.Info("No tours need to be downloaded")
		} else {
			log.Info("Found", len(tours), "which need to be downloaded")
		}

	} else {
		allTours = tours
	}

	if len(tours) > 0 {
		log.Info("Downloading with a concurrency of", args.Concurrency)
		wg := waitgroup.NewWaitGroup(args.Concurrency)

		var downloadCount int

		for _, tour := range tours {

			tourToDownload := tour
			label := fmt.Sprintf("%10d | %-15s | %-15s | %s", tour.ID, tour.FormattedSport(), tour.Type, tour.Name)

			wg.Add(func() {

				if err := func() error {

					out, err := client.Download(tour)
					if err != nil {
						return err
					}

					dstPath := filepath.Join(args.To, tourToDownload.Filename())
					if err = saveTourFile(out, dstPath, tourToDownload); err != nil {
						return err
					}

					log.Info("Downloaded:", label)

					return nil

				}(); err != nil {
					log.Error("Downloaded:", label, "|", err)
				}
				downloadCount++

			})

		}

		wg.Wait()

		log.Info("Downloaded", downloadCount, "tours")
	}

	allTourNames := map[string]bool{}
	for _, tour := range allTours {
		allTourNames[tour.Filename()] = true
	}

	items, err := filepath.Glob(filepath.Join(args.To, "*.gpx"))
	log.CheckError(err)
	for _, item := range items {
		if _, exists := allTourNames[filepath.Base(item)]; exists {
			continue
		}
		log.Info("Deleting:", filepath.Base(item))
		if err := os.Remove(item); err != nil {
			log.Warn("Failed to delete:", item, "|", err)
		}
	}

	log.Info("Saving tour list")
	dstPath := filepath.Join(args.To, "tours.json")
	err = saveFormattedJSON(resp, dstPath)
	log.CheckError(err)

	var out bytes.Buffer
	err = json.NewEncoder(&out).Encode(allTours)
	log.CheckError(err)

	log.Info("Saving parsed tour list")
	dstPath = filepath.Join(args.To, "tours_parsed.json")
	err = saveFormattedJSON(out.Bytes(), dstPath)
	log.CheckError(err)

}
