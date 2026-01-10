package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Capsule struct {
	id        string
	title     string
	msg       string
	media     Media
	emailList []string
	status    string
}

type Media struct {
	Files []File `json:"file"`
}
type File struct {
	Name   string `json:"name"`
	Bucket string `json:"bucket"`
	Path   string `json:"path"`
}
type emailPayload struct {
	title   string
	msg     string
	adjunct []*memories
}
type memories struct {
	name string
	data []byte
}

func FetchDueCapsules(ctx context.Context) ([]Capsule, error) {
	now := time.Now().UTC()
	fmt.Println("WORKER_NOW_UTC =", now.Format(time.RFC3339))

	db, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, err
	}
	defer db.Close()
	fmt.Print("Connection to DB was done\n")
	fmt.Println("DB_URL =", os.Getenv("DATABASE_URL"))
	
	// First, let's see ALL capsules
	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM public.capsules_capsule").Scan(&count)
	if err != nil {
		fmt.Println("Error counting capsules:", err)
	} else {
		fmt.Println("Total capsules in DB:", count)
	}
	
	// Check how many should be due
	var dueCount int
	err = db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM public.capsules_capsule 
		WHERE release_time <= $1
	`, now).Scan(&dueCount)
	if err != nil {
		fmt.Println("Error counting due capsules:", err)
	} else {
		fmt.Println("Capsules with release_time <= now:", dueCount)
	}
	
	// Show the actual release times
	timeRows, err := db.QueryContext(ctx, `
		SELECT id, release_time, status 
		FROM public.capsules_capsule 
		ORDER BY release_time
	`)
	if err == nil {
		fmt.Println("All capsules:")
		for timeRows.Next() {
			var id string
			var releaseTime time.Time
			var status string
			timeRows.Scan(&id, &releaseTime, &status)
			fmt.Printf("  ID: %s, Release: %s, Status: %s\n", id, releaseTime.Format(time.RFC3339), status)
		}
		timeRows.Close()
	}
	
	// Running our actual query
	rows, err := db.QueryContext(ctx, `
		SELECT 
			id,
			title,
			message,
			media,
			email_list,
			status
		FROM public.capsules_capsule
		WHERE release_time <= $1
	`, now)
	
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var capsules []Capsule
	for rows.Next() {
		var c Capsule
		var JSONBlob []byte
		var emailJSON []byte

		if err := rows.Scan(
			&c.id,
			&c.title,
			&c.msg,
			&JSONBlob,
			&emailJSON,
			&c.status,
		); err != nil {
			return nil, err
		}

		if len(JSONBlob) > 0 {
			if err := json.Unmarshal(JSONBlob, &c.media); err != nil {
				return nil, err
			}
		}
		if len(emailJSON) > 0 {
			if err := json.Unmarshal(emailJSON, &c.emailList); err != nil {
				return nil, err
			}
		}
		capsules = append(capsules, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	fmt.Printf("Fetched %d due capsules\n", len(capsules))
	return capsules, nil
}

func MarkDue(c Capsule) {
	c.status = "due"
}

func StreamMedia_fromBucket(fileName string, url string) (*memories, error) {
	var BASE string
	BASE = "https://yuootleblefkauksfscf.supabase.co/storage/v1/object/authenticated/Annex"
	url = BASE + "/" + url

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	apiKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("apikey", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch file: %s", resp.Status)
	}

	// Stream → memory
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Create object directly
	mem := &memories{
		name: fileName,
		data: data,
	}
	fmt.Print("Fetched item from bucket successfully...\n")
	return mem, nil
}

func ProcessCapsule(c Capsule) (bool, error) {
	var payload emailPayload
	payload.title = c.title
	payload.msg = c.msg
	var attachments []*memories
	fmt.Print("Processing the capsule\n")

	// now we need to fetch the file specified in the capsule via supabase
	for _, file := range c.media.Files {
		fmt.Print("Fetching files...\n")
		path := file.Path
		name := file.Name
		mem, err := StreamMedia_fromBucket(name, path)
		fmt.Print("Fetched = " + path + " " + name + " from the bucket\n")
		if err != nil {
			return false, err
		}
		attachments = append(attachments, mem)
	}
	payload.adjunct = attachments
	var email_list []string = c.emailList

	sentStatus, err := SendEmail(payload, email_list)
	if err != nil {
		return false, err
	}
	return sentStatus, nil
}

func MarkDone(c Capsule) {
	c.status = "done"
}