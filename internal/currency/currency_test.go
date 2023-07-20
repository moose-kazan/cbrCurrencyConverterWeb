package currency

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCurrencyList_parse(t *testing.T) {
	t.Run("valid data", func(t *testing.T) {
		var apidata string = `<?xml version="1.0" encoding="windows-1251"?>
		<ValCurs Date="15.07.2023" name="Foreign Currency Market">
		<Valute ID="R01239"><NumCode>978</NumCode><CharCode>EUR</CharCode><Nominal>1</Nominal><Name>Euro</Name><Value>101,1992</Value></Valute>
		<Valute ID="R01235"><NumCode>840</NumCode><CharCode>USD</CharCode><Nominal>1</Nominal><Name>US Dollar</Name><Value>90,1190</Value></Valute>
		</ValCurs>`
		var cl CurrencyList
		err := cl.parse([]byte(apidata))
		if err != nil {
			t.Errorf("Can't parse data: %v", err)
			return
		}
		if len(cl.Rates) != 3 {
			t.Errorf("Expected 3 currencies, but found %v", len(cl.Rates))
		}
	})

	t.Run("invalid data", func(t *testing.T) {
		var apidata string = "some wrong data"
		var cl CurrencyList
		err := cl.parse([]byte(apidata))
		if err == nil {
			t.Errorf("Bad data, but no error!")
			return
		}
	})
}

func TestCurrencyList_Convert(t *testing.T) {
	var cl CurrencyList
	cl.Rates = append(cl.Rates, CurrencyInfo{CharCode: "RUB", Value: "1", Nominal: 1})
	cl.Rates = append(cl.Rates, CurrencyInfo{CharCode: "USD", Value: "90", Nominal: 1})
	cl.Rates = append(cl.Rates, CurrencyInfo{CharCode: "EUR", Value: "100", Nominal: 1})
	cl.Rates = append(cl.Rates, CurrencyInfo{CharCode: "KZT", Value: "20", Nominal: 100})
	t.Run("Valid data", func(t *testing.T) {
		type currencyTest struct {
			from     string
			to       string
			expected float64
		}
		var testData []currencyTest = []currencyTest{
			{from: "RUB", to: "RUB", expected: 1},
			{from: "RUB", to: "USD", expected: 90},
			{from: "EUR", to: "RUB", expected: 0.01},
			{from: "KZT", to: "USD", expected: 450},
		}
		for _, testItem := range testData {
			result, err := cl.Convert(testItem.from, testItem.to)
			if err != nil {
				t.Errorf("Can't convert from %v to %v: %v", testItem.from, testItem.to, err)
			}
			if result != testItem.expected {
				t.Errorf("Convert from %v to %v. Expected: %v, got: %v", testItem.from, testItem.to, testItem.expected, result)
			}
		}
	})
	t.Run("Invalid data", func(t *testing.T) {
		_, err := cl.Convert("GBP", "USD")
		if err == nil {
			t.Errorf("Bad data, but no error reported!")
		}
	})
}

func TestCurrencyInfo_GetValue(t *testing.T) {
	type fields struct {
		XMLName  xml.Name
		ID       xml.Attr
		NumCode  string
		CharCode string
		Nominal  int
		Name     string
		Value    string
	}
	tests := []struct {
		name   string
		fields fields
		want   float64
	}{
		{name: "val 1", want: 10, fields: fields{Value: "10"}},
		{name: "val 2", want: 15.5, fields: fields{Value: "15,5"}},
		{name: "val 3", want: 20.5, fields: fields{Value: "20.5"}},
		{name: "val 4", want: 0, fields: fields{Value: "25x5"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ci := CurrencyInfo{
				XMLName:  tt.fields.XMLName,
				ID:       tt.fields.ID,
				NumCode:  tt.fields.NumCode,
				CharCode: tt.fields.CharCode,
				Nominal:  tt.fields.Nominal,
				Name:     tt.fields.Name,
				Value:    tt.fields.Value,
			}
			if got := ci.GetValue(); got != tt.want {
				t.Errorf("CurrencyInfo.GetValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCurrencyList_Fetch(t *testing.T) {
	var datavalid string = `<?xml version="1.0" encoding="windows-1251"?>
	<ValCurs Date="15.07.2023" name="Foreign Currency Market">
	<Valute ID="R01239"><NumCode>978</NumCode><CharCode>EUR</CharCode><Nominal>1</Nominal><Name>Euro</Name><Value>101,1992</Value></Valute>
	<Valute ID="R01235"><NumCode>840</NumCode><CharCode>USD</CharCode><Nominal>1</Nominal><Name>US Dollar</Name><Value>90,1190</Value></Valute>
	</ValCurs>`
	var datainvalid string = "some invalid data"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/valid" {
			fmt.Fprintln(w, datavalid)
		} else if r.RequestURI == "/invalid" {
			fmt.Fprintln(w, datainvalid)
		} else {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintln(w, "Hello, client")
		}
	}))
	defer srv.Close()

	t.Run("valid data", func(t *testing.T) {
		var cl CurrencyList
		cl.url = srv.URL + "/valid"
		err := cl.Fetch()
		if err != nil {
			t.Errorf("Can't parse data: %v", err)
			return
		}
		if len(cl.Rates) != 3 {
			t.Errorf("Expected 3 currencies, but found %v", len(cl.Rates))
		}
	})

	t.Run("invalid data", func(t *testing.T) {
		var cl CurrencyList
		cl.url = srv.URL + "/invalid"
		err := cl.Fetch()
		if err == nil {
			t.Errorf("Bad data, but no error!")
			return
		}
	})
	t.Run("bad url", func(t *testing.T) {
		var cl CurrencyList
		cl.url = srv.URL + "/404"
		err := cl.Fetch()
		if err == nil {
			t.Errorf("Bad url, but no error!")
			return
		}
	})
}
