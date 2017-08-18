package main

import (
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/shopspring/decimal"
	"golang.org/x/net/html/charset"
)

type Cb struct {
	ExchangeRateList []ExchangeRate `xml:"Valute"`
}

type ExchangeRate struct {
	CharCode string `xml:"CharCode"`
	Nominal  string `xml:"Nominal"`
	Name     string `xml:"Name"`
	Value    string `xml:"Value"`
}

func RublesInCurrency(currencyCode string, cb Cb) (currencyRate decimal.Decimal) {

	currencyRate = decimal.Zero

	if currencyCode == "RUR" {
		return decimal.NewFromFloat(1)
	}

	for _, element := range cb.ExchangeRateList {
		if element.CharCode == currencyCode {
			currencyValueString := strings.Replace(element.Value, ",", ".", -1)
			currencyValue, err := decimal.NewFromString(currencyValueString)
			if err != nil {
				log.Fatal(err)
			}
			currencyNominal, err := decimal.NewFromString(element.Nominal)
			if err != nil {
				log.Fatal(err)
			}
			currencyRate = currencyValue.Div(currencyNominal)
			break
		}
	}

	return
}

func getCb(url string) (cb Cb) {

	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	cb = Cb{}

	buffer := bytes.NewBuffer(body)
	decoder := xml.NewDecoder(buffer)
	decoder.CharsetReader = charset.NewReaderLabel
	err = decoder.Decode(&cb)
	if err != nil {
		log.Fatal(err)
	}

	return

}

func main() {

	// default values
	currencyCodePtr := flag.String("currency", "USD", "Currency code")
	valuePtr := flag.String("value", "1", "Value")

	flag.Parse()

	currencyCode := *currencyCodePtr
	value, err := decimal.NewFromString(*valuePtr)
	if err != nil {
		log.Fatal(err)
	}

	cb := getCb("http://www.cbr.ru/scripts/XML_daily.asp")

	// input currency to RUR
	rubles := value.Mul(RublesInCurrency(currencyCode, cb))

	// RUR to target currency
	fmt.Println(value, currencyCode, " = ", rubles.Div(RublesInCurrency("RUR", cb)), "RUR")
	fmt.Println(value, currencyCode, " = ", rubles.Div(RublesInCurrency("EUR", cb)), "EUR")
	fmt.Println(value, currencyCode, " = ", rubles.Div(RublesInCurrency("AMD", cb)), "AMD")

}
