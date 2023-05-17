package main

import (
	"encoding/csv"
	"flag"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	if (err != nil) && (!os.IsNotExist(err)) {
		log.Fatal("Error loading .env file")
	}

	var username string
	var password string
	var clubId string
	var debug bool

	flag.StringVar(&username, "username", os.Getenv("DUPR_USERNAME"), "User name to log into DUPR.")
	flag.StringVar(&password, "password", os.Getenv("DUPR_PASSWORD"), "Password to log into DUPR.")
	flag.StringVar(&clubId, "club_id", os.Getenv("DUPR_CLUB_ID"), "We will get players belowing to this DUPR club id.")
	flag.BoolVar(&debug, "debug", false, "Debug mode, print out data json")
	flag.Parse()
	if username == "" {
		log.Fatalln("username is required")
	}
	if password == "" {
		log.Fatalln("password is required")
	}
	if clubId == "" {
		log.Fatalln("clubID is required")
	}

	log.Println("username is " + username)
	log.Println("password is " + "******")
	log.Println("club_id is " + clubId)
	login(username, password, debug)

	members, _ := GetMembersByClub(clubId)

	outfile, err := os.Create("./players.csv")
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("%d members found.\n", len(members))
	csvout := csv.NewWriter(outfile)
	for _, m := range members {
		err := csvout.Write([]string{m.FullName, m.Ratings.Display()})
		if err != nil {
			log.Fatalf("Error writing to CSV file %v\n", err.Error())
		}
		// log.Printf("\"%s\", %s\n", m.FullName, m.Ratings.Display())
	}
	csvout.Flush()
	outfile.Close()
}
