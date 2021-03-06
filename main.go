package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
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

	fullDstPath, _ := filepath.Abs(*toPtr)
	log.Info("Exporting:", *emailPtr, "to:", fullDstPath)

	log.Info("Komoot User ID:", userID)

	tours, resp, err := client.Tours(userID)
	log.Info("Found", len(tours), "planned tours")

	if *noIncrementalPtr == false {

		log.Info("Incremental download, checking what has changed")

		changedTours := []komoot.Tour{}

		for _, tour := range tours {
			dstPath := filepath.Join(*toPtr, tour.Filename())
			if !fileExists(dstPath) {
				changedTours = append(changedTours, tour)
			}
		}

		tours = changedTours

		if len(tours) == 0 {
			log.Info("No tours need to be downloaded")
			return
		}

		log.Info("Found", len(tours), "which need to be downloaded")

	}

	log.Info("Downloading with a concurrency of", *concurrencyPtr)
	wg := waitgroup.NewWaitGroup(*concurrencyPtr)

	var downloadCount int
	var recreateCount int

	for _, tour := range tours {

		tourToDownload := tour
		label := fmt.Sprintf("%10d | %-7s | %-15s | %s", tour.ID, tour.Status, tour.FormattedSport(), tour.Name)

		if !tour.IsCycling() {
			continue
		}

		wg.Add(func() {

			gpx, full, err := client.Download(tourToDownload)
			if err != nil {
				log.Error("Downloaded:", label, "|", err)
				return
			}

			deleteWithPattern(*toPtr, fmt.Sprintf("%d_*.gpx", tourToDownload.ID))

			dstPath := filepath.Join(*toPtr, tourToDownload.Filename())

			err = ioutil.WriteFile(dstPath, gpx, 0755)
			if err != nil {
				log.Error("Downloaded:", label, "|", err)
				return
			}

			os.Chtimes(dstPath, tour.Date, tour.ChangedAt)

			if !full {
				log.Warn(" Recreated:", label)
				recreateCount++
			} else {
				log.Info("Downloaded:", label)
				downloadCount++
			}

		})

	}

	wg.Wait()

	log.Info("Downloaded", downloadCount, "cycling tours")
	log.Info("Recreated", recreateCount, "cycling tours")

	log.Info("Saving tour list")
	dstPath := filepath.Join(*toPtr, "tours.json")
	err = saveFormattedJSON(resp, dstPath)
	log.CheckError(err)

	var out bytes.Buffer
	err = json.NewEncoder(&out).Encode(tours)
	log.CheckError(err)

	log.Info("Saving parsed tour list")
	dstPath = filepath.Join(*toPtr, "tours_parsed.json")
	err = saveFormattedJSON(out.Bytes(), dstPath)
	log.CheckError(err)

}
