package main

import (
	"log"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"strconv"
	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
)

type Person struct {
	Id int `json:"id"`
	Name string `json:"name"`
	CreatedAt string `json:"createdAt"`
	Address string `json:"address"`
	Url string `json:"url"`
	Photos []string `json:"photos"`
}

func main() {
	router := mux.NewRouter()

	handler := negroni.New()

	handler.Use(negroni.HandlerFunc(func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}))

	handler.UseHandler(router)

	router.HandleFunc("/mg", func (w http.ResponseWriter, r *http.Request) {
		form := url.Values{}
		form.Add("quant", "10")

		req, err := http.NewRequest("POST", "https://desaparecidos.policiacivil.mg.gov.br/arquivo/getBannersAjax", strings.NewReader(form.Encode()))
		if err != nil {
			log.Fatal(err)
		}

		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Fatal(err)
		}

		var response map[string] struct {
			Id int `json:"id"`
		}

		json.NewDecoder(res.Body).Decode(&response)

		var people []Person

		for index := range response {
			var person Person

			person.Url = "https://desaparecidos.policiacivil.mg.gov.br/desaparecido/exibir/" + strconv.Itoa(response[index].Id)

			req, err := http.NewRequest("GET", person.Url, nil)
			if err != nil {
				log.Fatal(err)
			}

			res, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Fatal(err)
			}
			document, err := goquery.NewDocumentFromReader(res.Body)
			if err != nil {
				log.Fatal(err)
			}

			document.Find("dt").Each(func(i int, e *goquery.Selection) {
				person.Id = response[index].Id
				if e.Text() == "Nome do desaparecido:" {
					person.Name = e.Next().Text()
				}
				if e.Text() == "Data do desaparecimento" {
					person.CreatedAt = e.Next().Text()
				}
				if e.Text() == "Município/Cidade do desaparecimento" {
					person.Address = e.Next().Text()
				}
			})

			document.Find("h2").Each(func(i int, e *goquery.Selection) {
				if e.Text() == "FOTOS" {
					src, exists := e.Next().Children().Children().Children().Attr("src")
					if err != nil {
						log.Fatal(err)
					}
					if exists {
						person.Photos = append([]string{"https://desaparecidos.policiacivil.mg.gov.br/" + src}, person.Photos...)
					}
				} else {
					src, exists := e.Next().Children().Attr("src")
					if err != nil {
						log.Fatal(err)
					}
					if exists {
						person.Photos = append(person.Photos, "https://desaparecidos.policiacivil.mg.gov.br/" + src)
					}
				}
			})

			people = append(people, person)
		}

		json.NewEncoder(w).Encode(people)
	}).Methods("GET")

	http.ListenAndServe(":8080", handler)
}