package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/pieterclaerhout/export-komoot/komoot"
	"github.com/pieterclaerhout/go-log"
	"github.com/pieterclaerhout/go-waitgroup"
)

func main() {

	log.PrintTimestamp = true
	log.PrintColors = true

	emailPtr := flag.String("email", "", "Your Komoot email address")
	passwordPtr := flag.String("password", "", "Your Komoot password")
	toPtr := flag.String("to", "", "The path to export to")
	noIncrementalPtr := flag.Bool("no-incremental", false, "If specified, all data is redownloaded")
	concurrencyPtr := flag.Int("concurrency", 4, "The number of simultaneous downloads")
	flag.Parse()

	client := komoot.NewClient(*emailPtr, *passwordPtr)

	userID, err := client.Login()
	log.CheckError(err)

	log.Info("Exporting:", *emailPtr, "to:", *toPtr)

	log.Info("> Komoot User ID:", userID)

	tours, resp, err := client.Tours(userID)
	log.Info("> Found", len(tours), "planned tours")

	if *noIncrementalPtr == false {

		log.Info("> Incremental download, checking what has changed")

		changedTours := []komoot.Tour{}

		for _, tour := range tours {
			dstPath := filepath.Join(*toPtr, tour.Filename())
			if !fileExists(dstPath) {
				changedTours = append(changedTours, tour)
			}
		}

		tours = changedTours

		if len(tours) == 0 {
			log.Info("> No tours need to be downloaded")
			return
		}

		log.Info("> Found", len(tours), "which need to be downloaded")

	}

	log.Info("> Downloading with a concurrency of", *concurrencyPtr)
	wg := waitgroup.NewWaitGroup(*concurrencyPtr)

	var downloadCount int

	for _, tour := range tours {

		tourToDownload := tour
		label := fmt.Sprintf("%10d | %10s | %-15s | %s", tour.ID, tour.Status, tour.Sport, tour.Name)

		if !tour.IsCycling() {
			continue
		}

		wg.Add(func() {

			log.Info("> Downloading:", label, "|", tourToDownload.ChangedAt)

			gpx, err := client.Download(int(tourToDownload.ID))
			if err != nil {
				log.Error("> Downloading:", label, "|", err)
				return
			}

			dstPath := filepath.Join(*toPtr, tourToDownload.Filename())

			err = ioutil.WriteFile(dstPath, gpx, 0755)
			if err != nil {
				log.Error("> Downloading:", label, "|", err)
				return
			}

			downloadCount++

		})

	}

	wg.Wait()

	log.Info("> Downloaded", downloadCount, "cycling tours")

	log.Info("> Saving tour list")

	var out bytes.Buffer
	err = json.Indent(&out, resp, "", "  ")
	log.CheckError(err)

	dstPath := filepath.Join(*toPtr, "tours.json")
	err = ioutil.WriteFile(dstPath, out.Bytes(), 0755)
	log.CheckError(err)

}
