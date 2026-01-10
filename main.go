package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"github.com/joho/godotenv"
)

func main(){
	godotenv.Load()
	http.HandleFunc("/run",auth(runJob))
	port:=os.Getenv("PORT")
	if port==""{
		port="8080"
	}
	log.Println("Go Worker Running on : "+port+ "\n")
	log.Fatal(http.ListenAndServe(":"+port,nil))
}

func auth(next http.HandlerFunc) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request){
		// here we validate to check if the request is coming from a authorized user with a valid bearer key or not
		secret:=r.Header.Get("Authorization")
		if secret != "Bearer "+os.Getenv("WORKER_SECRET"){
			http.Error(w,"unauthorized",http.StatusUnauthorized)
			return
		}
		fmt.Print("Calling the handler\n")
		// if the req. is valid we call the function that has to process this job and will return that response 
		next(w,r)
	}
}

func runJob(w http.ResponseWriter, r *http.Request) {
	capsules, err := FetchDueCapsules(r.Context())
	if len(capsules)==0{
		fmt.Print("No due capsules\n")
	}
	if err!=nil{
		http.Error(w,err.Error(),500)
		return	
	}
	for _,capsule:= range capsules{
		// we process the capsule ie fetch the files from bucket and bind the media in to the email attachements and call the emailer function for all the emails in the list 
		status,err:=ProcessCapsule(capsule)
		if err!=nil{
			MarkDue(capsule)
			continue
		}
		if status{
			MarkDone(capsule)
		}
	}
	w.Write([]byte("ok"))
}